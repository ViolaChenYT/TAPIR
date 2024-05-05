package versionstore

type VersionedKVStoreImpl struct {
	store     map[string][]VersionedValue            // <key, (write_time, value)> pairs of storage
	lastReads map[string]map[Timestamp, Timestamp]   // <key, <write_time, last_read_time>> recording last read time of each version
}

func NewVersionedKVStoreImpl() *VersionedKVStoreImpl {
	return &VersionedKVStoreImpl{
		store: make(map[string][]VersionedValue),
		lastReads: make(map[string]map[Timestamp, Timestamp])
	}
}

func (vs *VersionedKVStoreImpl) Get(key string) (VersionedValue, bool) {
	versionedVals := vs.store[key]
	if len(versionedVals) > 0 {
		return versionedVals[len(versionedVals)-1], true // Return the latest value
	}
	return VersionedValue{}, false
}

func (vs *VersionedKVStoreImpl) Put(key string, value string, time Timestamp) {
	vs.store[key] = append(vs.store[key], VersionedValue{write_time: time, value: value})
}

func (vs *VersionedKVStoreImpl) CommitGet(key string, readTime Timestamp, commitTime Timestamp) {
	vs.lastReads[key][readTime] = commitTime
}

func (vs *VersionedKVStoreImpl) GetLastRead(key string, time Timestamp) (Timestamp, bool) {
	// Check if someone has read this version before 
	if lastRead, ok := vs.lastReads[key][]
}