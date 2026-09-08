package domain

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrExecutionAlreadyStarted = errors.New("knock execution already started")
	ErrExecutionNotRunning     = errors.New("knock execution is not running")
	ErrInvalidStepIndex        = errors.New("invalid step index for execution")
)

type ExecutionState string

const (
	StatePending   ExecutionState = "PENDING"
	StateRunning   ExecutionState = "RUNNING"
	StateCompleted ExecutionState = "COMPLETED"
	StateFailed    ExecutionState = "FAILED"
	StateCancelled ExecutionState = "CANCELLED"
)

// KnockExecution é a raiz de agregação (Aggregate Root) que gerencia o ciclo
// de vida, os estados, os resultados e os eventos de uma sessão de port-knocking.
type KnockExecution struct {
	mu          sync.Mutex
	id          string
	sequence    KnockSequence
	state       ExecutionState
	currentStep int
	startTime   time.Time
	endTime     time.Time
	results     []HitResult
	events      []Event
}

func NewKnockExecution(id string, sequence KnockSequence) *KnockExecution {
	return &KnockExecution{
		id:          id,
		sequence:    sequence,
		state:       StatePending,
		currentStep: 0,
		results:     make([]HitResult, 0, len(sequence.Targets)),
		events:      make([]Event, 0),
	}
}

func (e *KnockExecution) ID() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.id
}

func (e *KnockExecution) State() ExecutionState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.state
}

func (e *KnockExecution) Sequence() KnockSequence {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.sequence
}

func (e *KnockExecution) CurrentStep() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.currentStep
}

func (e *KnockExecution) Results() []HitResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	res := make([]HitResult, len(e.results))
	copy(res, e.results)
	return res
}

// Start inicia a execução da sequência, garantindo o estado consistente e emitindo KnockStartedEvent.
func (e *KnockExecution) Start(targetHost string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StatePending {
		return ErrExecutionAlreadyStarted
	}
	e.state = StateRunning
	e.startTime = time.Now()
	e.currentStep = 0

	e.events = append(e.events, NewKnockStartedEvent(
		targetHost,
		len(e.sequence.Targets),
		e.sequence.Delay,
	))
	return nil
}

// StartHit registra o início do envio do pacote para uma porta e emite PortKnockStartedEvent.
func (e *KnockExecution) StartHit(target KnockTarget, resolvedIP string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning {
		return ErrExecutionNotRunning
	}

	e.events = append(e.events, NewPortKnockStartedEvent(
		target,
		resolvedIP,
		e.currentStep+1,
		len(e.sequence.Targets),
	))
	return nil
}

// RecordHit registra o sucesso de uma batida em porta e avança o passo da execução.
func (e *KnockExecution) RecordHit(result HitResult) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning {
		return ErrExecutionNotRunning
	}

	e.currentStep++
	e.results = append(e.results, result)

	e.events = append(e.events, NewPortKnockedEvent(
		result.Target,
		result.ResolvedIP,
		e.currentStep,
		len(e.sequence.Targets),
		result.Duration,
	))
	return nil
}

// Complete finaliza a execução com sucesso.
func (e *KnockExecution) Complete(targetHost string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning {
		return ErrExecutionNotRunning
	}

	e.state = StateCompleted
	e.endTime = time.Now()
	totalDuration := e.endTime.Sub(e.startTime)

	e.events = append(e.events, NewKnockCompletedEvent(
		targetHost,
		len(e.sequence.Targets),
		totalDuration,
	))
	return nil
}

// Fail encerra a execução em estado de falha e emite KnockFailedEvent.
func (e *KnockExecution) Fail(targetHost string, err error) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning {
		return fmt.Errorf("%w: cannot fail execution in state %s", ErrExecutionNotRunning, e.state)
	}

	e.state = StateFailed
	e.endTime = time.Now()

	e.events = append(e.events, NewKnockFailedEvent(
		targetHost,
		e.currentStep+1,
		err,
	))
	return nil
}

// Cancel cancela a execução em andamento.
func (e *KnockExecution) Cancel() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != StateRunning && e.state != StatePending {
		return fmt.Errorf("%w: cannot cancel in state %s", ErrExecutionNotRunning, e.state)
	}
	e.state = StateCancelled
	e.endTime = time.Now()
	return nil
}

// CollectEvents esvazia e retorna a fila de eventos de domínio acumulados no agregado.
func (e *KnockExecution) CollectEvents() []Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	events := e.events
	e.events = make([]Event, 0)
	return events
}
