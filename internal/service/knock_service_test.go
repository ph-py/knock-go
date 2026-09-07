package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/knock-win/knock/internal/adapters/cli"
	"github.com/knock-win/knock/internal/adapters/network"
	"github.com/knock-win/knock/internal/bootstrap"
	"github.com/knock-win/knock/internal/domain"
	"github.com/knock-win/knock/internal/service"
)

func TestExecuteKnockSequenceSuccessInHighGear(t *testing.T) {
	fakeKnocker := network.NewFakeKnocker()
	fakeSleeper := network.NewFakeSleeper()
	bufferPrinter := cli.NewBufferPrinter()

	app := bootstrap.Bootstrap(bootstrap.Config{
		Knocker: fakeKnocker,
		Sleeper: fakeSleeper,
		Printer: bufferPrinter,
		Verbose: true,
	})

	port1, _ := domain.NewPort(7000)
	port2, _ := domain.NewPort(8000)
	port3, _ := domain.NewPort(9000)

	t1, _ := domain.NewKnockTarget("myserver.com", port1, domain.ProtocolTCP)
	t2, _ := domain.NewKnockTarget("myserver.com", port2, domain.ProtocolUDP)
	t3, _ := domain.NewKnockTarget("myserver.com", port3, domain.ProtocolTCP)

	delay := 500 * time.Millisecond
	seq, err := domain.NewKnockSequence([]domain.KnockTarget{t1, t2, t3}, delay)
	if err != nil {
		t.Fatalf("unexpected sequence error: %v", err)
	}

	cmd := service.ExecuteKnockSequenceCommand{
		ExecutionID: "test-run-1",
		Host:        "myserver.com",
		Sequence:    seq,
		IPVersion:   domain.IPDefault,
		Verbose:     true,
	}

	ctx := context.Background()
	err = app.Execute(ctx, cmd)
	if err != nil {
		t.Fatalf("expected execution to succeed, got error: %v", err)
	}

	// Verifica se todas as 3 portas foram batidas na ordem
	if len(fakeKnocker.RecordedHits) != 3 {
		t.Fatalf("expected 3 hits, got %d", len(fakeKnocker.RecordedHits))
	}
	if fakeKnocker.RecordedHits[0].Port.Uint16() != 7000 || fakeKnocker.RecordedHits[0].Protocol != domain.ProtocolTCP {
		t.Errorf("expected hit 0 to be 7000/tcp, got %+v", fakeKnocker.RecordedHits[0])
	}
	if fakeKnocker.RecordedHits[1].Port.Uint16() != 8000 || fakeKnocker.RecordedHits[1].Protocol != domain.ProtocolUDP {
		t.Errorf("expected hit 1 to be 8000/udp, got %+v", fakeKnocker.RecordedHits[1])
	}
	if fakeKnocker.RecordedHits[2].Port.Uint16() != 9000 || fakeKnocker.RecordedHits[2].Protocol != domain.ProtocolTCP {
		t.Errorf("expected hit 2 to be 9000/tcp, got %+v", fakeKnocker.RecordedHits[2])
	}

	// Verifica se foram feitos 2 sleeps (entre a 1ª e 2ª batida, e entre a 2ª e 3ª batida)
	if len(fakeSleeper.RecordedSleeps) != 2 {
		t.Fatalf("expected 2 sleeps, got %d", len(fakeSleeper.RecordedSleeps))
	}
	for _, d := range fakeSleeper.RecordedSleeps {
		if d != delay {
			t.Errorf("expected sleep duration %v, got %v", delay, d)
		}
	}

	// Verifica saída verbose (Observer pattern formatado igual ao knock.c)
	output := bufferPrinter.String()
	expectedLines := []string{
		"hitting tcp 192.168.1.100:7000",
		"hitting udp 192.168.1.100:8000",
		"hitting tcp 192.168.1.100:9000",
	}
	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but got:\n%s", expected, output)
		}
	}
}

func TestExecuteKnockSequenceFailureInHighGear(t *testing.T) {
	fakeKnocker := network.NewFakeKnocker()
	expectedErr := errors.New("network unreachable")
	fakeKnocker.ErrToReturn = expectedErr

	fakeSleeper := network.NewFakeSleeper()
	bufferPrinter := cli.NewBufferPrinter()

	app := bootstrap.Bootstrap(bootstrap.Config{
		Knocker: fakeKnocker,
		Sleeper: fakeSleeper,
		Printer: bufferPrinter,
		Verbose: true,
	})

	port, _ := domain.NewPort(7000)
	target, _ := domain.NewKnockTarget("myserver.com", port, domain.ProtocolTCP)
	seq, _ := domain.NewKnockSequence([]domain.KnockTarget{target}, 10*time.Millisecond)

	cmd := service.ExecuteKnockSequenceCommand{
		ExecutionID: "test-run-fail",
		Host:        "myserver.com",
		Sequence:    seq,
		IPVersion:   domain.IPDefault,
		Verbose:     true,
	}

	err := app.Execute(context.Background(), cmd)
	if err == nil {
		t.Fatalf("expected error from failed knock, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped error containing %v, got %v", expectedErr, err)
	}
}

func TestExecuteKnockSequenceContextCancellation(t *testing.T) {
	fakeKnocker := network.NewFakeKnocker()
	fakeSleeper := network.NewFakeSleeper()

	app := bootstrap.Bootstrap(bootstrap.Config{
		Knocker: fakeKnocker,
		Sleeper: fakeSleeper,
	})

	port, _ := domain.NewPort(7000)
	target, _ := domain.NewKnockTarget("myserver.com", port, domain.ProtocolTCP)
	seq, _ := domain.NewKnockSequence([]domain.KnockTarget{target, target}, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := service.ExecuteKnockSequenceCommand{
		ExecutionID: "test-run-cancel",
		Host:        "myserver.com",
		Sequence:    seq,
		IPVersion:   domain.IPDefault,
	}

	err := app.Execute(ctx, cmd)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}
