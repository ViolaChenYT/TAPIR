package IR

import (
	//
	"fmt"
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

type Replica struct {
	replica_id int
	partition  int // partition number
	state      int // state of the replica (NORMAL or VIEWCHANGING)
	listener   net.Listener
	close      chan bool
	record     *record
}

type record struct {
	values map[operation]int
}

func emptyRecord() *record {
	return &record{
		values: make(map[operation]int),
	}
}

func NewReplica(id int) (*Replica, error) {
	replica := Replica{
		replica_id: id,
		state:      NORMAL,
		close:      make(chan bool),
		record:     emptyRecord(),
	}
	base := 8080 // port to listen on
	rpc.Register(replica)
	// rpc.HandleHTTP()
	ln, err := net.Listen("tcp", string(base+id))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	replica.listener = ln
	go rpc.Accept(ln)

	return &replica, nil
}

// the rpc function
func (r *Replica) handleOperation(request Message, reply *Message) error {
	return nil
}

// func (r *Replica) handleConnection(conn net.Conn) {
// 	defer conn.Close()
// 	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
// 	for {
// 		msg, err := rw.ReadString('\n')
// 		if err != nil {
// 			if err != io.EOF {
// 				fmt.Println(err)
// 			}
// 			return
// 		}
// 		fmt.Println(msg)
// 	}
// }

func (r *Replica) ExecInconsistent(operation) error {
	return nil
}

func (r *Replica) ExecConsensus(operation) (result, error) {
	return result{}, nil
}

func (r *Replica) Sync() error {
	return nil
}

// func (r *Replica) Merge(d, u) (record, error) {
// 	return record{}, nil
// }

func (r *Replica) Close() {
	r.listener.Close()
	r.close <- true
}
