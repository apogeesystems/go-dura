package dura

type Operation struct {
	Snapshot OperationSnapshot
}

type OperationSnapshot struct {
	repo    string
	op      *CaptureStatus
	error   error
	latency float32
}

func (o *Operation) ShouldLog() bool {
	return o.Snapshot.op != nil || o.Snapshot.error != nil
}
