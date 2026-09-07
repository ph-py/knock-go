package network

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ph-py/knock-go/internal/domain"
	"github.com/ph-py/knock-go/internal/ports"
)

// NetKnocker é o adaptador real de rede que implementa ports.Knocker.
// Desacoplado das estratégias específicas através da interface HitStrategy.
type NetKnocker struct {
	tcpStrategy HitStrategy
	udpStrategy HitStrategy
	resolver    *net.Resolver
}

func NewNetKnocker(tcpStrategy HitStrategy, udpStrategy HitStrategy) *NetKnocker {
	return &NetKnocker{
		tcpStrategy: tcpStrategy,
		udpStrategy: udpStrategy,
		resolver:    net.DefaultResolver,
	}
}

func (k *NetKnocker) resolveIP(ctx context.Context, host string, ipVer domain.IPVersion) (string, error) {
	// Se já for um IP literal
	if parsed := net.ParseIP(host); parsed != nil {
		if ipVer == domain.IPv4Only && parsed.To4() == nil {
			return "", fmt.Errorf("host %s is not an IPv4 address", host)
		}
		if ipVer == domain.IPv6Only && (parsed.To4() != nil || parsed.To16() == nil) {
			return "", fmt.Errorf("host %s is not an IPv6 address", host)
		}
		return parsed.String(), nil
	}

	network := "ip"
	switch ipVer {
	case domain.IPv4Only:
		network = "ip4"
	case domain.IPv6Only:
		network = "ip6"
	}

	ips, err := k.resolver.LookupIP(ctx, network, host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve host %s: %w", host, err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for host %s", host)
	}

	// Filtra de acordo com a versão de IP se for IPDefault ou específico
	for _, ip := range ips {
		if ipVer == domain.IPv4Only && ip.To4() != nil {
			return ip.String(), nil
		}
		if ipVer == domain.IPv6Only && ip.To4() == nil && ip.To16() != nil {
			return ip.String(), nil
		}
		if ipVer == domain.IPDefault {
			return ip.String(), nil
		}
	}

	return ips[0].String(), nil
}

func (k *NetKnocker) Hit(ctx context.Context, target domain.KnockTarget, ipVer domain.IPVersion) (*domain.HitResult, error) {
	resolvedIP, err := k.resolveIP(ctx, target.Host, ipVer)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	var strat HitStrategy
	switch target.Protocol {
	case domain.ProtocolUDP:
		strat = k.udpStrategy
	case domain.ProtocolTCP:
		strat = k.tcpStrategy
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", target.Protocol)
	}

	if err := strat.Send(ctx, resolvedIP, target.Port); err != nil {
		return nil, err
	}

	duration := time.Since(start)

	return &domain.HitResult{
		Target:     target,
		ResolvedIP: resolvedIP,
		Duration:   duration,
	}, nil
}

// Garantia em tempo de compilação de que NetKnocker implementa ports.Knocker
var _ ports.Knocker = (*NetKnocker)(nil)
