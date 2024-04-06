package IR

import (
	"testing"
)

// test adding and initiating servers and replicas
func TestSetup(t *testing.T) {
	// Test code
	c1 := NewClient(1)
	err := c1.Start()
	if err != nil {
		t.Error("Error in starting client")
	}
	r1 := NewReplica(1)
	err = r1.Start()
	if err != nil {
		t.Error("Error in starting replica")
	}
}
