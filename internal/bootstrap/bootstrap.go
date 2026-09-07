package bootstrap

import (
	"context"
	"time"

	"github.com/knock-win/knock/internal/adapters/cli"
	"github.com/knock-win/knock/internal/adapters/network"
	"github.com/knock-win/knock/internal/domain"
	"github.com/knock-win/knock/internal/ports"
	"github.com/knock-win/knock/internal/service"
)

// Application é o ponto de contato do Composition Root (Bootstrap),
// contendo os serviços e o MessageBus configurados com injeção de dependências.
type Application struct {
	Bus     *service.MessageBus
	Printer ports.Printer
}

// Config contém as opções de configuração do container de dependências.
type Config struct {
	Knocker ports.Knocker
	Sleeper ports.Sleeper
	Printer ports.Printer
	Verbose bool
}

// Bootstrap inicializa a aplicação, instanciando os adaptadores e registrando
// os handlers no MessageBus (Composition Root / Manual DI).
func Bootstrap(cfg Config) *Application {
	if cfg.Knocker == nil {
		tcpStrat := network.NewTCPHitStrategy(200 * time.Millisecond)
		udpStrat := network.NewUDPHitStrategy()
		cfg.Knocker = network.NewNetKnocker(tcpStrat, udpStrat)
	}

	if cfg.Sleeper == nil {
		cfg.Sleeper = network.NewRealSleeper()
	}

	if cfg.Printer == nil {
		cfg.Printer = cli.NewConsolePrinter()
	}

	bus := service.NewMessageBus()
	svc := service.NewKnockService(cfg.Knocker, cfg.Sleeper, bus)

	// Registra o handler para o comando de execução
	bus.RegisterCommandHandler(service.ExecuteKnockSequenceCommand{}, func(ctx context.Context, cmd interface{}) error {
		c, ok := cmd.(service.ExecuteKnockSequenceCommand)
		if !ok {
			return nil
		}
		return svc.ExecuteKnockSequence(ctx, c)
	})

	// Se o modo verbose estiver ativo, registra o ouvinte de eventos (Observer Pattern)
	if cfg.Verbose {
		verboseHandler := service.NewVerboseOutputEventHandler(cfg.Printer)
		bus.RegisterEventHandler(domain.PortKnockedEvent{}, verboseHandler.Handle)
	}

	return &Application{
		Bus:     bus,
		Printer: cfg.Printer,
	}
}

// Execute executa uma sequência de batidas despachando o comando pelo MessageBus.
func (a *Application) Execute(ctx context.Context, cmd service.ExecuteKnockSequenceCommand) error {
	return a.Bus.Dispatch(ctx, cmd)
}
