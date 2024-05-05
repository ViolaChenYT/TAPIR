package tapir

import (
	//

	"log"

	"github.com/ViolaChenYT/TAPIR/IR"
)

const QUORUM_SIZE = 3 // TODO: change

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
}

func NewClient(id int, closest_replica int) (*TapirClientImpl, error) {
	client := TapirClientImpl{
		client_id:  id,
		t_id:       0, // TODO: change
		replica_id: closest_replica,
	}

	// Create replica proxy
	cl, err := IR.NewClient(id, []string{"8080"})
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
	c.t_id++

	// Create a transaction
	c.txn = NewTransaction(c.t_id)
}

func (c *TapirClientImpl) Read(key string) (string, error) {
	// If key is in the transaction's write set, the client returns value from the write set
	if val, ok := c.txn.GetWriteSet()[key]; ok {
		return val, nil
	}
	// If the transaction has already read key, it returns a cached copy
	readSet, timeset := c.txn.GetReadSet()
	if val, ok := readSet[key]; ok {
		return val, nil
	}
	timestamp := timeset[key]
	// Otherwise, the client sends Read(key) to the replica
	// TODO: send request to read from closest replica
	// c.ir_client.InvokeUnlogged(c.closest_replica)
	// c.ir_client.

	// On response, client puts (key, version) into the transaction's read set, and returns object to the application
	val, timestamp := "", timestamp // Placeholders

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
	reply, err := c.ir_client.InvokeConsensus(c.decide) // pass decide function
	if err != nil {
		log.Panicf("Error invoking consensus: %v", err)
		return false
	}
	if reply.State == IR.PREPARE_OK {
		// Commit to all replicas
		c.ir_client.InvokeInconsistent() // TODO: how to evoke Commit() on replicas?
		return true
	}

	// TODO: handle retry

	// Otherwise, abort
	c.Abort()
	return false
}

func (c *TapirClientImpl) Abort() {
	// TODO: evoke abort through ir_client
	current_op := IR.Operation{} // Placeholder, should be the current operation
	c.ir_client.InvokeInconsistent(current_op)
}

/** IR support method: TAPIR decide algorithm */
func (c *TapirClientImpl) decide(results []*IR.Result) PrepareResult {
	// Merges inconsistent Prepare results from replicas into a single result
	ok_count := 0
	abstain_count := 0
	var max_retry_ts *Timestamp = nil

	for _, result_struct := range results {
		result := result_struct.State
		if result == IR.PREPARE_OK {
			ok_count++
		}
		if result == IR.ABORT {
			return NewPrepareResult(ABORT)
		}
		if result == IR.ABSTAIN {
			abstain_count++
		}
		if result == IR.RETRY {
			max_retry_ts = laterTime(max_retry_ts, NewTimestamp(c.client_id))
		}
	}

	if ok_count > QUORUM_SIZE {
		return NewPrepareResult(PREPARE_OK)
	}

	if abstain_count > QUORUM_SIZE {
		return NewPrepareResult(ABORT)
	}

	if max_retry_ts != nil {
		return CreateRetry(*max_retry_ts)
	}

	return NewPrepareResult(ABORT)
}
