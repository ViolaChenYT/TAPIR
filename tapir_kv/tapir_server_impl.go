package tapir

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"

	IR "github.com/ViolaChenYT/TAPIR/IR"
	. "github.com/ViolaChenYT/TAPIR/common"
)

// Server represents a Tapir server
type TapirServerImpl struct {
	store TapirReplica
	id    int
}

// NewServer creates a new instance of Server
func NewTapirServer(id int, serverAddr string) IR.Replica {
	server := TapirServerImpl{
		store: NewReplica(id),
		id:    id,
	}
	go server.Listen(serverAddr)
	return &server
}

func (r *TapirServerImpl) Listen(serverAddr string) {
	rpc.Register(r)
	ln, err := net.Listen("tcp", "localhost:"+serverAddr)
	CheckError(err)
	log.Println("Replica", serverAddr, "listening")
	go rpc.Accept(ln)
}

func (r *TapirServerImpl) HandleOperation(request Message, reply *Message) error {
	log.Println("Handling Operation")
	// write operation id and op to its record as tentative and responds to client with <reply,id>
	if request.Type == MsgPropose {
		// write id and op to its record as tentative
		r.ExecInconsistent(request.Request)
		return nil
	} else if request.Type == MsgFinalize {
		// write id and op to its record as finalized
		r.ExecConsensus(request.Request)
		return nil
	} else { // MsgReply
		return fmt.Errorf("replica shouldn't get message reply or confirm")
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
