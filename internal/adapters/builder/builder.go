package builder

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/knock-win/knock/internal/domain"
)

// KnockSequenceBuilder implementa o Builder Pattern para construir
// e validar uma sequência de batidas a partir de opções e argumentos CLI.
type KnockSequenceBuilder struct {
	host            string
	defaultProtocol domain.Protocol
	delay           time.Duration
	targets         []domain.KnockTarget
	errs            []error
}

func NewKnockSequenceBuilder() *KnockSequenceBuilder {
	return &KnockSequenceBuilder{
		defaultProtocol: domain.ProtocolTCP,
		delay:           0,
		targets:         make([]domain.KnockTarget, 0),
		errs:            make([]error, 0),
	}
}

func (b *KnockSequenceBuilder) SetHost(host string) *KnockSequenceBuilder {
	b.host = strings.TrimSpace(host)
	return b
}

func (b *KnockSequenceBuilder) SetDefaultProtocol(proto domain.Protocol) *KnockSequenceBuilder {
	b.defaultProtocol = proto
	return b
}

func (b *KnockSequenceBuilder) SetDelay(delay time.Duration) *KnockSequenceBuilder {
	if delay < 0 {
		b.errs = append(b.errs, domain.ErrNegativeDelay)
	} else {
		b.delay = delay
	}
	return b
}

// AddTargetFromString faz o parsing de argumentos no formato "porta" ou "porta:proto"
// Exemplos aceitos: "22", "80:tcp", "53:udp", "123:UDP"
func (b *KnockSequenceBuilder) AddTargetFromString(spec string) *KnockSequenceBuilder {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		b.errs = append(b.errs, fmt.Errorf("empty port specification"))
		return b
	}

	parts := strings.Split(spec, ":")
	portStr := parts[0]
	proto := b.defaultProtocol

	if len(parts) > 2 {
		b.errs = append(b.errs, fmt.Errorf("invalid port format %q (expected port[:proto])", spec))
		return b
	}

	if len(parts) == 2 {
		parsedProto, err := domain.ParseProtocol(parts[1])
		if err != nil {
			b.errs = append(b.errs, fmt.Errorf("invalid protocol in %q: %w", spec, err))
			return b
		}
		proto = parsedProto
	}

	portNum, err := strconv.Atoi(portStr)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("invalid port number %q in %q", portStr, spec))
		return b
	}

	domainPort, err := domain.NewPort(portNum)
	if err != nil {
		b.errs = append(b.errs, err)
		return b
	}

	target, err := domain.NewKnockTarget(b.host, domainPort, proto)
	if err != nil {
		b.errs = append(b.errs, err)
		return b
	}

	b.targets = append(b.targets, target)
	return b
}

// AddTargetsFromStrings adiciona uma lista de especificações "port[:proto]".
func (b *KnockSequenceBuilder) AddTargetsFromStrings(specs []string) *KnockSequenceBuilder {
	for _, s := range specs {
		b.AddTargetFromString(s)
	}
	return b
}

// Build valida e consolida a sequência de batidas.
func (b *KnockSequenceBuilder) Build() (*domain.KnockSequence, error) {
	if len(b.errs) > 0 {
		// Retorna o primeiro erro acumulado
		return nil, b.errs[0]
	}
	if b.host == "" {
		return nil, domain.ErrEmptyHost
	}
	if len(b.targets) == 0 {
		return nil, domain.ErrEmptySequence
	}

	seq, err := domain.NewKnockSequence(b.targets, b.delay)
	if err != nil {
		return nil, err
	}
	return &seq, nil
}
