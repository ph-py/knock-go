package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/knock-win/knock/internal/domain"
)

func TestKnockExecutionLifecycle(t *testing.T) {
	port1, _ := domain.NewPort(1000)
	port2, _ := domain.NewPort(2000)
	t1, _ := domain.NewKnockTarget("localhost", port1, domain.ProtocolTCP)
	t2, _ := domain.NewKnockTarget("localhost", port2, domain.ProtocolUDP)

	seq, err := domain.NewKnockSequence([]domain.KnockTarget{t1, t2}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to create sequence: %v", err)
	}

	exec := domain.NewKnockExecution("exec-1", seq)

	if exec.State() != domain.StatePending {
		t.Fatalf("expected state %s, got %s", domain.StatePending, exec.State())
	}

	// RecordHit before start should fail
	err = exec.RecordHit(domain.HitResult{Target: t1, ResolvedIP: "127.0.0.1", Duration: time.Millisecond})
	if err != domain.ErrExecutionNotRunning {
		t.Errorf("expected ErrExecutionNotRunning, got %v", err)
	}

	// Start execution
	if err := exec.Start("localhost"); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if exec.State() != domain.StateRunning {
		t.Errorf("expected state %s, got %s", domain.StateRunning, exec.State())
	}

	// Events should have KnockStartedEvent
	events := exec.CollectEvents()
	if len(events) != 1 || events[0].EventName() != "KnockStartedEvent" {
		t.Fatalf("expected 1 KnockStartedEvent, got %v", events)
	}
	// After collect, events should be empty
	if len(exec.CollectEvents()) != 0 {
		t.Errorf("expected events to be drained")
	}

	// Record hit 1
	err = exec.RecordHit(domain.HitResult{Target: t1, ResolvedIP: "127.0.0.1", Duration: 2 * time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected record hit error: %v", err)
	}
	if exec.CurrentStep() != 1 {
		t.Errorf("expected current step 1, got %d", exec.CurrentStep())
	}

	// Record hit 2
	err = exec.RecordHit(domain.HitResult{Target: t2, ResolvedIP: "127.0.0.1", Duration: 3 * time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected record hit error: %v", err)
	}
	if exec.CurrentStep() != 2 {
		t.Errorf("expected current step 2, got %d", exec.CurrentStep())
	}

	// Check PortKnockedEvent events
	events = exec.CollectEvents()
	if len(events) != 2 {
		t.Fatalf("expected 2 PortKnockedEvent events, got %d", len(events))
	}
	if events[0].EventName() != "PortKnockedEvent" || events[1].EventName() != "PortKnockedEvent" {
		t.Errorf("unexpected event names: %v, %v", events[0].EventName(), events[1].EventName())
	}

	// Complete execution
	err = exec.Complete("localhost")
	if err != nil {
		t.Fatalf("unexpected complete error: %v", err)
	}
	if exec.State() != domain.StateCompleted {
		t.Errorf("expected state %s, got %s", domain.StateCompleted, exec.State())
	}

	events = exec.CollectEvents()
	if len(events) != 1 || events[0].EventName() != "KnockCompletedEvent" {
		t.Fatalf("expected 1 KnockCompletedEvent, got %v", events)
	}
}

func TestKnockExecutionFailure(t *testing.T) {
	port, _ := domain.NewPort(1000)
	t1, _ := domain.NewKnockTarget("localhost", port, domain.ProtocolTCP)
	seq, _ := domain.NewKnockSequence([]domain.KnockTarget{t1}, 10*time.Millisecond)

	exec := domain.NewKnockExecution("exec-fail", seq)
	_ = exec.Start("localhost")

	fakeErr := errors.New("network unreachable")
	if err := exec.Fail("localhost", fakeErr); err != nil {
		t.Fatalf("unexpected fail error: %v", err)
	}
	if exec.State() != domain.StateFailed {
		t.Errorf("expected state %s, got %s", domain.StateFailed, exec.State())
	}

	events := exec.CollectEvents()
	var foundFailEvent bool
	for _, e := range events {
		if failedEvent, ok := e.(domain.KnockFailedEvent); ok {
			foundFailEvent = true
			if failedEvent.Err != fakeErr {
				t.Errorf("expected err %v, got %v", fakeErr, failedEvent.Err)
			}
		}
	}
	if !foundFailEvent {
		t.Errorf("expected KnockFailedEvent in collected events")
	}
}
