package tapir

type VersionedKVStoreImpl struct {
	store     map[string][]VersionedValue          // <key, (write_time, value)> pairs of storage
	lastReads map[string](map[Timestamp]Timestamp) // <key, <write_time, last_read_time>> recording last read time of each version
}

func NewVersionedKVStoreImpl() *VersionedKVStoreImpl {
	return &VersionedKVStoreImpl{
		store:     make(map[string][]VersionedValue),
		lastReads: make(map[string](map[Timestamp]Timestamp)),
	}
}

func (vs *VersionedKVStoreImpl) Get(key string) (VersionedValue, bool) {
	versionedVals, ok := vs.store[key]
	if !ok {
		// key not found
		return VersionedValue{}, false
	}

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
	versionedVal, ok := vs.getValue(key, time)
	if !ok {
		// key not found
		return Timestamp{}, false
	}

	return vs.lastReads[key][versionedVal.write_time], true
}

func (vs *VersionedKVStoreImpl) GetRange(key string, time Timestamp) (Timestamp, Timestamp, bool) {
	versionedVals, ok := vs.store[key]
	if !ok {
		// key not found
		return Timestamp{}, Timestamp{}, false
	}

	startTime := Timestamp{}
	endTime := Timestamp{}
	valid := false
	// Iterate through the versionedVals backwards
	for i := len(versionedVals) - 1; i >= 0; i-- {
		if versionedVals[i].write_time.LessThanOrEqualTo(time) {
			startTime = versionedVals[i].write_time
			if i < len(versionedVals)-2 {
				endTime = versionedVals[i+1].write_time
			}
			valid = true
			break
		}
	}
	return startTime, endTime, valid
}

// Return <value, write_time> valid at the given timestamp
func (vs *VersionedKVStoreImpl) getValue(key string, validTime Timestamp) (VersionedValue, bool) {
	versionedVals, ok := vs.store[key]
	if !ok {
		// key not found
		return VersionedValue{}, false
	}

	// Iterate through the versionedVals backwards
	for i := len(versionedVals) - 1; i >= 0; i-- {
		if versionedVals[i].write_time.LessThanOrEqualTo(validTime) {
			return versionedVals[i], true // Found valid value
		}
	}

	return VersionedValue{}, false
}
