package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/knock-win/knock/internal/ports"
)

// ConsolePrinter implementa a porta ports.Printer enviando saída para Stdout e Stderr.
type ConsolePrinter struct {
	stdout io.Writer
	stderr io.Writer
}

func NewConsolePrinter() *ConsolePrinter {
	return &ConsolePrinter{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (p *ConsolePrinter) Print(format string, args ...interface{}) {
	fmt.Fprintf(p.stdout, format, args...)
}

func (p *ConsolePrinter) Error(format string, args ...interface{}) {
	fmt.Fprintf(p.stderr, format, args...)
}

var _ ports.Printer = (*ConsolePrinter)(nil)

// BufferPrinter é um dublê de teste que acumula a saída em memória para inspeção.
type BufferPrinter struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func NewBufferPrinter() *BufferPrinter {
	return &BufferPrinter{}
}

func (p *BufferPrinter) Print(format string, args ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(&p.buf, format, args...)
}

func (p *BufferPrinter) Error(format string, args ...interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintf(&p.buf, format, args...)
}

func (p *BufferPrinter) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buf.String()
}

var _ ports.Printer = (*BufferPrinter)(nil)
