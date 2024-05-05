package tapir

// Transaction represents a transaction with read and write sets
type Transaction struct {
	id       int
	readSet  map[string]string
	readTime map[string]Timestamp
	writeSet map[string]string
}

// NewTransaction creates a new Transaction instance
func NewTransaction(id int) *Transaction {
	return &Transaction{
		id:       id,
		readSet:  make(map[string]string), // value and read time from store
		readTime: make(map[string]Timestamp),
		writeSet: make(map[string]string),
	}
}

// GetReadSet returns the read set of the transaction
func (t *Transaction) GetReadSet() (map[string]string, map[string]Timestamp) {
	return t.readSet, t.readTime
}

// GetWriteSet returns the write set of the transaction
func (t *Transaction) GetWriteSet() map[string]string {
	return t.writeSet
}

// AddReadSet adds an entry to the read set of the transaction
func (t *Transaction) AddReadSet(key string, value string, readTime Timestamp) {
	t.readSet[key] = value
	t.readTime[key] = readTime
}

// AddWriteSet adds an entry to the write set of the transaction
func (t *Transaction) AddWriteSet(key, value string) {
	t.writeSet[key] = value
}
