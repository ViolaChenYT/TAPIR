package IR

import (
	//

	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	. "github.com/ViolaChenYT/TAPIR/common"
)

const ( // number of replicas that can be down at any one point
	timeout = 1
) // total 2f + 1 = 5 replicas?
type ConsensusDecide func(results []*Response) *Response

type Client struct {
	client_id        int
	operation_cnt    int
	conn             net.Conn
	close            chan bool // close channel
	allReplicas      []*rpc.Client
	replicaAddresses []string
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
		log.Println(addr, " connected")
		CheckError(err)
		client.allReplicas[i] = cli
	}
	return &client, nil
}

func (c *Client) callOneReplica(cli *rpc.Client, msg Message, wg *sync.WaitGroup, results map[int]Response) *Message {
	defer wg.Done()
	reply := Message{}
	cli.Call("Replica.HandleOperation", msg, &reply)
	results[c.client_id] = *reply.Response
	return &reply
}

func (c *Client) msgOneReplica(cli *rpc.Client, msg Message) {
	reply := Message{}
	cli.Call("Replica.HandleOperation", msg, &reply)
}

func (c *Client) InvokeInconsistent(req *Request) error {
	results := make(map[int]Response)
	var wg sync.WaitGroup
	f := (len(c.allReplicas) - 1) / 2
	wg.Add(f + 1)
	for _, cli := range c.allReplicas {
		msg := NewPropose(req.TxnID, req)
		go c.callOneReplica(cli, msg, &wg, results)
	}
	// wait for reply
	wg.Wait()
	for _, cli := range c.allReplicas {
		msg := NewFinalize(req.TxnID)
		go c.msgOneReplica(cli, msg)
	}
	c.operation_cnt++
	return nil
}

func (c *Client) InvokeConsensus(req *Request, decide ConsensusDecide) (*Response, error) {
	results := make(map[int]Response)
	consensusRes := Response{}
	var wg sync.WaitGroup
	f := (len(c.allReplicas) - 1) / 2
	wg.Add(f + 1)
	timer := time.NewTimer(timeout * time.Second)
	for _, cli := range c.allReplicas {
		msg := NewPropose(req.TxnID, req)
		go c.callOneReplica(cli, msg, &wg, results)
	}

	<-timer.C
	value_cnt := make(map[string]int)
	max_val, max_cnt := "", 0
	for _, res := range results {
		res_val := res.Value
		if _, ok := value_cnt[res_val]; ok {
			value_cnt[res_val]++
		} else {
			value_cnt[res_val] = 1
		}
		if value_cnt[res_val] > max_cnt {
			max_val, max_cnt = res_val, value_cnt[res_val]
		}
	}
	if (max_cnt) >= (3*f/2)+1 {
		for _, cli := range c.allReplicas {
			consensusRes.Value = max_val
			msg := Finalize(req.TxnID, &consensusRes)
			go c.msgOneReplica(cli, msg)
		}
	} else {
		wg.Wait()
		var result_arr []*Response
		for _, res := range results {
			result_arr = append(result_arr, &res)
		}
		consensusRes = *decide(result_arr)
		finalize_msg := Finalize(req.TxnID, &consensusRes)
		var wg2 sync.WaitGroup
		wg2.Add(f + 1)
		for _, cli := range c.allReplicas {
			go c.callOneReplica(cli, finalize_msg, &wg2, results)
		}
		wg2.Wait() // wait for "Confirm" from replicas
	}
	c.operation_cnt++
	return &consensusRes, nil
}

func (c *Client) Close() {
	c.close <- true
}
