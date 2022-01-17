package dura

type Operation struct {
	Snapshot OperationSnapshot `json:"snapshot"`
}

type OperationSnapshot struct {
	repo    string         `json:"repo"`
	op      *CaptureStatus `json:"op,omitempty"`
	error   error          `json:"error,omitempty"`
	latency float32        `json:"latency"`
}

func (o *Operation) ShouldLog() bool {
	return o.Snapshot.op != nil || o.Snapshot.error != nil
}
