package common

import (
	"fmt"
)

// Transaction represents a transaction with read and write sets
type Transaction struct {
	ID       int
	ReadSet  map[string]string
	ReadTime map[string]*Timestamp
	WriteSet map[string]string
}

// NewTransaction creates a new Transaction instance
func NewTransaction(id int) *Transaction {
	return &Transaction{
		ID:       id,
		ReadSet:  make(map[string]string), // value and read time from store
		ReadTime: make(map[string]*Timestamp),
		WriteSet: make(map[string]string),
	}
}

// AddReadSet adds an entry to the read set of the transaction
func (t *Transaction) AddReadSet(key string, value string, readTime *Timestamp) {
	t.ReadSet[key] = value
	t.ReadTime[key] = readTime
}

// AddWriteSet adds an entry to the write set of the transaction
func (t *Transaction) AddWriteSet(key, value string) {
	t.WriteSet[key] = value
}

func (t Transaction) String() string {
	readSetStr := "{"
	for key, value := range t.ReadSet {
		readSetStr += fmt.Sprintf("%s: (%s, %v), ", key, value, t.ReadTime[key])
	}
	readSetStr += "}"

	writeSetStr := "{"
	for key, value := range t.WriteSet {
		writeSetStr += fmt.Sprintf("%s: %s, ", key, value)
	}
	writeSetStr += "}"

	return fmt.Sprintf("Transaction ID: %d\nRead Set: %s\n Write Set: %s\n", t.ID, readSetStr, writeSetStr)
}
