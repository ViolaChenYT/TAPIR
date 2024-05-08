package tapir_kv

import (
	"errors"
	"fmt"
	"log"

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
}

func NewReplica(id int) TapirReplica {
	r := TapirReplicaImpl{
		store:    NewVersionedKVStore(),
		prepared: make(map[int]*TimedTransaction),
		ID:       id,
	}
	return &r
}

func (r *TapirReplicaImpl) Prepare(txn *Transaction, timestamp *Timestamp) (*Response, error) {
	// Check prepared for txn.id
	log.Println(r.ID, "Trying Preparing transaction", txn)
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
		return versionedVal.Value, versionedVal.WriteTime, errors.New(fmt.Sprintf("Key %s not exist in replica %d.", key, r.ID))
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
	if timedTxn == nil {
		log.Panicln("AHHHHHHHHHHHHHHHHHHH")
	}
	log.Println(timedTxn.txn)
	readTimes := timedTxn.txn.ReadTime
	for key, version := range readTimes {
		// Update version for read operations
		log.Println("About to call Commit Get for key: ", key)
		r.store.CommitGet(key, version, timestamp)
	}
	for key, value := range timedTxn.txn.WriteSet {
		// Update value and version for write operations
		log.Println("About to call Put for key: ", key)
		r.store.Put(key, value, timestamp)
	}

	log.Println("Current store ", r.ID, r.store)

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
	log.Println(txn)
	preparedReads := r.getPreparedReads()
	preparedWrites := r.getPreparedWrites()

	readVals, readTimes := txn.ReadSet, txn.ReadTime
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

	for key := range txn.WriteSet {
		lastRead, ok := r.store.GetLastRead(key, timestamp)

		if !ok {
			// No conflict if it has not been read
			continue
		}
		maxReadTimestamp := MaxTimestamp(preparedReads[key])
		if maxReadTimestamp == nil {
			continue
		} else if timestamp.LessThan(maxReadTimestamp) {
			// log.Panicln("possible B")
			lastPreparedRead := maxReadTimestamp
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
		readVals := timedTxn.txn.ReadSet
		for key := range readVals {
			if readsKey, ok := reads[key]; ok {
				reads[key] = append(readsKey, timedTxn.time)
			} else {
				reads[key] = []*Timestamp{timedTxn.time}
			}
		}
	}
	return reads
}

// Return timestamps of prepared writes
func (r *TapirReplicaImpl) getPreparedWrites() map[string][]*Timestamp {
	writes := make(map[string][]*Timestamp)
	for _, timedTxn := range r.prepared {
		for key := range timedTxn.txn.WriteSet {
			if writesKey, ok := writes[key]; ok {
				writes[key] = append(writesKey, timedTxn.time)
			} else {
				writes[key] = []*Timestamp{timedTxn.time}
			}
		}
	}
	return writes
}
