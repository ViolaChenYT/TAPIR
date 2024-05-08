package tapir_kv

import (
	"fmt"
	"log"
	"sync"

	"github.com/pingcap/go-ycsb/tapir/IR"
	. "github.com/pingcap/go-ycsb/tapir/common"
)

// TapirClientImpl is an implementation of the TapirClient interface
type TapirClientImpl struct {
	// Unique ID for this client
	client_id int

	// Ongoing transaction ID
	t_id int

	// Buffered transaction
	txn *Transaction

	// IR protocol client
	ir_client *IR.Client

	// Closet replica for read ops
	replica_id int

	// Size of majority replicas
	quorum_size int

	lock sync.Mutex
}

func NewTapirClient(config *Configuration) (TapirClient, error) {
	client := TapirClientImpl{
		client_id:   config.Client.TAPIR_ID,
		t_id:        0,
		replica_id:  config.Client.ClosestReplicaID,
		quorum_size: config.QuorumSize(),
	}

	// Create replica proxy
	cl, err := IR.NewIRClient(config)
	if err != nil {
		log.Panicf("Error creating IR client: %v", err)
		return nil, err
	}
	client.ir_client = cl
	// Run the transport in a new thread
	go client.run_client()

	return &client, nil
}

// Runs the transport event loop.
func (c *TapirClientImpl) run_client() {
	// TODO
}

func (c *TapirClientImpl) Begin() {
	// TODO: Implement a lock if previous transaction has not completed
	c.lock.Lock()
	c.t_id++

	// Create a transaction
	c.txn = NewTransaction(c.t_id)
}

func (c *TapirClientImpl) Read(key string) (string, error) {
	// If key is in the transaction's write set, the client returns value from the write set
	if val, ok := c.txn.WriteSet[key]; ok {
		return val, nil
	}
	// If the transaction has already read key, it returns a cached copy
	readSet, timeset := c.txn.ReadSet, c.txn.ReadTime
	if val, ok := readSet[key]; ok {
		return val, nil
	}
	timestamp := timeset[key]

	// Otherwise, the client sends Read(key) to the replica
	read_request := &Request{
		Op:    OP_GET,
		TxnID: c.t_id,
		Get:   &GetMessage{Key: key, Timestamp: NewTimestamp(c.client_id)},
	}
	response, err := c.ir_client.InvokeUnlogged(c.replica_id, read_request)
	CheckError(err)

	// On response, client puts (key, version) into the transaction's read set, and returns object to the application
	val, timestamp := response.Value, response.Timestamp // Placeholders

	c.txn.AddReadSet(key, val, timestamp)
	return val, nil
}

func (c *TapirClientImpl) Write(key string, value string) error {
	// Client buffers key and value in the write set until commit and returns immediately
	c.txn.AddWriteSet(key, value)

	// TODO: return some response
	return nil
}

func (c *TapirClientImpl) Commit() bool {
	// Client selects a proposed timestamp (local_time, client_id)
	// timestamp := NewTimestamp(c.client_id)
	// Client invokes Prepare(tx, timestamp) as an IR consensus operation.

	prepare_request := &Request{
		Op:      OP_PREPARE,
		TxnID:   c.t_id,
		Prepare: &PrepareMessage{Txn: c.txn, Timestamp: NewTimestamp(c.client_id)},
	}
	response, err := c.ir_client.InvokeConsensus(prepare_request, c.decide) // pass decide function
	if err != nil {
		log.Panicf("Error invoking consensus: %v", err)
		return false
	}
	log.Println("prepare passed, status: " + ReplyTypeString(response.Status))

	if response.Status == RPLY_OK {
		commit_request := &Request{
			Op:     OP_COMMIT,
			TxnID:  c.t_id,
			Commit: &CommitMessage{Timestamp: NewTimestamp(c.client_id)},
		}
		// Commit to all replicas
		log.Println("started commit request")
		c.ir_client.InvokeInconsistent(commit_request) // TODO: how to evoke Commit() on replicas?
		c.Continue()
		return true
	}

	// TODO: handle retry

	// Otherwise, abort
	c.Abort()
	return false
}

func (c *TapirClientImpl) Abort() {
	// TODO: evoke abort through ir_client
	abort_request := &Request{
		Op:    OP_ABORT,
		TxnID: c.t_id,
	}
	c.ir_client.InvokeInconsistent(abort_request)
	c.Continue()
}

/** IR support method: TAPIR decide algorithm */
func (c *TapirClientImpl) decide(results []*Response) *Response {
	// Merges inconsistent Prepare results from replicas into a single result
	ok_count := 0
	abstain_count := 0
	var max_retry_ts *Timestamp = nil

	for _, result_struct := range results {
		result := result_struct.Status
		if result == RPLY_OK {
			ok_count++
		}
		if result == RPLY_ABORT {
			return NewResponse(RPLY_ABORT)
		}
		if result == RPLY_ABSTAIN {
			abstain_count++
		}
		if result == RPLY_RETRY {
			max_retry_ts = LaterTime(max_retry_ts, NewTimestamp(c.client_id))
		}
	}

	if ok_count > c.quorum_size {
		return NewResponse(RPLY_OK)
	}

	if abstain_count > c.quorum_size {
		return NewResponse(RPLY_ABORT)
	}

	if max_retry_ts != nil {
		return NewResponseWithTime(RPLY_RETRY, max_retry_ts)
	}

	return NewResponse(RPLY_ABORT)
}

func (c *TapirClientImpl) String() string {
	return fmt.Sprintf("TAPIR Client {\n"+
		"  id: %d,\n"+
		"  ongoing_transaction: %d,\n"+
		"  closest_replica_id: %d\n"+
		"}",
		c.client_id, c.t_id, c.replica_id)
}

func (c *TapirClientImpl) Continue() {
	log.Println("Releaseing lock")
	c.lock.Unlock()
}
