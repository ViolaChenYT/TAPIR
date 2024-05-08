//go:build exclude

package tapir_kv

type Client interface {
	// Begin a transaction.
	Begin()

	// Get the value corresponding to key.
	Get(key string) (string, error)

	// Set the value for the given key.
	Put(key string, value string) error

	// Commit all Get(s) and Put(s) since Begin().
	Commit() bool

	// Abort all Get(s) and Put(s) since Begin().
	Abort()

	// Returns statistics (slice of integers) about most recent transaction.
	Stats() []int

	// Sharding logic: Given key, generates a number between 0 to nshards-1
	KeyToShard(key string, nshards uint64) uint64
}
