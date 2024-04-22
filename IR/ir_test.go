package IR

import (
	"testing"
)

// test adding and initiating servers and replicas
func TestSetup(t *testing.T) {
	// Test code
	c1, err := NewClient(1, "localhost", 8080)
	r1, err := NewReplica(1)
	if err != nil {
		t.Error("Error in starting replica")
	}
	c1.Close()
	r1.Close()
}
