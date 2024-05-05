package tapir

// TapirApp is a table database storing records
type TapirApp interface {
	// Read reads a record from the database and returns a map of each field/value pair.
	Read(table string, key string, fields []string) (map[string][]byte, error)

	// Update updates a record in the database.
	Update(table string, key string, values map[string][]byte) error

	// Insert inserts a record into the database.
	Insert(table string, key string, values map[string][]byte) error

	// Delete deletes a record from the database.
	Delete(table string, key string) error

	// Start starts a transaction.
	Start() error

	// Commit commits a transaction.
	Commit() error

	// Abort aborts a transaction.
	Abort() error
}
