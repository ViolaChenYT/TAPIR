package tapir

import (
	"errors"
	"fmt"
	"log"

	. "github.com/ViolaChenYT/TAPIR/IR"
	. "github.com/ViolaChenYT/TAPIR/common"
)

// Server represents a Tapir server
type TapirServer struct {
	store TapirReplica
	id    int
}

// NewServer creates a new instance of Server
func NewTapirServer(id int) IRAppReplica {
	return &TapirServer{
		store: NewReplica(id),
		id:    id,
	}
}

func (server *TapirServer) ExecInconsistentUpcall(op *Request) error {
	switch op.Op {
	case OP_COMMIT:
		log.Println("asking for commit")
		server.store.Commit(op.TxnID, op.Commit.Timestamp)
	case OP_ABORT:
		server.store.Abort(op.TxnID)
	default:
		return errors.New("Unrecognized inconsistent operation")
	}
	return nil
}

func (server *TapirServer) ExecConsensusUpcall(op *Request) (*Response, error) {
	if op.Op == OP_PREPARE {
		reply, err := server.store.Prepare(op.Prepare.Txn, op.Prepare.Timestamp)
		return reply, err
	}

	return nil, errors.New("Unrecognized consensus operation")
}

func (server *TapirServer) ExecUnloggedUpcall(op *Request) (*Response, error) {
	if op.Op == OP_GET {
		val, timestamp, err := server.store.Read(op.Get.Key)
		return NewReadResponse(val, timestamp), err
	}
	return nil, errors.New("Unrecognized unlogged operation")
}

func (server *TapirServer) String() string {
	return fmt.Sprintf("TAPIR Server(id: %d)", server.id)
}
