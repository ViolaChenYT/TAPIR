package common

import "time"

type Timestamp struct {
	Timestamp time.Time
	ID        int
}

// NewTimestamp creates a new Timestamp instance
func NewTimestamp(clientID int) *Timestamp {
	return &Timestamp{
		Timestamp: time.Now(),
		ID:        clientID,
	}
}

func NewCustomTimestamp(clientID int, timestamp time.Time) *Timestamp {
	return &Timestamp{
		Timestamp: timestamp,
		ID:        clientID,
	}
}

func EmptyTime() *Timestamp {
	return &Timestamp{
		Timestamp: time.Time{},
		ID:        -1,
	}
}

func (t *Timestamp) Equals(other *Timestamp) bool {
	return t.Timestamp.Equal(other.Timestamp) && t.ID == other.ID
}

func (t *Timestamp) NotEquals(other *Timestamp) bool {
	return !t.Equals(other)
}

func (t *Timestamp) GreaterThan(other *Timestamp) bool {
	if t.Timestamp.Equal(other.Timestamp) {
		return t.ID > other.ID
	}
	return t.Timestamp.After(other.Timestamp)
}

func (t *Timestamp) LessThan(other *Timestamp) bool {
	return !t.GreaterThan(other) && !t.Equals(other)
}

func (t *Timestamp) LessThanOrEqualTo(other *Timestamp) bool {
	return t.Equals(other) || t.LessThan(other)
}

func LaterTime(t1 *Timestamp, t2 *Timestamp) *Timestamp {
	if t1.GreaterThan(t2) {
		return t1
	}
	return t2
}

// Helpers
func MinTimestamp(timestamps []*Timestamp) *Timestamp {
	if len(timestamps) == 0 {
		return nil
	}
	min := timestamps[0]
	for _, ts := range timestamps[1:] {
		if ts.LessThan(min) {
			min = ts
		}
	}
	return min
}

func MaxTimestamp(timestamps []*Timestamp) *Timestamp {
	if len(timestamps) == 0 {
		return nil
	}
	max := timestamps[0]
	for _, ts := range timestamps[1:] {
		if ts.GreaterThan(max) {
			max = ts
		}
	}
	return max
}
