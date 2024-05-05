package tapir

import (
	"time"
	"github.com/ViolaChenYT/TAPIR/IR/libstore/libstore_impl"
)

// TapirReplicaImpl represents an implementation of the TapirReplica interface
type TapirReplicaImpl struct {
	store            Libstore                            // versioned data store
	prepared         map[int](Timestamp, transaction) // list of transactions replica is prepared to commit
}

func NewReplica(store Libstore) *TapirReplicaImpl {
	return &TapirReplicaImpl{
		store:           store,
		prepared:        make(map[int](Timestamp, transaction))
	}
}

func (r *TapirReplicaImpl) Prepare(txn Transaction, timestamp Timestamp) (Result, Timestamp, error) {
	// Check prepared for txn.id
	if log_timestamp, _, ok := r.prepared[txn.id]; ok {
		if log_timestamp.Equals(timestamp) {
			// Transaction already prepared
			return PREPARE_OK, timestamp, nil
		} else {
			// Re-run the checks again for a new timestamp
			delete(r.prepared, txn.id)
		}
	}

	// Run OCC checks
	return r.occCheck(txn, timestamp)
}

func (r *TapirReplicaImpl) Read(key string) (string, Timestamp, error) {
	// Returns value and version, where version is the timestamp of the transaction that wrote that version
	
	return "", "", nil
}

func (r *TapirReplicaImpl) Commit(txn Transaction, timestamp Timestamp) error {
	// Updates its versioned store 
	for key, version := range txn.ReadSet {
		// Update version for read operations
		store.commitGet(key, version, timestamp)
	}
	for key, value, version := range txn.WriteSet {
		// Update value and version for write operations
		store.put(key, value, version, timestamp)
	}

	// Removes the transaction from prepared list 
	delete(r.prepared, txn.id)
}

func (r *TapirReplicaImpl) Abort() (txn Transaction, timestamp Timestamp) error {
	// Removes the transaction from prepared list 
	delete(r.prepared, txn.id)
}

// Private functions 

func (r *TapirReplicaImpl) occCheck(txn Transaction, timestamp Timestamp) (PrepareResult) {
	preparedReads = r.getPreparedReads()
	preparedWrites = r.getPreparedWrites()

	for key, version := range txn.GetReadSet() {
		if version.Before(store[key].latest_version) {
			return NewPrepareResult(ABORT)
		} else if len(preparedWrites[key]) > 0 && version.Before(minTimestamp(preparedWrites[key])) {
			return NewPrepareResult(ABSTAIN)
		}
	}

	for key := range txn.GetWriteSet() {
		if txn.Timestamp.LessThan(maxTimestamp(preparedReads[key])) {
			return CreateRetry(maxTimestamp(preparedReads[key]))
		} else if txn.Timestamp.LessThan(store[key].latest_version) {
			return CreateRetry(store[key].latest_version)
		}
	}

	r.prepared[txn.id] = timestamp, txn

	return NewPrepareResult(PREPARE_OK)
}

// Return timestamps of prepared reads
func (r *TapirReplicaImpl) getPreparedReads() map[string][]Timestamp {
	reads := make(map[string][]Timestamp)
	for timestamp, txn := range r.prepared {
		for key := range txn.GetReadSet() {
			append(reads[key], timestamp)
		}
	}
	return reads
}

// Return timestamps of prepared writes
func (r *TapirReplicaImpl) getPreparedWrites() map[string][]Timestamp {
	writes := make(map[string][]Timestamp)
	for timestamp, txn := range r.prepared {
		for key := range txn.GetWriteSet() {
			append(writes[key], timestamp)
		}
	}
	return writes
}