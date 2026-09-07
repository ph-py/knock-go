package network

import (
	"context"
	"sync"
	"time"

	"github.com/knock-win/knock/internal/domain"
	"github.com/knock-win/knock/internal/ports"
)

// RealSleeper implementa a porta ports.Sleeper usando o relógio real do sistema.
type RealSleeper struct{}

func NewRealSleeper() *RealSleeper {
	return &RealSleeper{}
}

func (s *RealSleeper) Sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

var _ ports.Sleeper = (*RealSleeper)(nil)

// FakeKnocker é um dublê de teste (Fake/Spy) que simula o comportamento
// de envio de pacotes em memória para testes unitários determinísticos e rápidos.
type FakeKnocker struct {
	mu           sync.Mutex
	RecordedHits []domain.KnockTarget
	FixedIP      string
	ErrToReturn  error
}

func NewFakeKnocker() *FakeKnocker {
	return &FakeKnocker{
		RecordedHits: make([]domain.KnockTarget, 0),
		FixedIP:      "192.168.1.100",
	}
}

func (f *FakeKnocker) Hit(ctx context.Context, target domain.KnockTarget, ipVer domain.IPVersion) (*domain.HitResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.ErrToReturn != nil {
		return nil, f.ErrToReturn
	}

	f.RecordedHits = append(f.RecordedHits, target)
	return &domain.HitResult{
		Target:     target,
		ResolvedIP: f.FixedIP,
		Duration:   time.Millisecond,
	}, nil
}

var _ ports.Knocker = (*FakeKnocker)(nil)

// FakeSleeper é um dublê de teste que registra as durações de pausa sem aguardar o tempo real.
type FakeSleeper struct {
	mu             sync.Mutex
	RecordedSleeps []time.Duration
	ErrToReturn    error
}

func NewFakeSleeper() *FakeSleeper {
	return &FakeSleeper{
		RecordedSleeps: make([]time.Duration, 0),
	}
}

func (f *FakeSleeper) Sleep(ctx context.Context, d time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.ErrToReturn != nil {
		return f.ErrToReturn
	}

	f.RecordedSleeps = append(f.RecordedSleeps, d)
	return nil
}

var _ ports.Sleeper = (*FakeSleeper)(nil)
