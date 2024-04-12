package IR

import (
	"bufio" //
	"fmt"
	"io"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

type Client struct {
	client_id     int
	operation_cnt int
	conn          net.Conn
	close         chan bool // close channel
	rpcclient     *rpc.Client
}

type operation struct {
	op_type   string
	key       string
	value     string
	timestamp time.Time
}

type result struct {
	op_type string
	key     string
	value   string
}

func NewClient(id int, serverHost string, serverPort int) (*Client, error) {
	cli, err := rpc.DialHTTP("tcp", net.JoinHostPort(serverHost, strconv.Itoa(serverPort)))
	if err != nil {
		return nil, err
	}
	client := Client{
		client_id:     id,
		operation_cnt: 0,
		close:         make(chan bool),
		rpcclient:     cli,
	}
	conn, err := net.Dial("tcp", "localhost:8080")
	client.conn = conn
	if err != nil {
		fmt.Println(err)
	}
	go client.handleConnection()
	return &client, nil
}

func (c *Client) handleConnection() {
	rw := bufio.NewReadWriter(bufio.NewReader(c.conn), bufio.NewWriter(c.conn))
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

func (c *Client) decide([]result) result {
	return result{}
}

func (c *Client) InvokeInconsistent(operation) error {
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

func (c *Client) InvokeConsensus(operation, []result) (result, error) {
	// use c.decide to pick the result
	/// ignore below
	defer c.conn.Close()
	rw := bufio.NewReadWriter(bufio.NewReader(c.conn), bufio.NewWriter(c.conn))
	for {
		_, err := rw.WriteString(fmt.Sprintf("Consensus! %d\n", c.operation_cnt))
		if err != nil {
			fmt.Println(err)
			return result{}, err
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println(err)
			return result{}, err
		}
		c.operation_cnt++
	}
}
