package analysis

// RunHandle links an analysis run to its rate-limit attempt record.
type RunHandle struct {
	RunID     int64
	AttemptID int64
}

const (
	attemptStatusStarted   = "started"
	attemptStatusCompleted = "completed"
	attemptStatusFailed    = "failed"
)
