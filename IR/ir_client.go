package IR

import (
	"bufio" //
	"fmt"
	"net"
	"net/rpc"
	"time"
)

type Client struct {
	ServerAddresses  []string
	client_id        int
	operation_cnt    int
	conn             net.Conn
	close            chan bool // close channel
	allReplicas      []*rpc.Client
	replicaAddresses []string
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

func NewClient(id int, serverAddresses []string) (*Client, error) {
	client := Client{
		client_id:        id,
		operation_cnt:    0,
		close:            make(chan bool),
		replicaAddresses: serverAddresses,
		allReplicas:      make([]*rpc.Client, len(serverAddresses)),
	}
	for i, addr := range serverAddresses {
		cli, err := rpc.DialHTTP("tcp", addr)
		checkError(err)
		client.allReplicas[i] = cli
	}
	// newConnectMsg, _ := json.marshal(NewConnect(id))
	// client.rpcclient.Call("Replica.Connect", newConnectMsg, nil)
	// go client.handleConnection()
	return &client, nil
}

// func (c *Client) handleConnection() {
// 	rw := bufio.NewReadWriter(bufio.NewReader(c.conn), bufio.NewWriter(c.conn))
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

func (c *Client) operationProcess(op operation) {
	// send to server
	for _, cli := range c.allReplicas {
		request := Message{
			Type:        MsgPropose,
			OperationID: c.operation_cnt,
			Op:          &op,
		}
		reply := Message{}
		cli.Call("Replica.handleOperation", request, &reply)
	}
	// wait for reply

	// send back to client
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

func (c *Client) Close() {
	c.close <- true
}
