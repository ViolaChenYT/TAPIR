package IR

import (
	. "github.com/ViolaChenYT/TAPIR/common"
)

// Merge a list of potentially disagreeing reply to one reply
type ConsensusDecide func(results []*Response) *Response

func (c *Client) InvokeInconsistent(req *Request) error {
	return nil
}

func (c *Client) InvokeConsensus(req *Request, decide ConsensusDecide) (*Response, error) {
	return &Response{}, nil
}

func (c *Client) InvokeUnlogged(replicaIdx int, request Request) (*Response, error) {
	return &Response{}, nil
}
