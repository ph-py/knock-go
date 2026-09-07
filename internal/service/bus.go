package service

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/knock-win/knock/internal/domain"
)

// CommandHandlerFunc é o tipo da função que processa um comando.
type CommandHandlerFunc func(ctx context.Context, cmd interface{}) error

// EventHandlerFunc é o tipo da função que lida com um evento de domínio.
type EventHandlerFunc func(ctx context.Context, evt domain.Event) error

// MessageBus é o barramento de mensagens responsável pelo despacho desacoplado
// de comandos (1 destinatário) e publicação de eventos (múltiplos ouvintes/Observer pattern).
type MessageBus struct {
	mu              sync.RWMutex
	commandHandlers map[reflect.Type]CommandHandlerFunc
	eventHandlers   map[reflect.Type][]EventHandlerFunc
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		commandHandlers: make(map[reflect.Type]CommandHandlerFunc),
		eventHandlers:   make(map[reflect.Type][]EventHandlerFunc),
	}
}

// RegisterCommandHandler registra um manipulador único para um tipo específico de comando.
func (b *MessageBus) RegisterCommandHandler(cmdSample interface{}, handler CommandHandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := reflect.TypeOf(cmdSample)
	b.commandHandlers[t] = handler
}

// RegisterEventHandler adiciona um ouvinte para um tipo específico de evento de domínio (Observer Pattern).
func (b *MessageBus) RegisterEventHandler(evtSample domain.Event, handler EventHandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := reflect.TypeOf(evtSample)
	b.eventHandlers[t] = append(b.eventHandlers[t], handler)
}

// Dispatch executa o comando através de seu handler registrado.
func (b *MessageBus) Dispatch(ctx context.Context, cmd interface{}) error {
	b.mu.RLock()
	t := reflect.TypeOf(cmd)
	handler, exists := b.commandHandlers[t]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no command handler registered for %v", t)
	}

	return handler(ctx, cmd)
}

// Publish distribui um evento para todos os event handlers inscritos.
func (b *MessageBus) Publish(ctx context.Context, evt domain.Event) error {
	b.mu.RLock()
	t := reflect.TypeOf(evt)
	handlers := b.eventHandlers[t]
	b.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

// PublishAll publica uma lista de eventos de domínio em ordem.
func (b *MessageBus) PublishAll(ctx context.Context, events []domain.Event) error {
	for _, evt := range events {
		if err := b.Publish(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}
