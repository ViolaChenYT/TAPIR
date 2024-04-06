package IR

import (
	"bufio" //
	"fmt"
	"io"
	"net"
)

type Replica struct {
	replica_id int
	partition  int // partition number
}

func NewReplica(id int) *Replica {
	replica := Replica{replica_id: id}
	return &replica
}

func (r *Replica) Start() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
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
