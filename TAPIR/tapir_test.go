package tapir

import (
	"fmt"
	"log"
	"testing"
	"time"

	IR "github.com/ViolaChenYT/TAPIR/IR"
	. "github.com/ViolaChenYT/TAPIR/common"
)

const (
	replica_id = 0
	txn_id     = 123
	key0       = "hello"
	val0       = "world"
	key1       = "viola"
	val1       = "chen"
	key2       = "ruyu"
	val2       = "yan"
)

// createAscendingTimes generates a slice of Timestamps with ascending timestamps
func createAscendingTimes(count int) []*Timestamp {
	now := time.Now()
	output := make([]*Timestamp, count)
	for i := 0; i < count; i++ {
		output[i] = NewCustomTimestamp(i, now.Add(time.Duration(i)*time.Second))
	}
	return output
}

func TestReplicaSetup(t *testing.T) {
	replica := NewReplica(replica_id)
	val, timestamp, err := replica.Read(key0)

	// Check if the returned values meet expectations
	if val != "" {
		t.Errorf("Expected val to be empty, got: %s", val)
	}
	if timestamp != nil {
		t.Errorf("Expected timestamp to be nil, got: %v", timestamp)
	}
	if err == nil {
		t.Errorf("Expected err to be not nil, got: %v", err)
	}
}

func TestReplicaCommit(t *testing.T) {
	// Grab some timestamps for laster use
	timestamps := createAscendingTimes(5)

	// Create test transaction
	txn := NewTransaction(txn_id)
	txn.AddWriteSet(key0, val0)
	txn.AddWriteSet(key1, val1)
	txn.AddReadSet(key0, val0, timestamps[0])

	replica := NewReplica(replica_id)
	response, err := replica.Prepare(txn, timestamps[1])

	// Check prepare status
	if response.Status != RPLY_OK {
		t.Errorf("Expected prepare response OK, got: %v", response.Status)
	}
	if err != nil {
		t.Errorf("Expected prepare without error, got: %v", err)
	}

	err = replica.Commit(txn.ID, timestamps[2])
	// Check prepare status
	if err != nil {
		t.Errorf("Expected commit without error, got: %v", err)
	}

	// Check if commit values can be read
	val, timestamp, err := replica.Read(key0)

	// Check if the returned values meet expectations
	if val != val0 {
		t.Errorf("Expected val to be %s, got: %s", val0, val)
	}
	if timestamp != timestamps[2] {
		t.Errorf("Expected timestamp to be %v, got: %v", timestamps[2], timestamp)
	}
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
}

func TestSetup(t *testing.T) {
	port := 55209
	replica := NewReplica(port)
	log.Println("ok", replica)
	time.Sleep(time.Second)
	client, err := IR.NewClient(1, []string{fmt.Sprintf("%d", port)})
	if err != nil {
		t.Fatal("Failed to dial server:", err)
	}
	log.Println("ok", client)
}
