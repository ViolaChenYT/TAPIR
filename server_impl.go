package tapir

import "github.com/ViolaChenYT/TAPIR/IR"

type server struct {
}

// server is a storage server consisting of multiple replicas
func (s *server) Get(*IR.Operation, *IR.Result) {
	// Test code
}

func (s *server) Prepare(*IR.Operation, *IR.Result) {
	// Test code
}

func (s *server) Commit(*IR.Operation, *IR.Result) {
	// Test code
}

func (s *server) Abort(*IR.Operation, *IR.Result) {
	// Test code
}

func (s *server) Put(*IR.Operation, *IR.Result) {
	// Test code
}

func (s *server) GetPreparedWrites(*IR.Operation, *IR.Result) {
	// Test code
}

func (s *server) GetPreparedReads(*IR.Operation, *IR.Result) {
	// Test code
}
