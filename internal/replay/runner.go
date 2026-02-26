package replay

func Run(jobID string, recompute func(string) error) {
	markJobRunning(jobID)
	err := recompute(jobID)
	if err != nil {
		markJobFailed(jobID, err.Error())
		return
	}
	markJobCompleted(jobID)
}

func markJobRunning(jobID string) {
	if j := Get(jobID); j != nil {
		j.Status = "running"
	}
}
func markJobFailed(jobID, _ string) {
	if j := Get(jobID); j != nil {
		j.Status = "failed"
	}
}
func markJobCompleted(jobID string) {
	if j := Get(jobID); j != nil {
		j.Status = "completed"
	}
}
