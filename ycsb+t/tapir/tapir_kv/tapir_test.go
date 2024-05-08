package tapir_kv

import (
	"fmt"
	"log"
	"testing"
	"time"

	. "github.com/pingcap/go-ycsb/tapir/IR"
	. "github.com/pingcap/go-ycsb/tapir/common"
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

// var replica_addresses = []string{"55209", "55210", "55211"}
// testing in terminal:
// go test -run <name of specific test>
// eg. go test -run TestReplicaSetup

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

func TestSimpleCommit(t *testing.T) {
	config := GetConfigA()
	var replica IRReplica = nil
	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		replica = NewIRReplica(id, addr, store)
	}
	log.Println("Replica created: ", replica)

	client, err := NewTapirClient(config)

	client.Begin()
	client.Write(key0, val0)

	val, err := client.Read(key0)
	if val != val0 {
		t.Errorf("Expected val to be %s, got: %s", val0, val)
	}
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}

	log.Println("test commit")
	ok := client.Commit()
	if !ok {
		t.Errorf("Commit failed, expected to suceed")
	}

	log.Println("ok commit")
}

func TestSimpleReadFromStore(t *testing.T) {
	config := GetConfigA()

	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		NewIRReplica(id, addr, store)
	}

	client, _ := NewTapirClient(config)
	// First Transaction: Commit a write
	client.Begin()
	client.Write(key0, val0)
	ok := client.Commit()
	// time.Sleep(time.Second)

	if !ok {
		t.Errorf("First commit failed, expected to suceed")
	}
	log.Println("Write transaction done!")

	// Second Transaction: Read from the previous written entry
	client.Begin()
	val, err := client.Read(key0)
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
	if val != val0 {
		t.Errorf("Expected val to be %s, got: %s", val0, val)
	}

	ok = client.Commit()
	if !ok {
		t.Errorf("Second commit failed, expected to suceed")
	}
}

func TestSimpleAbort(t *testing.T) {
	config := GetConfigA()

	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		NewIRReplica(id, addr, store)
	}

	client, _ := NewTapirClient(config)
	// First Transaction: Commit a write
	client.Begin()
	client.Write(key0, val0)
	client.Commit()

	// Second Transaction: Abort a write
	client.Begin()
	client.Write(key0, val1)
	val, _ := client.Read(key0)
	if val != val1 {
		t.Errorf("Expected val to be %s, got: %s", val1, val)
	}
	client.Abort()

	// Third Transaction: Read
	client.Begin()
	val, err := client.Read(key0)
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
	if val != val0 {
		t.Errorf("Expected val to be %s, got: %s", val0, val)
	}

	client.Commit()
}

func TestCommit(t *testing.T) {
	fmt.Println("TestCommit")
	config := GetConfigB()
	var replicas = []IRReplica{}
	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		replica := NewIRReplica(id, addr, store)
		replicas = append(replicas, replica)
		log.Println("ok", replica)
		// time.Sleep(time.Second)
	}
	log.Println("ok", replicas[0], replicas[1], replicas[2])
	// client 1:
	// now there's 3 servers
	// closest replica port = 55209
	// all replica = []string{"55209", "55210", "55211"}
	client, err := NewTapirClient(config)
	if err != nil {
		t.Fatal("Failed to dial server:", err)
	}
	log.Println("start testing txn")
	timestamps := createAscendingTimes(5)
	// Create test transaction
	txn := NewTransaction(txn_id)
	txn.AddWriteSet(key0, val0)
	txn.AddWriteSet(key1, val1)
	txn.AddReadSet(key0, val0, timestamps[0])
	client.Begin()
	client.Write(key0, val0)
	client.Write(key1, val1)
	client.Write(key0, val2)
	val, err := client.Read(key0)
	if val != val2 {
		t.Errorf("Expected val to be %s, got: %s", val2, val)
		return
	}
	log.Println("ok read write")
	log.Println("test commit")
	client.Commit()
	log.Println("after commit")
	client.Begin()
	v1, err := client.Read(key1)
	if v1 != val1 {
		t.Errorf("Expected val to be %s, got: %s", val1, v1)
		return
	}
	// time.Sleep(time.Second)
	log.Println("ok commit")
}

func TestMostBasicSetup(t *testing.T) {
	config := GetConfigA()
	// server 1: address localhost:55209
	var replica IRReplica = nil
	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		replica = NewIRReplica(id, addr, store)
		log.Println("ok", replica)
		// time.Sleep(time.Second)
	}
	// client 1:
	// now there's only 1 server
	// closest replica port = 55209
	// all replica = []string{"55209"}
	client, err := NewTapirClient(config)
	if err != nil {
		t.Fatal("Failed to dial server:", err)
	}
	log.Println("ok", client)
	replica.Stop()
}

func Test3ReplicaSetup(t *testing.T) {
	config := GetConfigB()
	var replicas = []IRReplica{}
	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		replica := NewIRReplica(id, addr, store)
		replicas = append(replicas, replica)
		log.Println("ok", replica)
	}
	log.Println("ok", replicas[0], replicas[1], replicas[2])
	// client 1:
	// now there's 3 servers
	// closest replica port = 55209
	// all replica = []string{"55209", "55210", "55211"}
	client, err := NewTapirClient(config)
	if err != nil {
		t.Fatal("Failed to dial server:", err)
	}
	log.Println("ok", client)
}

func TestAbort(t *testing.T) {
	config := GetConfigA()
	var replica IRReplica = nil
	for id, addr := range config.Replicas {
		store := NewTapirServer(id)
		replica = NewIRReplica(id, addr, store)
	}
	log.Println("Replica created: ", replica)

	client, err := NewTapirClient(config)

	client.Begin()
	client.Write(key0, val0)

	val, err := client.Read(key0)
	if val != val0 {
		t.Errorf("Expected val to be %s, got: %s", val0, val)
	}
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}

	log.Println("test commit")
	ok := client.Commit()
	if !ok {
		t.Errorf("Commit failed, expected to suceed")
	}
	client.Abort()
}
