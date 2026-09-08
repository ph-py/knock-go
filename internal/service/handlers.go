package service

import (
	"context"
	"fmt"

	"github.com/ph-py/knock-go/internal/domain"
	"github.com/ph-py/knock-go/internal/ports"
)

// KnockService centraliza os casos de uso do serviço de port knocking.
type KnockService struct {
	knocker ports.Knocker
	sleeper ports.Sleeper
	bus     *MessageBus
}

func NewKnockService(knocker ports.Knocker, sleeper ports.Sleeper, bus *MessageBus) *KnockService {
	return &KnockService{
		knocker: knocker,
		sleeper: sleeper,
		bus:     bus,
	}
}

// ExecuteKnockSequenceHandler processa o comando ExecuteKnockSequenceCommand.
func (s *KnockService) ExecuteKnockSequence(ctx context.Context, cmd ExecuteKnockSequenceCommand) error {
	exec := domain.NewKnockExecution(cmd.ExecutionID, cmd.Sequence)

	if err := exec.Start(cmd.Host); err != nil {
		return err
	}
	if err := s.bus.PublishAll(ctx, exec.CollectEvents()); err != nil {
		return err
	}

	resolvedIP, err := s.knocker.Resolve(ctx, cmd.Host, cmd.IPVersion)
	if err != nil {
		_ = exec.Fail(cmd.Host, err)
		_ = s.bus.PublishAll(ctx, exec.CollectEvents())
		return fmt.Errorf("failed to resolve host %s: %w", cmd.Host, err)
	}

	targets := cmd.Sequence.Targets
	for i, target := range targets {
		select {
		case <-ctx.Done():
			_ = exec.Cancel()
			_ = s.bus.PublishAll(ctx, exec.CollectEvents())
			return ctx.Err()
		default:
		}

		if err := exec.StartHit(target, resolvedIP); err != nil {
			return err
		}
		if err := s.bus.PublishAll(ctx, exec.CollectEvents()); err != nil {
			return err
		}

		hitResult, err := s.knocker.Hit(ctx, target, cmd.IPVersion)
		if err != nil {
			_ = exec.Fail(cmd.Host, err)
			_ = s.bus.PublishAll(ctx, exec.CollectEvents())
			return fmt.Errorf("failed to knock %s (%s): %w", target.Address(), target.Protocol, err)
		}

		if err := exec.RecordHit(*hitResult); err != nil {
			return err
		}
		if err := s.bus.PublishAll(ctx, exec.CollectEvents()); err != nil {
			return err
		}

		// Atraso entre batidas (aplica-se entre batidas, ou seja, se não for a última)
		if i < len(targets)-1 && cmd.Sequence.Delay > 0 {
			if err := s.sleeper.Sleep(ctx, cmd.Sequence.Delay); err != nil {
				_ = exec.Fail(cmd.Host, err)
				_ = s.bus.PublishAll(ctx, exec.CollectEvents())
				return err
			}
		}
	}

	if err := exec.Complete(cmd.Host); err != nil {
		return err
	}
	return s.bus.PublishAll(ctx, exec.CollectEvents())
}

const (
	ansiGreenBold = "\033[1;32m"
	ansiRedBold   = "\033[1;31m"
	ansiReset     = "\033[0m"
)

// VerboseOutputEventHandler reage a eventos imprimindo na porta de saída (Printer).
type VerboseOutputEventHandler struct {
	printer     ports.Printer
	linePending bool
}

func NewVerboseOutputEventHandler(printer ports.Printer) *VerboseOutputEventHandler {
	return &VerboseOutputEventHandler{printer: printer}
}

func (h *VerboseOutputEventHandler) Handle(ctx context.Context, evt domain.Event) error {
	switch e := evt.(type) {
	case domain.PortKnockStartedEvent:
		h.printer.Print("hitting %s %s:%s ... ", e.Target.Protocol, e.ResolvedIP, e.Target.Port.String())
		h.linePending = true
	case domain.PortKnockedEvent:
		h.printer.Print("%sOK%s\n", ansiGreenBold, ansiReset)
		h.linePending = false
	case domain.KnockFailedEvent:
		if h.linePending {
			h.printer.Print("%sFAIL%s\n", ansiRedBold, ansiReset)
			h.linePending = false
		}
	}
	return nil
}
