package versionstore

import (
	tapir "github.com/ViolaChenYT/TAPIR"
)

type VersionedValue struct {
	write_time tapir.Timestamp
	value      string
}

// Define VersionedKVStore interface
type VersionedKVStore interface {
	// Read the most recent value and timestamp of the given key
	Get(key string) (VersionedValue, bool)

	// Write the given key-value pair to the store
	Put(key string, value string, time tapir.Timestamp)

	// Commit a read by udpating the timestamp of the latest read transaction for the version of the key that the transaction read
	CommitGet(key string, readTime tapir.Timestamp, commitTime Timestamp)

	// Get the last read for the write valid at the given timestamp
	GetLastRead(key string, time Timestamp) (Timestamp, bool)
}
