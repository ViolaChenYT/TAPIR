package tapir

import (
	. "github.com/ViolaChenYT/TAPIR/common"
)

// TapirServer represents a replica server interfacing the TAPIR and IR protocol
type TapirServer interface {
	// Invoke inconsistent operation (commit, abort), no return value
	ExecInconsistent(op *Request) error

	// Invoke consensus operation (prepare)
	ExecConsensus(op *Request) (*Response, error)

	// Invoke unlogged operation (only support read)
	ExecUnlogged(op *Request) (*Response, error)
}
