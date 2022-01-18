package dura

type Operation struct {
	Snapshot OperationSnapshot `json:"snapshot"`
}

type OperationSnapshot struct {
	Repo    string         `json:"repo"`
	Op      *CaptureStatus `json:"op,omitempty"`
	Error   error          `json:"error,omitempty"`
	Latency float32        `json:"latency"`
}

func (o *Operation) ShouldLog() bool {
	return o.Snapshot.Op != nil || o.Snapshot.Error != nil
}
