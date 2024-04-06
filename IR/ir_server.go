package IR

import (
	"bufio" //
	"fmt"
	"io"
	"net"
)

type Server struct {
	server_id int
}

func NewServer(id int) (*Server, error) {
	server := Server{server_id: id}
	return &server, nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
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
