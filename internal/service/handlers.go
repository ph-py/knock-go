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

	targets := cmd.Sequence.Targets
	for i, target := range targets {
		select {
		case <-ctx.Done():
			_ = exec.Cancel()
			_ = s.bus.PublishAll(ctx, exec.CollectEvents())
			return ctx.Err()
		default:
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

// VerboseOutputEventHandler reage a eventos imprimindo na porta de saída (Printer).
type VerboseOutputEventHandler struct {
	printer ports.Printer
}

func NewVerboseOutputEventHandler(printer ports.Printer) *VerboseOutputEventHandler {
	return &VerboseOutputEventHandler{printer: printer}
}

func (h *VerboseOutputEventHandler) Handle(ctx context.Context, evt domain.Event) error {
	switch e := evt.(type) {
	case domain.PortKnockedEvent:
		// Formato idêntico ao original knock.c:
		// vprint("hitting %s %s:%s\n", proto, ipname, port);
		h.printer.Print("hitting %s %s:%s\n", e.Target.Protocol, e.ResolvedIP, e.Target.Port.String())
	}
	return nil
}
