package librpc

import (
	"github.com/ViolaChenYT/TAPIR/common/storagerpc"
)

type RemoteLeaseCallbacks interface {
	RevokeLease(*storagerpc.RevokeLeaseArgs, *storagerpc.RevokeLeaseReply) error
}

type LeaseCallbacks struct {
	// Embed all methods into the struct. See the Effective Go section about
	// embedding for more details: golang.org/doc/effective_go.html#embedding
	RemoteLeaseCallbacks
}

// Wrap wraps l in a type-safe wrapper struct to ensure that only the desired
// LeaseCallbacks methods are exported to receive RPCs.
func Wrap(l RemoteLeaseCallbacks) RemoteLeaseCallbacks {
	return &LeaseCallbacks{l}
}
