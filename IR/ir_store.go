package IR

import (
	//
	"fmt"
	"log"
	"net"
	"net/rpc"
)

const ( // state of the replica
	NORMAL       = iota
	VIEWCHANGING = iota
)

const ( // state of operations
	TENTATIVE = iota
	FINALIZED = iota
)

const ( // state of reply)
	REPLY_OK   = iota
	REPLY_FAIL = iota
)

type Replica struct {
	replica_id int
	partition  int // partition number
	state      int // state of the replica (NORMAL or VIEWCHANGING)
	listener   net.Listener
	close      chan bool
	record     *Record
}

type Record struct {
	values map[Operation]int
}

func emptyRecord() *Record {
	return &Record{
		values: make(map[Operation]int),
	}
}

func NewReplica(id int) (*Replica, error) {
	replica := Replica{
		replica_id: id,
		state:      NORMAL,
		close:      make(chan bool),
		record:     emptyRecord(),
	}
	go replica.Listen(id)
	return &replica, nil
}

func (r *Replica) Listen(base int) {
	rpc.Register(r)
	ln, err := net.Listen("tcp", "localhost:"+fmt.Sprint(base))
	checkError(err)
	log.Println("Replica localhost:"+fmt.Sprint(base), "listening")
	r.listener = ln
	go rpc.Accept(ln)
}

// the rpc function
func (r *Replica) HandleOperation(request Message, reply *Message) error {
	fmt.Println("Replica ", r.replica_id, " received operation ", request.OperationID)
	// write operation id and op to its record as tentative and responds to client with <reply,id>
	if request.Type == MsgPropose {
		return nil
	} else if request.Type == MsgFinalize {
		return nil
	} else if request.Type == MsgConfirm {
		return nil
	} else { // MsgReply
	}
	return nil
}

func (r *Replica) Close() error {
	r.listener.Close()
	r.close <- true
	return nil
}
