package tapir

import (
	"bufio" //
	"fmt"
	"net"
	"net/rpc"
	"time"
	"IR/ir_client"
)

// TapirClientImpl is an implementation of the TapirClient interface
type TapirClientImpl struct {
	// Unique ID for this client
	int client_id;

	// Ongoing transaction ID
	int t_id;

	// Buffered transaction 
	Transaction txn;

	// IR protocol client
	IR.Client ir_client;

	// Closet replica for read ops
	int replica_id;
}

func NewClient(id int, closest_replica int) (*TapirClientImpl, error) {
	client := TapirClientImpl{
		client_id:        id,
		t_id:              0, // TODO: change
		replica_id:       closest_replica
	}

	// Create replica proxy
	client.ir_client = IR.NewClient();

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
	if (val, timestamp) ok := c.txn.GetReadSet()[key]; ok {
		return val, nil
	}
	// Otherwise, the client sends Read(key) to the replica
	// TODO: send request to read from closest replica 
	// c.ir_client.InvokeUnlogged(c.closest_replica)

	// On response, client puts (key, version) into the transaction's read set, and returns object to the application
	val, timestamp = "", time.Time{} // Placeholders

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
	timestamp = (time.Time(), c.client_id)
	// Client invokes Prepare(tx, timestamp) as an IR consensus operation.
	reply = c.ir_client.InvokeConsensus(c.decide) // pass decide function 

	if reply.result == PREPARE_OK {
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
	c.ir_client.InvokeInconsistent()
}


/** IR support method: TAPIR decide algorithm */
func (c *TapirClientImpl) decide(results Result[]) {
	// Merges inconsistent Prepare results from replicas into a single result 
	int ok_count = 0
	int abstain_count = 0
	time.Time max_retry_ts = 0;

	for result := range results {
		if result == PREPARE_OK {
			ok_count++
		}
		if result == ABORT {
			return ABORT
		}
		if result == ABSTAIN {
			abstain_count++
		}
		if result == RETRY {
			max_retry_ts = max(max_retry_ts, result.timestamp)
		}
	}

	if ok_count > QUORUM_SIZE {
		return PREPARE_OK
	}

	if abstain_count > QUORUM_SIZE {
		return ABORT
	}

	if max_retry_ts {
		return RETRY, max_retry_ts 
	}

	return ABORT
}