package tapir

import (
	"errors"

	. "github.com/ViolaChenYT/TAPIR/common"
)

// Server represents a Tapir server
type TapirServerImpl struct {
	store TapirReplica
}

// NewServer creates a new instance of Server
func NewTapirServer() TapirServer {
	return &TapirServerImpl{
		store: NewReplica(),
	}
}

// ExecInconsistent implements the ExecInconsistent method of the TapirServer interface
func (server *TapirServerImpl) ExecInconsistent(op *Request) error {
	switch op.Op {
	case OP_COMMIT:
		server.store.Commit(op.TxnID, op.Commit.Timestamp)
	case OP_ABORT:
		server.store.Abort(op.TxnID)
	default:
		return errors.New("Unrecognized inconsistent operation")
	}
	return nil
}

// ExecConsensus implements the ExecConsensus method of the TapirServer interface
func (server *TapirServerImpl) ExecConsensus(op *Request) (*Response, error) {
	if op.Op == OP_PREPARE {
		reply, err := server.store.Prepare(op.Prepare.Txn, op.Prepare.Timestamp)
		return reply, err
	}

	return nil, errors.New("Unrecognized consensus operation")
}

func (server *TapirServerImpl) ExecUnlogged(op *Request) (*Response, error) {
	if op.Op == OP_GET {
		val, timestamp, err := server.store.Read(op.Get.Key)
		return NewReadResponse(val, timestamp), err
	}
	return nil, errors.New("Unrecognized unlogged operation")
}
