package IR

import (
	"bufio" //
	"fmt"
	"io"
	"net"
	"net/rpc"
)

const (
	NORMAL       = iota
	VIEWCHANGING = iota
)

type Replica struct {
	replica_id    int
	partition     int // partition number
	state         int // state of the replica (NORMAL or VIEWCHANGING)
	otherReplicas []*rpc.Client
}

type record struct {
	values map[operation]bool
}

func NewReplica(id int) *Replica {
	replica := Replica{replica_id: id}
	return &replica
}

func (r *Replica) Start() error {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go r.handleConnection(conn)
	}
}

func (r *Replica) handleConnection(conn net.Conn) {
	defer conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	for {
		msg, err := rw.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
		fmt.Println(msg)
	}
}

func (r *Replica) ExecInconsistent(operation) error {
	return nil
}

func (r *Replica) ExecConsensus(operation) (result, error) {
	return result{}, nil
}

func (r *Replica) Sync() error {
	return nil
}

func (r *Replica) Merge(d, u) (record, error) {
	return record{}, nil
}
