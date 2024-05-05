package tapir

// import "time"

// TapirClient represents a client for interacting with the Tapir protocol
type TapirClient interface {

	// Begin a transaction
	Begin()

	// Read the value corresponding to key.
	Read(key string) (string, error)

	// Set the value for the given key.
	Write(key string, value string) error

	// Commit all Get(s) and Put(s) since Begin().
	Commit() bool

	// Abort all Get(s) and Put(s) since Begin().
	Abort()
}
