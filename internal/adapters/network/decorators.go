package network

import (
	"context"
	"time"

	"github.com/ph-py/knock-go/internal/domain"
	"github.com/ph-py/knock-go/internal/ports"
)

// MetricCallback é um callback para registrar métricas de telemetria ou logging.
type MetricCallback func(target domain.KnockTarget, duration time.Duration, err error)

// LoggingKnockerDecorator implementa o Decorator Pattern adicionando
// medição de performance e rastreamento de chamadas sem modificar a implementação subjacente.
type LoggingKnockerDecorator struct {
	inner    ports.Knocker
	onKnock  MetricCallback
}

func NewLoggingKnockerDecorator(inner ports.Knocker, onKnock MetricCallback) *LoggingKnockerDecorator {
	return &LoggingKnockerDecorator{
		inner:   inner,
		onKnock: onKnock,
	}
}

func (d *LoggingKnockerDecorator) Resolve(ctx context.Context, host string, ipVer domain.IPVersion) (string, error) {
	return d.inner.Resolve(ctx, host, ipVer)
}

func (d *LoggingKnockerDecorator) Hit(ctx context.Context, target domain.KnockTarget, ipVer domain.IPVersion) (*domain.HitResult, error) {
	start := time.Now()
	res, err := d.inner.Hit(ctx, target, ipVer)
	elapsed := time.Since(start)

	if d.onKnock != nil {
		d.onKnock(target, elapsed, err)
	}

	return res, err
}

var _ ports.Knocker = (*LoggingKnockerDecorator)(nil)
