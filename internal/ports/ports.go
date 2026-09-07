package ports

import (
	"context"
	"time"

	"github.com/ph-py/knock-go/internal/domain"
)

// Knocker é a porta (interface) para o envio de batidas de rede aos alvos.
// Permite que a camada de serviço desacople completamente de sockets e SO.
type Knocker interface {
	Hit(ctx context.Context, target domain.KnockTarget, ipVer domain.IPVersion) (*domain.HitResult, error)
}

// Sleeper é a porta que abstrai a passagem de tempo e atrasos entre batidas.
// Permite testes unitários instantâneos sem atrasos de I/O de relógio.
type Sleeper interface {
	Sleep(ctx context.Context, d time.Duration) error
}

// Printer é a porta que abstrai a escrita de saídas para o usuário (stdout/stderr).
type Printer interface {
	Print(format string, args ...interface{})
	Error(format string, args ...interface{})
}
