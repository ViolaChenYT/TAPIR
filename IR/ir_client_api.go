package IR

func (c *Client) InvokeInconsistent(Operation) error {
	return nil
}

func (c *Client) InvokeConsensus(Operation, []Result) (Result, error) {
	return Result{}, nil
}
