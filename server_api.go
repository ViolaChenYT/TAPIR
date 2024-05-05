package tapir

import (
	"github.com/ViolaChenYT/TAPIR/IR"
)

// need to define a lot of types

// return value of get should be <timestamp, value>

type Server interface {
	Get(*IR.Operation, *IR.Result)
	Prepare(*IR.Operation, *IR.Result)
	Commit(*IR.Operation, *IR.Result)
	Abort(*IR.Operation, *IR.Result)
	Put(*IR.Operation, *IR.Result)
	GetPreparedWrites(*IR.Operation, *IR.Result)
	GetPreparedReads(*IR.Operation, *IR.Result)
}
