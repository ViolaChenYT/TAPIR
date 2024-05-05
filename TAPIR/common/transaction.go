package tapir

// Transaction represents a transaction with read and write sets
type Transaction struct {
	id       int
	readSet  map[string](string, Timestamp)
	writeSet map[string]string
}

// NewTransaction creates a new Transaction instance
func NewTransaction(int id) *Transaction {
	return &Transaction{
		id:       id,
		readSet:  make(map[string](string, Timestamp)), // value and read time from store
		writeSet: make(map[string]string),
	}
}

// GetReadSet returns the read set of the transaction
func (t *Transaction) GetReadSet() map[string]Timestamp {
	return t.readSet
}

// GetWriteSet returns the write set of the transaction
func (t *Transaction) GetWriteSet() map[string]string {
	return t.writeSet
}

// AddReadSet adds an entry to the read set of the transaction
func (t *Transaction) AddReadSet(key string, value string, readTime Timestamp) {
	t.readSet[key] = (value, readTime)
}

// AddWriteSet adds an entry to the write set of the transaction
func (t *Transaction) AddWriteSet(key, value string) {
	t.writeSet[key] = value
}