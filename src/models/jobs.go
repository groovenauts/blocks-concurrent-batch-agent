package models

type Jobs []*Job

func (jobs Jobs) AllFinished() bool {
	if len(jobs) == 0 {
		return false
	}

	for _, job := range jobs {
		if job.Status.Living() {
			return false
		}
	}

	return true
}
