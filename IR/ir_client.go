package IR

import (
	"bufio" //
	"fmt"
	"net"
	"time"
)

type Client struct {
	client_id     int
	operation_cnt int
	conn          net.Conn
}

type operation struct {
	op_type   string
	key       string
	value     string
	timestamp time.Time
}

func NewClient(id int) *Client {
	client := Client{client_id: id, operation_cnt: 0}
	return &client
}

func (c *Client) Start() error {
	conn, err := net.Dial("tcp", "localhost:8080")
	c.conn = conn
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer c.conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	for {
		_, err := rw.WriteString(fmt.Sprintf("Hello, Server! %d\n", c.operation_cnt))
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println(err)
			return err
		}
		c.operation_cnt++
	}
}

func (c *Client) InvokeInconsistent() error {
	defer c.conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(c.conn), bufio.NewWriter(c.conn))
	for {
		_, err := rw.WriteString(fmt.Sprintf("Inconsistent! %d\n", c.operation_cnt))
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println(err)
			return err
		}
		c.operation_cnt++
	}
}

func (c *Client) InvokeConsensus() {
	defer c.conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(c.conn), bufio.NewWriter(c.conn))
	for {
		_, err := rw.WriteString(fmt.Sprintf("Consensus! %d\n", c.operation_cnt))
		if err != nil {
			fmt.Println(err)
			return
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println(err)
			return
		}
		c.operation_cnt++
	}
}
