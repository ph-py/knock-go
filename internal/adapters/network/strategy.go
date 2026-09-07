package network

import (
	"context"
	"net"
	"time"

	"github.com/ph-py/knock-go/internal/domain"
)

// HitStrategy é a interface do Strategy Pattern para transmissão de pacotes em diferentes protocolos.
type HitStrategy interface {
	Send(ctx context.Context, targetIP string, port domain.Port) error
}

// TCPHitStrategy implementa a estratégia de batida TCP (envio de pacote SYN).
type TCPHitStrategy struct {
	dialTimeout time.Duration
}

func NewTCPHitStrategy(dialTimeout time.Duration) *TCPHitStrategy {
	if dialTimeout <= 0 {
		dialTimeout = 200 * time.Millisecond
	}
	return &TCPHitStrategy{dialTimeout: dialTimeout}
}

func (s *TCPHitStrategy) Send(ctx context.Context, targetIP string, port domain.Port) error {
	addr := net.JoinHostPort(targetIP, port.String())
	dialer := net.Dialer{
		Timeout: s.dialTimeout,
	}

	// Inicia a tentativa de conexão enviando o pacote SYN na rede.
	// Em port-knocking, as portas geralmente estão fechadas ou filtradas pelo firewall.
	// Assim que o pacote SYN sai pela placa de rede, o objetivo foi atingido (como no knock.c).
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if conn != nil {
		_ = conn.Close()
	}

	// Se houver erro de rede como conexão recusada ou timeout, é esperado para portas fechadas/filtradas.
	// Apenas cancelamento de contexto deve ser retornado como erro.
	if ctx.Err() != nil {
		return ctx.Err()
	}
	_ = err // Pacote SYN transmitido com sucesso
	return nil
}

// UDPHitStrategy implementa a estratégia de batida UDP (envio de datagrama de 1 byte).
type UDPHitStrategy struct{}

func NewUDPHitStrategy() *UDPHitStrategy {
	return &UDPHitStrategy{}
}

func (s *UDPHitStrategy) Send(ctx context.Context, targetIP string, port domain.Port) error {
	addr := net.JoinHostPort(targetIP, port.String())
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Envia datagrama de 1 byte "" idêntico a sendto(sd, "", 1, 0, ...) do knock.c
	payload := []byte{0}
	_, err = conn.Write(payload)
	return err
}
