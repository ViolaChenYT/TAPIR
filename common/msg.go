// DO NOT MODIFY THIS FILE!

package common

type OpType int

const (
	OP_GET OpType = iota
	OP_PREPARE
	OP_COMMIT
	OP_ABORT
)

// GetMessage represents the GetMessage message
type GetMessage struct {
	Key       string
	Timestamp *Timestamp
}

// PrepareMessage represents the PrepareMessage message
type PrepareMessage struct {
	Txn       *Transaction
	Timestamp *Timestamp
}

// CommitMessage represents the CommitMessage message
type CommitMessage struct {
	Timestamp *Timestamp
}

// Request represents the Request message
type Request struct {
	Op      OpType
	TxnID   int
	Get     *GetMessage
	Prepare *PrepareMessage
	Commit  *CommitMessage
}

type ReplyType int

const (
	RPLY_OK ReplyType = iota
	RPLY_ABORT
	RPLY_RETRY
	RPLY_ABSTAIN
)

type Response struct {
	Status    ReplyType
	Value     string
	Timestamp *Timestamp
}

func NewResponse(status ReplyType) *Response {
	return &Response{
		Status: status,
	}
}

func NewResponseWithTime(status ReplyType, timestamp *Timestamp) *Response {
	return &Response{
		Status:    status,
		Timestamp: timestamp,
	}
}

func NewReadResponse(value string, timestamp *Timestamp) *Response {
	return &Response{
		Status:    RPLY_OK,
		Timestamp: timestamp,
		Value:     value,
	}
}
