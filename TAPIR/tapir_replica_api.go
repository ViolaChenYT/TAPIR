package tapir

// TapirReplica represents a Key-value store with support for transactions using TAPIR.
type TapirReplica interface {

	// Begin a transaction
	Prepare(txn Transaction, timestamp Timestamp) (Result, Timestamp, error)

	// Read the value corresponding to key, return value and version
	Read(key string) (string, Timestamp, error)

	// Commit the transaction
	Commit(txn Transaction, timestamp Timestamp) error

	// Abort the transaction
	Abort(txn Transaction, timestamp Timestamp) error
}
