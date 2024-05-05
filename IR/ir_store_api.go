package IR

// replica upcalls (like, a call to user code)
func (r *Replica) ExecInconsistent(op Operation) error {
	return nil
}

func (r *Replica) ExecConsensus(op Operation) (Result, error) {
	return Result{}, nil
}

func (r *Replica) Sync() error {
	return nil
}

func (r *Replica) Merge(d []Operation, u []Operation) (Record, error) {
	return Record{}, nil
}
