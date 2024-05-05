package tapir

type TimedTransaction struct {
	txn  Transaction
	time Timestamp
}

// TapirReplicaImpl represents an implementation of the TapirReplica interface
type TapirReplicaImpl struct {
	store    Libstore                 // versioned data store
	prepared map[int]TimedTransaction // list of transactions replica is prepared to commit
}

func NewReplica(store Libstore) *TapirReplicaImpl {
	return &TapirReplicaImpl{
		store:    store,
		prepared: make(map[int]TimedTransaction),
	}
}

func (r *TapirReplicaImpl) Prepare(txn Transaction, timestamp Timestamp) (PrepareResult, error) {
	// Check prepared for txn.id
	if prepared_txn, ok := r.prepared[txn.id]; ok {
		if prepared_txn.time.Equals(timestamp) {
			// Transaction already prepared
			return NewPrepareResult(PREPARE_OK), nil
		} else {
			// Re-run the checks again for a new timestamp
			delete(r.prepared, txn.id)
		}
	}

	// Run OCC checks
	return r.occCheck(txn, timestamp), nil
}

func (r *TapirReplicaImpl) Read(key string) (string, Timestamp, error) {
	// Returns value and version, where version is the timestamp of the transaction that wrote that version

	return "", *NewTimestamp(0), nil
}

func (r *TapirReplicaImpl) Commit(txn Transaction, timestamp Timestamp) error {
	// Updates its versioned store
	_, readTimes := txn.GetReadSet()
	for key, version := range readTimes {
		// Update version for read operations
		r.store.commitGet(key, version, timestamp)
	}
	for key, value := range txn.GetWriteSet() {
		// Update value and version for write operations
		r.store.put(key, value, timestamp)
	}

	// Removes the transaction from prepared list
	delete(r.prepared, txn.id)
}

func (r *TapirReplicaImpl) Abort(txn Transaction) error {
	// Removes the transaction from prepared list
	delete(r.prepared, txn.id)
	return nil
}

// Private functions

func (r *TapirReplicaImpl) occCheck(txn Transaction, timestamp Timestamp) PrepareResult {
	preparedReads := r.getPreparedReads()
	preparedWrites := r.getPreparedWrites()

	readVals, readTimes := txn.GetReadSet()
	for key := range readVals {
		version := readTimes[key]
		if version.LessThan(store[key].latest_version) {
			return NewPrepareResult(ABORT)
		} else if len(preparedWrites[key]) > 0 && version.LessThan(minTimestamp(preparedWrites[key])) {
			return NewPrepareResult(ABSTAIN)
		}
	}

	for key := range txn.GetWriteSet() {
		if timestamp.LessThan(maxTimestamp(preparedReads[key])) {
			return CreateRetry(maxTimestamp(preparedReads[key]))
		} else if timestamp.LessThan(store[key].latest_version) {
			return CreateRetry(store[key].latest_version)
		}
	}

	r.prepared[txn.id] = TimedTransaction{txn, timestamp}

	return NewPrepareResult(PREPARE_OK)
}

// Return timestamps of prepared reads
func (r *TapirReplicaImpl) getPreparedReads() map[string][]Timestamp {
	reads := make(map[string][]Timestamp)
	for _, timedTxn := range r.prepared {
		readVals, _ := timedTxn.txn.GetReadSet()
		for key := range readVals {
			reads[key] = append(reads[key], timedTxn.time)
		}
	}
	return reads
}

// Return timestamps of prepared writes
func (r *TapirReplicaImpl) getPreparedWrites() map[string][]Timestamp {
	writes := make(map[string][]Timestamp)
	for _, timedTxn := range r.prepared {
		for key := range timedTxn.txn.GetWriteSet() {
			writes[key] = append(writes[key], timedTxn.time)
		}
	}
	return writes
}
