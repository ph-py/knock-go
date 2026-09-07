package builder_test

import (
	"testing"
	"time"

	"github.com/ph-py/knock-go/internal/adapters/builder"
	"github.com/ph-py/knock-go/internal/domain"
)

func TestKnockSequenceBuilder(t *testing.T) {
	b := builder.NewKnockSequenceBuilder().
		SetHost("example.com").
		SetDelay(100 * time.Millisecond).
		SetDefaultProtocol(domain.ProtocolTCP).
		AddTargetFromString("7000").
		AddTargetFromString("8000:udp").
		AddTargetFromString("9000:tcp")

	seq, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(seq.Targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(seq.Targets))
	}

	if seq.Targets[0].Protocol != domain.ProtocolTCP || seq.Targets[0].Port.Uint16() != 7000 {
		t.Errorf("expected target 0 to be 7000/tcp, got %+v", seq.Targets[0])
	}
	if seq.Targets[1].Protocol != domain.ProtocolUDP || seq.Targets[1].Port.Uint16() != 8000 {
		t.Errorf("expected target 1 to be 8000/udp, got %+v", seq.Targets[1])
	}
	if seq.Targets[2].Protocol != domain.ProtocolTCP || seq.Targets[2].Port.Uint16() != 9000 {
		t.Errorf("expected target 2 to be 9000/tcp, got %+v", seq.Targets[2])
	}
	if seq.Delay != 100*time.Millisecond {
		t.Errorf("expected delay 100ms, got %v", seq.Delay)
	}
}

func TestKnockSequenceBuilderErrors(t *testing.T) {
	// Missing host
	_, err := builder.NewKnockSequenceBuilder().
		AddTargetFromString("7000").
		Build()
	if err != domain.ErrEmptyHost {
		t.Errorf("expected ErrEmptyHost, got %v", err)
	}

	// Invalid port number
	_, err = builder.NewKnockSequenceBuilder().
		SetHost("example.com").
		AddTargetFromString("abc").
		Build()
	if err == nil {
		t.Errorf("expected error for invalid port, got nil")
	}

	// Port out of range
	_, err = builder.NewKnockSequenceBuilder().
		SetHost("example.com").
		AddTargetFromString("70000").
		Build()
	if err == nil {
		t.Errorf("expected error for out of range port, got nil")
	}

	// Invalid protocol
	_, err = builder.NewKnockSequenceBuilder().
		SetHost("example.com").
		AddTargetFromString("80:icmp").
		Build()
	if err == nil {
		t.Errorf("expected error for invalid protocol, got nil")
	}
}
