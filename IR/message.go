// DO NOT MODIFY THIS FILE!

package IR

import (
	"fmt"
	"strconv"
)

// MsgType is an integer code describing an LSP message type.
type MsgType int

const (
	MsgConnect MsgType = iota // Sent by clients to make a connection w/ the server.
	MsgPropose
	MsgReply
	MsgFinalize
	MsgConfirm
)

// Message represents a message used by the LSP protocol.
type Message struct {
	Type        MsgType // One of the message types listed above.
	ConnID      int     // Unique client-server connection ID.
	OperationID int     // operation ID
	Size        int     // Size of the payload.
	Checksum    uint16  // Message checksum.
	Payload     []byte  // Data message payload.
	Result      *result
	Op          *operation
}

// NewConnect returns a new connect message.
func NewConnect(initialSeqNum int) *Message {
	return &Message{
		Type: MsgConnect,
	}
}

func NewPropose(opID int, op *operation) *Message {
	return &Message{
		Type:        MsgPropose,
		OperationID: opID,
		Op:          op,
	}
}

func NewReply(opID int, res *result) *Message {
	return &Message{
		Type:        MsgReply,
		OperationID: opID,
		Result:      res,
	}
}

func NewFinalize(opID int, res *result) *Message {
	return &Message{
		Type:        MsgFinalize,
		OperationID: opID,
		Result:      res,
	}
}

func NewConfirm(opID int) *Message {
	return &Message{
		Type:        MsgConfirm,
		OperationID: opID,
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

// NewData returns a new data message with the specified connection ID,
// sequence number, and payload.
// func NewData(connID, seqNum, size int, payload []byte, checksum uint16) *Message {
// 	return &Message{
// 		Type:     MsgData,
// 		ConnID:   connID,
// 		SeqNum:   seqNum,
// 		Size:     size,
// 		Payload:  payload,
// 		Checksum: checksum,
// 	}
// }

// // NewAck returns a new acknowledgement message with the specified
// // connection ID and sequence number.
// func NewAck(connID, seqNum int) *Message {
// 	return &Message{
// 		Type:   MsgAck,
// 		ConnID: connID,
// 		SeqNum: seqNum,
// 	}
// }

// // NewCAck returns a new cumulative acknowledgement message with
// // the specified connection ID and sequence number.
// func NewCAck(connID, seqNum int) *Message {
// 	return &Message{
// 		Type:   MsgCAck,
// 		ConnID: connID,
// 		SeqNum: seqNum,
// 	}
// }

// String returns a string representation of this message. To pretty-print a
// message, you can pass it to a format string like so:
//
//	msg := NewConnect()
//	fmt.Printf("Connect message: %s\n", msg)
func (m *Message) String() string {
	var name, payload, checksum string
	switch m.Type {
	case MsgConnect:
		name = "Connect"
	case MsgPropose:
		name = "Propose"
		checksum = " " + strconv.Itoa(int(m.Checksum))
		payload = " " + string(m.Payload)
	case MsgReply:
		name = "Reply"
	case MsgFinalize:
		name = "Finalize"
	case MsgConfirm:
		name = "Confirm"
	}
	return fmt.Sprintf("[%s %d %d%s%s]", name, m.ConnID, m.OperationID, checksum, payload)
}