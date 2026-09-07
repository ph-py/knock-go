package domain

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Erros de domínio
var (
	ErrInvalidPort     = errors.New("port must be between 1 and 65535")
	ErrInvalidProtocol = errors.New("protocol must be tcp or udp")
	ErrEmptySequence   = errors.New("knock sequence cannot be empty")
	ErrEmptyHost       = errors.New("host cannot be empty")
	ErrNegativeDelay   = errors.New("delay cannot be negative")
)

// Protocol é um Value Object que representa o protocolo de transporte.
type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
)

func ParseProtocol(p string) (Protocol, error) {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "tcp", "":
		return ProtocolTCP, nil
	case "udp":
		return ProtocolUDP, nil
	default:
		return "", fmt.Errorf("%w: '%s'", ErrInvalidProtocol, p)
	}
}

// IPVersion é um Value Object que expressa a preferência de família de endereços IP.
type IPVersion string

const (
	IPDefault IPVersion = "any"
	IPv4Only  IPVersion = "ipv4"
	IPv6Only  IPVersion = "ipv6"
)

// Port é um Value Object que encapsula e valida a porta de rede (1-65535).
type Port uint16

func NewPort(val int) (Port, error) {
	if val < 1 || val > 65535 {
		return 0, fmt.Errorf("%w: %d", ErrInvalidPort, val)
	}
	return Port(val), nil
}

func (p Port) Uint16() uint16 {
	return uint16(p)
}

func (p Port) String() string {
	return fmt.Sprintf("%d", p)
}

// KnockTarget é um Value Object imutável representando um alvo individual de batida.
type KnockTarget struct {
	Host     string
	Port     Port
	Protocol Protocol
}

func NewKnockTarget(host string, port Port, proto Protocol) (KnockTarget, error) {
	trimmedHost := strings.TrimSpace(host)
	if trimmedHost == "" {
		return KnockTarget{}, ErrEmptyHost
	}
	if proto != ProtocolTCP && proto != ProtocolUDP {
		return KnockTarget{}, ErrInvalidProtocol
	}
	return KnockTarget{
		Host:     trimmedHost,
		Port:     port,
		Protocol: proto,
	}, nil
}

func (t KnockTarget) Address() string {
	return net.JoinHostPort(t.Host, t.Port.String())
}

// KnockSequence é um Value Object que agrega a lista ordenada de alvos e o intervalo entre eles.
type KnockSequence struct {
	Targets []KnockTarget
	Delay   time.Duration
}

func NewKnockSequence(targets []KnockTarget, delay time.Duration) (KnockSequence, error) {
	if len(targets) == 0 {
		return KnockSequence{}, ErrEmptySequence
	}
	if delay < 0 {
		return KnockSequence{}, ErrNegativeDelay
	}
	// Cópia defensiva para garantir imutabilidade
	copied := make([]KnockTarget, len(targets))
	copy(copied, targets)

	return KnockSequence{
		Targets: copied,
		Delay:   delay,
	}, nil
}

// HitResult é um Value Object que contém as informações do resultado de uma batida em porta.
type HitResult struct {
	Target     KnockTarget
	ResolvedIP string
	Duration   time.Duration
}
