package domain_test

import (
	"testing"
	"time"

	"github.com/knock-win/knock/internal/domain"
)

func TestParseProtocol(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.Protocol
		wantErr  bool
	}{
		{"tcp", domain.ProtocolTCP, false},
		{"TCP", domain.ProtocolTCP, false},
		{"", domain.ProtocolTCP, false},
		{"udp", domain.ProtocolUDP, false},
		{"UDP", domain.ProtocolUDP, false},
		{"http", "", true},
		{"xyz", "", true},
	}

	for _, tt := range tests {
		got, err := domain.ParseProtocol(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseProtocol(%q) err = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if got != tt.expected {
			t.Errorf("ParseProtocol(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestNewPort(t *testing.T) {
	tests := []struct {
		port    int
		wantErr bool
	}{
		{0, true},
		{-1, true},
		{1, false},
		{80, false},
		{65535, false},
		{65536, true},
	}

	for _, tt := range tests {
		p, err := domain.NewPort(tt.port)
		if (err != nil) != tt.wantErr {
			t.Errorf("NewPort(%d) err = %v, wantErr %v", tt.port, err, tt.wantErr)
		}
		if !tt.wantErr && p.Uint16() != uint16(tt.port) {
			t.Errorf("NewPort(%d) = %d, want %d", tt.port, p.Uint16(), tt.port)
		}
	}
}

func TestNewKnockTarget(t *testing.T) {
	port, _ := domain.NewPort(8080)

	// Valid target
	target, err := domain.NewKnockTarget("localhost", port, domain.ProtocolTCP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Address() != "localhost:8080" {
		t.Errorf("expected localhost:8080, got %s", target.Address())
	}

	// Empty host
	_, err = domain.NewKnockTarget("   ", port, domain.ProtocolTCP)
	if err != domain.ErrEmptyHost {
		t.Errorf("expected ErrEmptyHost, got %v", err)
	}

	// Invalid protocol
	_, err = domain.NewKnockTarget("localhost", port, "invalid")
	if err != domain.ErrInvalidProtocol {
		t.Errorf("expected ErrInvalidProtocol, got %v", err)
	}
}

func TestNewKnockSequence(t *testing.T) {
	port80, _ := domain.NewPort(80)
	target, _ := domain.NewKnockTarget("example.com", port80, domain.ProtocolTCP)

	// Empty targets
	_, err := domain.NewKnockSequence([]domain.KnockTarget{}, 100*time.Millisecond)
	if err != domain.ErrEmptySequence {
		t.Errorf("expected ErrEmptySequence, got %v", err)
	}

	// Negative delay
	_, err = domain.NewKnockSequence([]domain.KnockTarget{target}, -1*time.Millisecond)
	if err != domain.ErrNegativeDelay {
		t.Errorf("expected ErrNegativeDelay, got %v", err)
	}

	// Valid
	seq, err := domain.NewKnockSequence([]domain.KnockTarget{target}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seq.Targets) != 1 || seq.Delay != 50*time.Millisecond {
		t.Errorf("unexpected sequence content: %+v", seq)
	}
}
