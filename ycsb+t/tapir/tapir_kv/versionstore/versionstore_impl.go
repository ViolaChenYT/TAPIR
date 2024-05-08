package versionstore

import (
	"fmt"
	"log"
	"sort"
	"sync"

	. "github.com/pingcap/go-ycsb/tapir/common"
)

type VersionedKVStoreImpl struct {
	store     map[string][]*VersionedValue           // <key, (write_time, value)> pairs of storage
	lastReads map[string](map[*Timestamp]*Timestamp) // <key, <write_time, last_read_time>> recording last read time of each version
	storelock sync.Mutex
	readslock sync.Mutex
}

func NewVersionedKVStore() VersionedKVStore {
	return &VersionedKVStoreImpl{
		store:     make(map[string][]*VersionedValue),
		lastReads: make(map[string](map[*Timestamp]*Timestamp)),
	}
}

func EmptyEntry() *VersionedValue {
	return &VersionedValue{
		WriteTime: EmptyTime(),
		Value:     "",
	}
}

func (vs *VersionedKVStoreImpl) Get(key string) (*VersionedValue, bool) {
	versionedVals, ok := vs.store[key]
	if !ok {
		// key not found
		return EmptyEntry(), false
	}

	if len(versionedVals) > 0 {
		return versionedVals[len(versionedVals)-1], true // Return the latest value
	}
	return EmptyEntry(), false
}

func (vs *VersionedKVStoreImpl) Put(key string, value string, time *Timestamp) {
	log.Println("Commiting to KV: ", key, value)
	vs.storelock.Lock()
	if key_entry, ok := vs.store[key]; ok {
		vs.store[key] = append(key_entry, &VersionedValue{WriteTime: time, Value: value})
	} else {
		log.Println("New entry created")
		vs.store[key] = []*VersionedValue{{WriteTime: time, Value: value}}
	}
	vs.storelock.Unlock()
}

func (vs *VersionedKVStoreImpl) CommitGet(key string, readTime *Timestamp, commitTime *Timestamp) {
	// Create the <version, last_read_time> map if not exists
	vs.readslock.Lock()
	if vs.lastReads[key] == nil {
		vs.lastReads[key] = make(map[*Timestamp]*Timestamp)
	}
	vs.lastReads[key][readTime] = commitTime
	vs.readslock.Unlock()
}

func (vs *VersionedKVStoreImpl) GetLastRead(key string, time *Timestamp) (*Timestamp, bool) {
	versionedVal, ok := vs.getValue(key, time)
	if !ok {
		// key not found
		return EmptyTime(), false
	}

	return vs.lastReads[key][versionedVal.WriteTime], true
}

func (vs *VersionedKVStoreImpl) GetRange(key string, time *Timestamp) (*Timestamp, *Timestamp, bool) {
	versionedVals, ok := vs.store[key]
	if !ok {
		// key not found
		return EmptyTime(), EmptyTime(), false
	}

	var startTime *Timestamp = EmptyTime()
	var endTime *Timestamp = EmptyTime()
	valid := false
	// Iterate through the versionedVals backwards
	for i := len(versionedVals) - 1; i >= 0; i-- {
		if versionedVals[i].WriteTime.LessThanOrEqualTo(time) {
			startTime = versionedVals[i].WriteTime
			if i < len(versionedVals)-2 {
				endTime = versionedVals[i+1].WriteTime
			}
			valid = true
			break
		}
	}
	return startTime, endTime, valid
}

// Return <value, write_time> valid at the given timestamp
func (vs *VersionedKVStoreImpl) getValue(key string, validTime *Timestamp) (*VersionedValue, bool) {
	versionedVals, ok := vs.store[key]
	if !ok {
		// key not found
		return EmptyEntry(), false
	}

	// Iterate through the versionedVals backwards
	for i := len(versionedVals) - 1; i >= 0; i-- {
		if versionedVals[i].WriteTime.LessThanOrEqualTo(validTime) {
			return versionedVals[i], true // Found valid value
		}
	}

	return EmptyEntry(), false
}

func (kv *VersionedKVStoreImpl) String() string {
	result := "VersionedKVStore:\n"
	for key, versions := range kv.store {
		result += fmt.Sprintf("Key: %s\n", key)
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].WriteTime.LessThan(versions[j].WriteTime)
		})
		for _, vv := range versions {
			lastRead := kv.lastReads[key][vv.WriteTime]
			result += fmt.Sprintf("\tWrite Time: %v, Value: %v, Last Read Time: %v\n",
				vv.WriteTime, vv.Value, lastRead)
		}
	}
	return result
}