package IR

import (
	. "github.com/ViolaChenYT/TAPIR/common"
)

// Implement IR Replica functions
type Replica interface {
	// Invoke inconsistent operation (commit, abort), no return value
	ExecInconsistent(op *Request) error

	// Invoke consensus operation (prepare)
	ExecConsensus(op *Request) (*Response, error)

	// Invoke unlogged operation (only support read)
	ExecUnlogged(op *Request) (*Response, error)
}
