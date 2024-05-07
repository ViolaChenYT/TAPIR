package tapir

import (
	"errors"
	"fmt"
	"log"
	"net"

	. "github.com/ViolaChenYT/TAPIR/common"
	. "github.com/ViolaChenYT/TAPIR/tapir_kv/versionstore"
)

type TimedTransaction struct {
	txn  *Transaction
	time *Timestamp
}

// TapirReplicaImpl represents an implementation of the TapirReplica interface
type TapirReplicaImpl struct {
	store    VersionedKVStore          // versioned data store
	prepared map[int]*TimedTransaction // list of transactions replica is prepared to commit
	ID       int                       // same as corredponding tapir server ID, may change
	listener net.Listener
	close    chan bool
}

func NewReplica(id int) TapirReplica {
	r := TapirReplicaImpl{
		store:    NewVersionedKVStore(),
		prepared: make(map[int]*TimedTransaction),
		ID:       id,
		close:    make(chan bool),
	}
	return &r
}

// the rpc function

func (r *TapirReplicaImpl) Close() error {
	r.listener.Close()
	r.close <- true
	return nil
}

func (r *TapirReplicaImpl) Prepare(txn *Transaction, timestamp *Timestamp) (*Response, error) {
	// Check prepared for txn.id
	log.Println(r.ID, "Trying Preparing transaction", txn.ID)
	if prepared_txn, ok := r.prepared[txn.ID]; ok {
		if prepared_txn.time.Equals(timestamp) {
			// Transaction already prepared
			return NewResponse(RPLY_OK), nil
		} else {
			// Re-run the checks again for a new timestamp
			delete(r.prepared, txn.ID)
		}
	} else {
		// New transaction
		newtime := NewTimestamp(timestamp.ID)
		r.prepared[txn.ID] = &TimedTransaction{txn, newtime}
		return r.occCheck(txn, newtime), nil
	}

	// Run OCC checks
	return r.occCheck(txn, timestamp), nil
}

func (r *TapirReplicaImpl) Read(key string) (string, *Timestamp, error) {
	// Returns value and version, where version is the timestamp of the transaction that wrote that version
	versionedVal, ok := r.store.Get(key)
	if ok {
		return versionedVal.Value, versionedVal.WriteTime, nil
	} else {
		return "", nil, errors.New(fmt.Sprintf("Key %s not exist in replica %d.", key, r.ID))
	}
}

func (r *TapirReplicaImpl) Commit(txnID int, timestamp *Timestamp) error {
	// for id, timedTxn := range r.prepared {
	// 	log.Println("Prepared transaction", id, ":", timedTxn)
	// }
	log.Println(r.ID, "currently", len(r.prepared), "prepared transactions-----------------")
	timedTxn := r.prepared[txnID]

	// Updates its versioned store
	log.Println("Committing transaction", txnID, "trying to get read set")
	log.Println(timedTxn)
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
	log.Println(r.ID, "deleting transaction", txnID)
	delete(r.prepared, txnID)
	return nil
}

func (r *TapirReplicaImpl) Abort(txnID int) error {
	// Removes the transaction from prepared list
	log.Println(r.ID, "Aborting transaction", txnID)
	delete(r.prepared, txnID)
	return nil
}

// Private functions

func (r *TapirReplicaImpl) occCheck(txn *Transaction, timestamp *Timestamp) *Response {
	log.Println("-----------------------Running OCC check for transaction", txn.ID, "on replica", r.ID)
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
