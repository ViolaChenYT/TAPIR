package IR

import (
	//
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

const ( // number of replicas that can be down at any one point
	f = 2 // seems the math tally up nicer
) // total 2f + 1 = 5 replicas?

type Client struct {
	client_id        int
	operation_cnt    int
	conn             net.Conn
	close            chan bool // close channel
	allReplicas      []*rpc.Client
	replicaAddresses []string
}

type Operation struct {
	op_type   string
	key       string
	value     string
	timestamp time.Time
}

type Result struct {
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
		specific_str := "localhost:" + addr
		log.Println("client trying Connecting to", specific_str)
		cli, err := rpc.Dial("tcp", specific_str)
		fmt.Println(addr, " connected")
		checkError(err)
		client.allReplicas[i] = cli
	}
	// newConnectMsg, _ := json.marshal(NewConnect(id))
	// client.rpcclient.Call("Replica.Connect", newConnectMsg, nil)
	// go client.handleConnection()
	return &client, nil
}

func (c *Client) callOneReplica(cli *rpc.Client, op Operation, wg *sync.WaitGroup, results chan Message) {
	defer wg.Done()
	request := Message{
		Type:        MsgPropose,
		OperationID: c.operation_cnt,
		Op:          &op,
	}
	reply := Message{}
	cli.Call("Replica.HandleOperation", request, &reply)
	results <- reply
}

func (c *Client) operationProcess(op Operation) {
	// send to server
	results := make(chan Message, len(c.allReplicas))
	var wg sync.WaitGroup
	for _, cli := range c.allReplicas {
		wg.Add(1)
		go c.callOneReplica(cli, op, &wg, results)
	}
	// wait for reply
	wg.Wait()
	close(results)
	// send back to client
	all_results := make([]*Result, len(c.allReplicas))
	i := 0
	for res := range results {
		// do something
		if res.Type == MsgReply {
			// do something
			op_result := res.Result
			all_results[i] = op_result
		} else {
			fmt.Println("Error: ", res)
		}
		i++
	}
	c.operation_cnt++
	// what about slow path?
}

func (c *Client) Close() {
	c.close <- true
}
