package api

import (
	"fmt"
	"net/http"
	"time"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

type JobHandler struct {
	pipeline_id_name string
	job_id_name      string
}

func (h *JobHandler) collection(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(plBy(h.pipeline_id_name, http.StatusNotFound, PlToOrg(withAuth(action))))
}

func (h *JobHandler) member(action echo.HandlerFunc) echo.HandlerFunc {
	return gae_support.With(jobBy(h.job_id_name, http.StatusNotFound, JobToPl(PlToOrg(withAuth(action)))))
}

// curl -v -X POST http://localhost:8080/pipelines/3/jobs --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *JobHandler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)

	job := &models.Job{}
	if err := c.Bind(job); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", c.Request())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	job.Pipeline = pl
	job.InitStatus(c.QueryParam("ready") == "true")

	switch pl.Status {
	case models.HibernationChecking:
		pl.PullingTaskSize = 1
		err := pl.BackToBeOpened(ctx)
		if err != nil {
			return err
		}
		err = PostPipelineTask(c, "subscribe_task", pl)
		if err != nil {
			return err
		}
	case models.Hibernating:
		err := pl.BackToBeReserved(ctx)
		if err != nil {
			return err
		}
		err = PostPipelineTask(c, "build_task", pl)
		if err != nil {
			return err
		}
	}
	err := job.CreateAndPublishIfPossible(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create Job: %v\n%v\n", job, err)
		return err
	}
	log.Debugf(ctx, "Created Job: %v\n", job)

	err = h.StartToWaitAndPublishIfNeeded(c, job)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, job)
}

// curl -v http://localhost:8080/pipelines/3/jobs
func (h *JobHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
	jobs, err := pl.JobAccessor().All(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jobs)
}

type BulkGetJobsPayload struct {
	JobIds []string `json:"job_ids"`
}
type BulkGetJobsMediaTypeJob struct {
	ID          string           `json:"id"`
	IdByClient  string           `json:"id_by_client"`
	Status      models.JobStatus `json:"status"`
	PublishedAt time.Time        `json:"published_at,omitempty"`
	StartTime   string           `json:"start_time"`
	FinishTime  string           `json:"finish_time"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
type BulkGetJobsMediaType struct {
	Jobs   map[string]*BulkGetJobsMediaTypeJob `json:"jobs"`
	Errors map[string]error                    `json:"errors"`
}

// curl -v -X POST http://localhost:8080/pipelines/3/bulk_get_jobs --data '{"job_ids":["1","2","5"]}' -H 'Content-Type: application/json'
func (h *JobHandler) BulkGetJobs(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	payload := &BulkGetJobsPayload{}
	if err := c.Bind(payload); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", c.Request())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pl := c.Get("pipeline").(*models.Pipeline)
	jobs, errors := pl.JobAccessor().BulkGet(ctx, payload.JobIds)
	r := &BulkGetJobsMediaType{
		Jobs:   h.JobsFromModelToMediaType(jobs),
		Errors: errors,
	}
	return c.JSON(http.StatusOK, r)
}

func (h *JobHandler) JobsFromModelToMediaType(models map[string]*models.Job) map[string]*BulkGetJobsMediaTypeJob {
	r := map[string]*BulkGetJobsMediaTypeJob{}
	for key, model := range models {
		r[key] = &BulkGetJobsMediaTypeJob{
			ID:          model.ID,
			IdByClient:  model.IdByClient,
			Status:      model.Status,
			PublishedAt: model.PublishedAt,
			StartTime:   model.StartTime,
			FinishTime:  model.FinishTime,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
		}
	}
	return r
}

type BulkJobStatusesMediaType struct {
	Jobs   map[string]int   `json:"jobs"`
	Errors map[string]error `json:"errors"`
}

// curl -v -X POST http://localhost:8080/pipelines/3/bulk_job_statuses --data '{"job_ids":["1","2","5"]}' -H 'Content-Type: application/json'
func (h *JobHandler) BulkJobStatuses(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	payload := &BulkGetJobsPayload{}
	if err := c.Bind(payload); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", c.Request())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	pl := c.Get("pipeline").(*models.Pipeline)
	acc := pl.JobAccessor()
	errors := map[string]error{}
	jobStatusMap := map[string]int{}
	for _, jobId := range payload.JobIds {
		jobs, err := acc.AllWith(ctx, func(q *datastore.Query) (*datastore.Query, error) {
			q = q.Filter("id_by_client =", jobId)
			return q, nil
		})
		if err != nil {
			errors[jobId] = err
		}
		for _, job := range jobs {
			jobStatusMap[job.IdByClient] = int(job.Status)
		}
	}
	r := &BulkJobStatusesMediaType{
		Jobs:   jobStatusMap,
		Errors: errors,
	}
	return c.JSON(http.StatusOK, r)
}

// curl -v http://localhost:8080/jobs/1
func (h *JobHandler) show(c echo.Context) error {
	job := c.Get("job").(*models.Job)
	return c.JSON(http.StatusOK, job)
}

// curl -v http://localhost:8080/jobs/1/getready
func (h *JobHandler) getReady(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	job := c.Get("job").(*models.Job)
	pl := c.Get("pipeline").(*models.Pipeline)
	job.Pipeline = pl
	job.Status = models.Ready
	err := job.UpdateAndPublishIfPossible(ctx) // This method might change the status to Publishing
	if err != nil {
		log.Errorf(ctx, "Failed to update Job: %v\n%v\n", job, err)
		return err
	}
	log.Debugf(ctx, "Updated Job: %v\n", pl)

	err = h.StartToWaitAndPublishIfNeeded(c, job)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, job)
}

func (h *JobHandler) StartToWaitAndPublishIfNeeded(c echo.Context, job *models.Job) error {
	if job.Status != models.Ready {
		return nil
	}
	return h.PostJobTask(c, job, "wait_task", time.Now())
}

func (h *JobHandler) PostJobTask(c echo.Context, job *models.Job, action string, eta time.Time) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/jobs/%s/%v", job.ID, action), map[string][]string{})
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	t.ETA = eta
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return nil
}

// curl -v http://localhost:8080/jobs/1/wait_task
func (h *JobHandler) WaitToPublishTask(c echo.Context) error {
	started := time.Now()
	pl := c.Get("pipeline").(*models.Pipeline)
	job := c.Get("job").(*models.Job)
	switch {
	case models.StatusesOpened.Include(pl.Status):
		err := h.PostJobTask(c, job, "publish_task", started)
		if err != nil {
			return nil
		}
		return c.JSON(http.StatusOK, job)
	default:
		err := h.PostJobTask(c, job, "wait_task", started.Add(30*time.Second))
		if err != nil {
			return nil
		}
		return c.JSON(http.StatusNoContent, job)
	}
}

// curl -v http://localhost:8080/jobs/1/publish_task
func (h *JobHandler) PublishTask(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	job := c.Get("job").(*models.Job)
	if job.Status != models.Publishing {
		job.Status = models.Publishing
		log.Debugf(ctx, "PublishAndUpdate#1: %v\n", job)
		err := job.Update(ctx)
		if err != nil {
			return err
		}
	}
	err := job.PublishAndUpdateWithTx(ctx)
	if err != nil {
		return err
	}

	h.IncreaseSubscribeTask(c, ctx, c.Get("pipeline").(*models.Pipeline))

	return c.JSON(http.StatusOK, job)
}

func (h *JobHandler) IncreaseSubscribeTask(c echo.Context, ctx context.Context, pl *models.Pipeline) {
	jobCount, err := pl.JobCount(ctx, models.Publishing, models.Published, models.Executing)
	if err != nil {
		log.Warningf(ctx, "Failed to get JobCount(publishing, published, executing) because of %v\n", err)
		return
	}

	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		return pl.CalcAndUpdatePullingTaskSize(ctx, jobCount, func(newTasks int) error {
			for i := 0; i < newTasks; i++ {
				if err := PostPipelineTask(c, "subscribe_task", pl); err != nil {
					log.Warningf(ctx, "Failed to start subscribe_task for %v because of %v\n", pl.ID, err)
					return err
				}
			}
			return nil
		})
	}, nil)
	if err != nil {
		log.Warningf(ctx, "Failed to CalcAndUpdatePullingTaskSize for %v because of %v\n", pl.ID, err)
	}

	return
}

// curl -v http://localhost:8080/jobs/1/cancel
func (h *JobHandler) Cancel(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	job := c.Get("job").(*models.Job)
	err := job.Cancel(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, job)
}
