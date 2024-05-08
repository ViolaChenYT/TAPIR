package tapir_kv

import (
	. "github.com/pingcap/go-ycsb/tapir/common"
)

// TapirReplica represents a Key-value store with support for transactions using TAPIR.
type TapirReplica interface {

	// Begin a transaction
	Prepare(txn *Transaction, timestamp *Timestamp) (*Response, error)

	// Read the value corresponding to key, return value and version
	Read(key string) (string, *Timestamp, error)

	// Commit the transaction
	Commit(txnID int, timestamp *Timestamp) error

	// Abort the transaction
	Abort(txnID int) error
}
