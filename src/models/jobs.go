package models

type Jobs []*Job

func (jobs Jobs) AllFinished() bool {
	if len(jobs) == 0 {
		return false
	}
	return jobs.All(func(job *Job) bool { return !job.Status.Living() })
}

func (jobs Jobs) All(f func(*Job) bool) bool {
	for _, job := range jobs {
		if !f(job) {
			return false
		}
	}
	return true
}

func (jobs Jobs) Finished() Jobs {
	result := Jobs{}
	for _, job := range jobs {
		if job.Status.Finished() {
			result = append(result, job)
		}
	}
	return result
}

func (jobs Jobs) IDs() []string {
	jobIDs := []string{}
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}
	return jobIDs
}
