package api

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/groovenauts/blocks-concurrent-batch-agent/gae_support"
	"github.com/groovenauts/blocks-concurrent-batch-agent/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

type handler struct{}

const (
	AUTH_HEADER = "Authorization"
)

func init() {
	h := &handler{}

	g := e.Group("/pipelines")
	g.Use(middleware.CORS())

	g.GET("", h.withAuth(h.index))
	g.GET("/subscriptions", h.withAuth(h.subscriptions))
	g.GET("/:id", h.withPipeline(h.withAuth, h.show))
	g.DELETE("/:id", h.withPipeline(h.withAuth, h.destroy))

	g.POST("", h.withAuth(h.create))
	g.POST("/:id/build_task", h.pipelineTask("build"))

	g.PUT("/:id/close", h.callPipelineTask("close"))
	g.POST("/:id/close_task", h.pipelineTask("close"))

	g.GET("/refresh", gae_support.With(h.refresh)) // Don't use withAuth because this is called from cron
	g.POST("/:id/refresh_task", h.pipelineTask("refresh"))
}

func (h *handler) withAuth(impl func(c echo.Context) error) func(c echo.Context) error {
	return gae_support.With(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		req := c.Request()
		raw := req.Header.Get(AUTH_HEADER)
		if raw == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		}
		re := regexp.MustCompile(`\ABearer\s+`)
		token := re.ReplaceAllString(raw, "")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		}
		_, err := models.FindAuthWithToken(ctx, token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
		}
		return impl(c)
	})
}

func (h *handler) withPipeline(wrapper func(func(echo.Context) error) func(echo.Context) error,
	impl func(c echo.Context, pl *models.Pipeline) error) func(c echo.Context) error {
	return wrapper(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param("id")
		pl, err := models.FindPipeline(ctx, id)
		switch {
		case err == models.ErrNoSuchPipeline:
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "@withPipeline %v id: %v\n", err, id)
			return err
		}
		return impl(c, pl)
	})
}

// curl -v -X PUT http://localhost:8080/pipelines/1/close
func (h *handler) callPipelineTask(action string) func(c echo.Context) error {
	return h.withPipeline(h.withAuth, func(c echo.Context, pl *models.Pipeline) error {
		id := c.Param("id")
		ctx := c.Get("aecontext").(context.Context)
		req := c.Request()
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/%s_task", id, action), map[string][]string{})
		t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, pl)
	})
}

// curl -v -X POST http://localhost:8080/pipelines/1/build_task
// curl -v -X	POST http://localhost:8080/pipelines/1/close_task
// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task
func (h *handler) pipelineTask(action string) func(c echo.Context) error {
	var wrapper func(impl func(c echo.Context) error) func(c echo.Context) error
	switch action {
	case "refresh":
		wrapper = gae_support.With
	default:
		wrapper = h.withAuth
	}
	return h.withPipeline(wrapper, func(c echo.Context, pl *models.Pipeline) error {
		ctx := c.Get("aecontext").(context.Context)
		err := pl.Process(ctx, action)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, pl)
	})
}

// curl -v -X POST http://localhost:8080/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *handler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	pl := &models.Pipeline{}
	if err := c.Bind(pl); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return err
	}
	err := models.CreatePipeline(ctx, pl)
	if err != nil {
		log.Errorf(ctx, "Failed to create pipeline: %v\n%v\n", pl, err)
		return err
	}
	log.Debugf(ctx, "Created pipeline: %v\n", pl)
	if !pl.Dryrun {
		t := taskqueue.NewPOSTTask("/pipelines/"+pl.ID+"/build_task", map[string][]string{})
		t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusCreated, pl)
}

// curl -v http://localhost:8080/pipelines
func (h *handler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pipelines, err := models.GetAllPipelines(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pipelines)
}

// curl -v http://localhost:8080/pipelines/subscriptions
func (h *handler) subscriptions(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	subscriptions, err := models.GetActiveSubscriptions(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, subscriptions)
}

// curl -v http://localhost:8080/pipelines/1
func (h *handler) show(c echo.Context, pl *models.Pipeline) error {
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X DELETE http://localhost:8080/pipelines/1
func (h *handler) destroy(c echo.Context, pl *models.Pipeline) error {
	ctx := c.Get("aecontext").(context.Context)
	if err := pl.Destroy(ctx); err != nil {
		switch err.(type) {
		case *models.InvalidOperation:
			res := map[string]interface{}{"message": err.Error()}
			c.JSON(http.StatusNotAcceptable, res)
		default:
			return err
		}
	}
	return c.JSON(http.StatusOK, pl)
}

// This is called from cron
// curl -v -X PUT http://localhost:8080/pipelines/refresh
func (h *handler) refresh(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	statuses := map[string]models.Status{"deploying": models.Deploying, "closing": models.Closing}
	res := map[string][]string{}
	for name, st := range statuses {
		ids, err := models.GetPipelineIDsByStatus(ctx, st)
		if err != nil {
			return err
		}
		for _, id := range ids {
			t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/refresh_task", id), map[string][]string{})
			if _, err := taskqueue.Add(ctx, t, ""); err != nil {
				return err
			}
		}
		res[name] = ids
	}
	return c.JSON(http.StatusOK, res)
}
