// The possible results returned by Prepare
package tapir

type PrepareState int

const (
	PREPARE_OK PrepareState = iota
	ABSTAIN
	ABORT
	RETRY
)

// Prepare results with possible timestamp
type PrepareResult struct {
	result PrepareState
	time   *Timestamp
}

// NewPrepareResult constructs a new PrepareResult instance.
func NewPrepareResult(result PrepareState) PrepareResult {
	return PrepareResult{
		result: result,
		time:   nil,
	}
}

func CreateRetry(timestamp Timestamp) PrepareResult {
	return PrepareResult{
		result: RETRY,
		time:   &timestamp,
	}
}
