package IR

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	. "github.com/ViolaChenYT/TAPIR/common"
)

// Server represents a Tapir server
type IRReplicaImpl struct {
	app      IRAppReplica
	id       int
	listener net.Listener
	record   *Record
	addr     *ReplicaAddress
	mu       *sync.Mutex
}

const ( // state of operations
	TENTATIVE = iota
	FINALIZED = iota
)

const ( // state of reply)
	REPLY_OK   = iota
	REPLY_FAIL = iota
)

type Record struct {
	values map[Request]int
}

func emptyRecord() *Record {
	return &Record{
		values: make(map[Request]int),
	}
}

// NewServer creates a new instance of Server
func NewIRReplica(id int, serverAddr *ReplicaAddress, app IRAppReplica) IRReplica {
	server := IRReplicaImpl{
		id:     id,
		app:    app,
		record: emptyRecord(),
		addr:   serverAddr,
		mu:     &sync.Mutex{},
	}
	server.Listen(serverAddr)
	return &server
}

func dummyIRReplica() IRReplica {
	return &IRReplicaImpl{}
}

func (r *IRReplicaImpl) Listen(serverAddr *ReplicaAddress) {
	log.Println("client", r.id, "Listening on", r.addr.SpecificString())
	rpc.RegisterName(fmt.Sprintf("IRReplica%d", r.id), r)
	ln, err := net.Listen("tcp", r.addr.SpecificString())
	CheckError(err)
	log.Println("Replica", r.id, r.addr.Port, "listening")
	r.listener = ln
	go rpc.Accept(ln)
}

func (r *IRReplicaImpl) HandleOperation(request *Message, reply *Message) error {
	log.Println(r.id, "Handling Operation", request.Request.Op.ToString(), request.Type.ToString())
	if request.Request.Commit != nil {
		log.Println("TS", request.Request.Commit.Timestamp)
	}

	// if request.ProtoType == CONSENSUS {
	// 	log.Println(request.Request.Prepare.Txn)
	// }

	// write operation id and op to its record as tentative and responds to client with <reply,id>
	if request.Type == MsgPropose {
		// r.app.ExecInconsistentUpcall(request.Request)
		r.mu.Lock()
		r.record.values[*request.Request] = TENTATIVE
		r.mu.Unlock()
		reply.Response = NewResponse(RPLY_OK)
		log.Println("received propose")
		return nil
	} else if request.Type == MsgFinalize {
		log.Println("received finalize", request.Request.Op.ToString())
		if request.Request.Op == OP_PREPARE {
			log.Println("received prepare txn", request.Request.Prepare.Txn)
			r.app.ExecConsensusUpcall(request.Request)
			r.mu.Lock()
			r.record.values[*request.Request] = FINALIZED
			r.mu.Unlock()
			reply.Response = NewResponse(RPLY_OK)
			reply.Response.Value = "ok"
			return nil
		}
		if request.Request.Op == OP_GET {
			log.Println("received get")
			val, err := r.app.ExecUnloggedUpcall(request.Request)
			if err != nil {
				log.Println("ExecUnloggedUpcall error: ", err)
			}
			reply.Response = val
			return nil
		}
		if request.Request.Op == OP_ABORT {
			log.Println("received abort")
			r.app.ExecInconsistentUpcall(request.Request)
			reply.Response = NewResponse(RPLY_ABORT)
			return nil
		}
		if request.Request.Op != OP_COMMIT {
			return fmt.Errorf("replica shouldn't get message reply or confirm")
		}
		// should be commit
		log.Println("received commit")
		// write id and op to its record as finalized
		proto := request.ProtoType
		if proto == CONSENSUS {
			log.Println("request.Request: consensus ", request.Request, request.Request.Commit.Timestamp)
			response, err := r.app.ExecConsensusUpcall(request.Request)
			if err != nil {
				log.Println("ExeConsensus error: ", err)
			}
			reply.Response = response
		} else if proto == INCONSISTENT {
		log.Println("request.Request: inconsistent", request.Request.Op, request.Request.TxnID, request.Request.Commit.Timestamp)
		err := r.app.ExecInconsistentUpcall(request.Request)
		if err != nil {
			log.Println("ExeInconsistent error: ", err)
		}
		reply.Response = NewResponse(RPLY_OK)
		} else {
			return fmt.Errorf("replica shouldn't get message reply or confirm")
		}
		return nil
	} else { // MsgReply
		log.Println("shouldn't happen")
		return fmt.Errorf("replica shouldn't get message reply or confirm")
	}
}

// Stops the server gracefully
func (r *IRReplicaImpl) Stop() {
	if r.listener != nil {
		if err := r.listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}
	log.Println("Server stopped")
}

func (server *IRReplicaImpl) String() string {
	return fmt.Sprintf("IR Replica(id: %d)", server.id)
}
