package service

import (
	"github.com/ph-py/knock-go/internal/domain"
)

// ExecuteKnockSequenceCommand é o comando de caso de uso para orquestrar
// a execução de uma sequência completa de batidas de portas.
type ExecuteKnockSequenceCommand struct {
	ExecutionID string
	Host        string
	Sequence    domain.KnockSequence
	IPVersion   domain.IPVersion
	Verbose     bool
}
