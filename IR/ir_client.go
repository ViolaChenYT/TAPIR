package IR

import (
	//

	"fmt"
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
	close            chan bool               // close channel
	allReplicas      map[int]*rpc.Client     // <replica_id, client>
	replicaAddresses map[int]*ReplicaAddress // <replica_id, address>
	f                int                     // max number of fault tolerance
}

func NewIRClient(config *Configuration) (*Client, error) {
	client := Client{
		client_id:        config.Client.IR_ID,
		operation_cnt:    0,
		close:            make(chan bool),
		replicaAddresses: config.Replicas,
		allReplicas:      make(map[int]*rpc.Client),
		f:                config.F,
	}
	// errCh := make(chan error, len(config.Replicas))

	for idx, addr := range config.Replicas {
		cli, err := rpc.Dial("tcp", addr.SpecificString())
		CheckError(err)
		client.allReplicas[idx] = cli
	}
	return &client, nil
}

func (c *Client) callOneReplica(rep int, cli *rpc.Client, msg Message, results map[int]*Response, semaphone chan bool) *Message {
	// port := c.replicaAddresses[rep].Port
	// real_cli, err := rpc.Dial("tcp", "localhost:"+port)
	// if msg.Request.Op == OP_PREPARE {
	// 	log.Println("calling 1 replica for prepare", msg.Request.Prepare.Txn)
	// }
	reply := Message{}
	err := cli.Call(fmt.Sprintf("IRReplica%d.HandleOperation", rep), &msg, &reply)
	if err != nil {
		log.Fatal("arith error: ", err)
	}
	if results != nil {
		results[rep] = reply.Response
		// log.Println("callOneReplica, client", c.client_id, "Op", msg.Request.Op, "txn", msg.Request.TxnID, "val", reply.Response.Value)
		semaphone <- true
	}
	return &reply
}

func (c *Client) msgOneReplica(rep int, cli *rpc.Client, msg Message) {
	reply := Message{}
	// if msg.Request.Op == OP_PREPARE {
	// 	log.Println("messaging 1 replica for prepare", msg.Request.TxnID)
	// }
	cli.Call(fmt.Sprintf("IRReplica%d.HandleOperation", rep), msg, &reply)
}

func (c *Client) InvokeInconsistent(req *Request) error {
	log.Println("InvokeInconsistent", req.Op.ToString(), req.TxnID)
	results := make(map[int]*Response)

	semaphone := make(chan bool, c.f+1)
	for id, cli := range c.allReplicas {
		msg := NewPropose(req.TxnID, req, INCONSISTENT)
		msg.Type = MsgPropose
		go c.callOneReplica(id, cli, msg, results, semaphone)
	}
	log.Println("invoke I propose done")
	done := make(chan bool)
	// wait for reply
	go func() {
		for i := 0; i < c.f+1; i++ {
			<-semaphone
		}
		done <- true
	}()
	<-done
	log.Println("Invoke I, finalizing")
	var wg sync.WaitGroup
	for idx, cli := range c.allReplicas {
		msg := NewFinalize(req.TxnID, INCONSISTENT)
		msg.Request = req
		// go c.msgOneReplica(idx, cli, msg)
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.msgOneReplica(idx, cli, msg)
		}()
	}
	wg.Wait()
	c.operation_cnt++
	return nil
}

func (c *Client) InvokeConsensus(req *Request, decide ConsensusDecide) (*Response, error) {
	log.Println("InvokeConsensus", req.Op, req.Prepare.Txn)
	results := make(map[int]*Response)
	consensusRes := Response{}
	sem := make(chan bool, len(c.allReplicas))
	timer := time.NewTimer(timeout * time.Second)
	for id, cli := range c.allReplicas {
		msg := NewPropose(req.TxnID, req, CONSENSUS)
		go c.callOneReplica(id, cli, msg, results, sem)
	}
	log.Println("waiting for consensus timer")
	<-timer.C
	value_cnt := make(map[string]int)
	max_val, max_cnt := "", 0
	for key, res := range results {
		res_val := res.Value
		log.Println("------------------", key, ": res_val", res_val)
		if _, ok := value_cnt[res_val]; ok {
			value_cnt[res_val]++
		} else {
			value_cnt[res_val] = 1
		}
		if value_cnt[res_val] > max_cnt {
			max_val = res_val
			max_cnt = value_cnt[res_val]
		}
	}
	log.Println("max_cnt", max_cnt, "max_val", max_val)
	if (max_cnt) >= (3*c.f/2)+1 {
		for idx, cli := range c.allReplicas {
			consensusRes.Value = max_val
			log.Println("fast path finalize")
			msg := Finalize(req.TxnID, &consensusRes)
			msg.Request = req
			log.Println("finalizing concensus")
			c.msgOneReplica(idx, cli, msg)
		}
		log.Println("fast path done")
	} else {
		log.Println("wait for slow path")
		done := make(chan bool)
		go func() {
			for i := 0; i < c.f+1; i++ {
				<-sem
			}
			done <- true
		}()
		log.Println("waiting for consensus ")
		<-done
		var result_arr []*Response
		for _, res := range results {
			result_arr = append(result_arr, res)
		}
		consensusRes = *decide(result_arr)
		finalize_msg := Finalize(req.TxnID, &consensusRes)
		finalize_msg.Request = req
		finalize_msg.ProtoType = CONSENSUS
		for idx, cli := range c.allReplicas {
			go c.msgOneReplica(idx, cli, finalize_msg)
		}
	}
	c.operation_cnt++
	return &consensusRes, nil
}

func (c *Client) InvokeUnlogged(replicaIdx int, req *Request) (*Response, error) {
	reqMsg := NewUnlogged(req)

	replyMsg := c.callOneReplica(replicaIdx, c.allReplicas[replicaIdx], reqMsg, nil, nil)
	return replyMsg.Response, nil
}

func (c *Client) Close() {
	c.close <- true
}
