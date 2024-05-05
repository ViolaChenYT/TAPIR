import (
	"time"
)

type Timestamp struct {
	timestamp time.Time
	id        int
}

// NewTimestamp creates a new Timestamp instance
func NewTimestamp(client_id int) *Timestamp {
	return &Timestamp{
		timestamp:  time.Time{},
		id: client_id,
	}
}

func (t Timestamp) Equals(other Timestamp) bool {
	return t.timestamp == other.timestamp && t.id == other.id
}

func (t Timestamp) NotEquals(other Timestamp) bool {
	return t.timestamp != other.timestamp || t.id != other.id
}

func (t Timestamp) GreaterThan(other Timestamp) bool {
	if t.timestamp == other.timestamp {
		return t.id > other.id
	}
	return t.timestamp > other.timestamp
}

func (t Timestamp) LessThan(other Timestamp) bool {
	if t.timestamp == other.timestamp {
		return t.id < other.id
	}
	return t.timestamp < other.timestamp
}

// Helpers
func minTimestamp(timestamps []Timestamp) Timestamp {
	min := timestamps[0]
	for _, ts := range timestamps {
		if ts.LessThan(min) {
			min = ts
		}
	}
	return min
}

func maxTimestamp(timestamps []Timestamp) Timestamp {
	max := timestamps[0]
	for _, ts := range timestamps {
		if ts.GreaterThan(max) {
			max = ts
		}
	}
	return max
}