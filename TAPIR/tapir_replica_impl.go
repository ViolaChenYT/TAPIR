package tapir

import (
	. "github.com/ViolaChenYT/TAPIR/TAPIR/versionstore"
	. "github.com/ViolaChenYT/TAPIR/common"
)

type TimedTransaction struct {
	txn  *Transaction
	time *Timestamp
}

// TapirReplicaImpl represents an implementation of the TapirReplica interface
type TapirReplicaImpl struct {
	store    VersionedKVStore          // versioned data store
	prepared map[int]*TimedTransaction // list of transactions replica is prepared to commit
}

func NewReplica() TapirReplica {
	return &TapirReplicaImpl{
		store:    NewVersionedKVStore(),
		prepared: make(map[int]*TimedTransaction),
	}
}

func (r *TapirReplicaImpl) Prepare(txn *Transaction, timestamp *Timestamp) (*Response, error) {
	// Check prepared for txn.id
	if prepared_txn, ok := r.prepared[txn.ID]; ok {
		if prepared_txn.time.Equals(timestamp) {
			// Transaction already prepared
			return NewResponse(RPLY_OK), nil
		} else {
			// Re-run the checks again for a new timestamp
			delete(r.prepared, txn.ID)
		}
	}

	// Run OCC checks
	return r.occCheck(txn, timestamp), nil
}

func (r *TapirReplicaImpl) Read(key string) (string, *Timestamp, error) {
	// Returns value and version, where version is the timestamp of the transaction that wrote that version

	return "", NewTimestamp(0), nil
}

func (r *TapirReplicaImpl) Commit(txnID int, timestamp *Timestamp) error {
	timedTxn := r.prepared[txnID]

	// Updates its versioned store
	_, readTimes := timedTxn.txn.GetReadSet()
	for key, version := range readTimes {
		// Update version for read operations
		r.store.CommitGet(key, version, timestamp)
	}
	for key, value := range timedTxn.txn.GetWriteSet() {
		// Update value and version for write operations
		r.store.Put(key, value, timestamp)
	}

	// Removes the transaction from prepared list
	delete(r.prepared, txnID)
	return nil
}

func (r *TapirReplicaImpl) Abort(txnID int) error {
	// Removes the transaction from prepared list
	delete(r.prepared, txnID)
	return nil
}

// Private functions

func (r *TapirReplicaImpl) occCheck(txn *Transaction, timestamp *Timestamp) *Response {
	preparedReads := r.getPreparedReads()
	preparedWrites := r.getPreparedWrites()

	readVals, readTimes := txn.GetReadSet()
	for key := range readVals {
		version := readTimes[key]
		lastVersionedVal, ok := r.store.Get(key)

		if !ok {
			// No conflict if we don't have this version
			continue
		}

		if version.LessThan(lastVersionedVal.WriteTime) {
			return NewResponse(RPLY_ABORT)
		} else if len(preparedWrites[key]) > 0 && version.LessThan(MinTimestamp(preparedWrites[key])) {
			return NewResponse(RPLY_ABSTAIN)
		}
	}

	for key := range txn.GetWriteSet() {
		lastRead, ok := r.store.GetLastRead(key, timestamp)

		if !ok {
			// No conflict if it has not been read
			continue
		}

		if timestamp.LessThan(MaxTimestamp(preparedReads[key])) {
			lastPreparedRead := MaxTimestamp(preparedReads[key])
			return NewResponseWithTime(RPLY_RETRY, lastPreparedRead)
		} else if timestamp.LessThan(lastRead) {
			return NewResponseWithTime(RPLY_RETRY, lastRead)
		}
	}

	r.prepared[txn.ID] = &TimedTransaction{txn, timestamp}

	return NewResponse(RPLY_OK)
}

// Return timestamps of prepared reads
func (r *TapirReplicaImpl) getPreparedReads() map[string][]*Timestamp {
	reads := make(map[string][]*Timestamp)
	for _, timedTxn := range r.prepared {
		readVals, _ := timedTxn.txn.GetReadSet()
		for key := range readVals {
			reads[key] = append(reads[key], timedTxn.time)
		}
	}
	return reads
}

// Return timestamps of prepared writes
func (r *TapirReplicaImpl) getPreparedWrites() map[string][]*Timestamp {
	writes := make(map[string][]*Timestamp)
	for _, timedTxn := range r.prepared {
		for key := range timedTxn.txn.GetWriteSet() {
			writes[key] = append(writes[key], timedTxn.time)
		}
	}
	return writes
}
