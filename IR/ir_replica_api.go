package IR

import (
	. "github.com/ViolaChenYT/TAPIR/common"
)

// IR Replica functions
type IRReplica interface {
	// Handle requests
	HandleOperation(request *Message, reply *Message) error
	// Stop the server
	Stop()
}

// IR Replica App functions
type IRAppReplica interface {
	// Invoke inconsistent operation (commit, abort), no return value
	ExecInconsistentUpcall(op *Request) error

	// Invoke consensus operation (prepare)
	ExecConsensusUpcall(op *Request) (*Response, error)

	// Invoke unlogged operation (only support read)
	ExecUnloggedUpcall(op *Request) (*Response, error)
}
