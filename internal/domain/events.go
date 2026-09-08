package domain

import (
	"time"
)

// Event é a interface marcadora para todos os eventos de domínio.
type Event interface {
	EventName() string
	OccurredAt() time.Time
}

// baseEvent fornece campos comuns para eventos de domínio.
type baseEvent struct {
	occurredAt time.Time
}

func newBaseEvent() baseEvent {
	return baseEvent{occurredAt: time.Now()}
}

func (b baseEvent) OccurredAt() time.Time {
	return b.occurredAt
}

// KnockStartedEvent é emitido quando uma sequência de batidas é iniciada.
type KnockStartedEvent struct {
	baseEvent
	TargetHost string
	TotalSteps int
	Delay      time.Duration
}

func NewKnockStartedEvent(targetHost string, totalSteps int, delay time.Duration) KnockStartedEvent {
	return KnockStartedEvent{
		baseEvent:  newBaseEvent(),
		TargetHost: targetHost,
		TotalSteps: totalSteps,
		Delay:      delay,
	}
}

func (e KnockStartedEvent) EventName() string {
	return "KnockStartedEvent"
}

// PortKnockStartedEvent é emitido imediatamente antes de disparar o pacote para a porta.
type PortKnockStartedEvent struct {
	baseEvent
	Target     KnockTarget
	ResolvedIP string
	Step       int
	TotalSteps int
}

func NewPortKnockStartedEvent(target KnockTarget, resolvedIP string, step, totalSteps int) PortKnockStartedEvent {
	return PortKnockStartedEvent{
		baseEvent:  newBaseEvent(),
		Target:     target,
		ResolvedIP: resolvedIP,
		Step:       step,
		TotalSteps: totalSteps,
	}
}

func (e PortKnockStartedEvent) EventName() string {
	return "PortKnockStartedEvent"
}

// PortKnockedEvent é emitido após cada tentativa de batida em uma porta específica.
type PortKnockedEvent struct {
	baseEvent
	Target     KnockTarget
	ResolvedIP string
	Step       int
	TotalSteps int
	Duration   time.Duration
}

func NewPortKnockedEvent(target KnockTarget, resolvedIP string, step, totalSteps int, duration time.Duration) PortKnockedEvent {
	return PortKnockedEvent{
		baseEvent:  newBaseEvent(),
		Target:     target,
		ResolvedIP: resolvedIP,
		Step:       step,
		TotalSteps: totalSteps,
		Duration:   duration,
	}
}

func (e PortKnockedEvent) EventName() string {
	return "PortKnockedEvent"
}

// KnockCompletedEvent é emitido quando toda a sequência é finalizada com sucesso.
type KnockCompletedEvent struct {
	baseEvent
	TargetHost    string
	TotalSteps    int
	TotalDuration time.Duration
}

func NewKnockCompletedEvent(targetHost string, totalSteps int, totalDuration time.Duration) KnockCompletedEvent {
	return KnockCompletedEvent{
		baseEvent:     newBaseEvent(),
		TargetHost:    targetHost,
		TotalSteps:    totalSteps,
		TotalDuration: totalDuration,
	}
}

func (e KnockCompletedEvent) EventName() string {
	return "KnockCompletedEvent"
}

// KnockFailedEvent é emitido caso ocorra uma falha durante a sequência de batidas.
type KnockFailedEvent struct {
	baseEvent
	TargetHost string
	Step       int
	Err        error
}

func NewKnockFailedEvent(targetHost string, step int, err error) KnockFailedEvent {
	return KnockFailedEvent{
		baseEvent:  newBaseEvent(),
		TargetHost: targetHost,
		Step:       step,
		Err:        err,
	}
}

func (e KnockFailedEvent) EventName() string {
	return "KnockFailedEvent"
}
