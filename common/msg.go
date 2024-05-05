// DO NOT MODIFY THIS FILE!

package common

import (
	"fmt"
)

type OpType int

const (
	OP_GET OpType = iota
	OP_PREPARE
	OP_COMMIT
	OP_ABORT
)

type MsgType int

const (
	MsgPropose MsgType = iota // Sent by clients to make a connection w/ the server.
	MsgReply
	MsgFinalize
	MsgConfirm
)

type PrepareState int

const (
	PREPARE_OK PrepareState = iota
	ABSTAIN
	ABORT
	RETRY
)

// type Operation struct {
// 	op_type   OperationType
// 	key       string
// 	value     string // optional
// 	timestamp time.Time
// }

// type Result struct {
// 	op_type string
// 	key     string
// 	value   string
// 	State   PrepareState
// }

// Message represents a message used by the LSP protocol.
type Message struct {
	Type        MsgType // One of the message types listed above.
	ConnID      int     // Unique client-server connection ID.
	OperationID int     // operation ID
	Response    *Response
	Request     *Request
}

func NewPropose(opID int, op *Request) Message {
	return Message{
		Type:        MsgPropose,
		OperationID: opID,
		Request:     op,
	}
}

func NewReply(opID int, res *Response) Message {
	return Message{
		Type:        MsgReply,
		OperationID: opID,
		Response:    res,
	}
}

func NewFinalize(opID int) Message {
	return Message{
		Type:        MsgFinalize,
		OperationID: opID,
	}
}
func Finalize(opID int, res *Response) Message {
	return Message{
		Type:        MsgFinalize,
		OperationID: opID,
		Response:    res,
	}
}

func NewConfirm(opID int) *Message {
	return &Message{
		Type:        MsgConfirm,
		OperationID: opID,
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

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
