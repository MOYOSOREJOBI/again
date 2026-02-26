package replay

import "sync"

var (
	mu   sync.Mutex
	jobs = map[string]*Job{}
)

func Put(job *Job) {
	mu.Lock()
	defer mu.Unlock()
	jobs[job.ID] = job
}

func Get(id string) *Job {
	mu.Lock()
	defer mu.Unlock()
	return jobs[id]
}
