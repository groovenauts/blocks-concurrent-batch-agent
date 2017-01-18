package pipeline

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func init() {
	h := &handler{}

	g := e.Group("/pipelines")
	g.Use(middleware.CORS())

	g.GET("", withAEContext(h.index))
	g.GET("/:id", withPipeline(h.show))
	g.DELETE("/:id", withPipeline(h.destroy))

	g.POST("", withAEContext(h.create))
	g.POST("/:id/build_task", pipelineTask("build"))

	actions := []string{"close", "update", "resize"}
	for _, action := range actions {
		g.PUT("/:id/"+action, callPipelineTask(action))
		g.POST("/:id/"+action+"_task", pipelineTask(action))
	}

	g.GET("/refresh", withAEContext(h.refresh)) // from cron
	g.POST("/:id/refresh_task", pipelineTask("refresh"))
}

func withAEContext(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		req := c.Request()
		ctx := appengine.NewContext(req)
		c.Set("aecontext", ctx)
		return impl(c)
	}
}

func withPipeline(impl func(c echo.Context, pl *Pipeline) error) func(c echo.Context) error {
	return withAEContext(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param("id")
		pl, err := FindPipeline(ctx, id)
		switch {
		case err == ErrNoSuchPipeline:
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "@withPipeline %v id: %v\n", err, id)
			return err
		}
		return impl(c, pl)
	})
}

// curl -v -X PUT http://localhost:8080/pipelines/1/close
// curl -v -X PUT http://localhost:8080/pipelines/1/update
// curl -v -X PUT http://localhost:8080/pipelines/1/resize
func callPipelineTask(action string) func(c echo.Context) error {
	return withPipeline(func(c echo.Context, pl *Pipeline) error {
		id := c.Param("id")
		ctx := c.Get("aecontext").(context.Context)
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/%s_task", id, action), map[string][]string{})
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, pl)
	})
}

func pipelineTask(action string) func(c echo.Context) error {
	return withPipeline(func(c echo.Context, pl *Pipeline) error {
		ctx := c.Get("aecontext").(context.Context)
		err := pl.process(ctx, action)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, pl)
	})
}

type (
	handler struct{}
)

// curl -v -X POST http://localhost:8080/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *handler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	plp := &PipelineProps{}
	if err := c.Bind(plp); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return err
	}
	pl, err := CreatePipeline(ctx, plp)
	if err != nil {
		return err
	}
	t := taskqueue.NewPOSTTask("/pipelines/"+pl.ID+"/build_task", map[string][]string{})
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, pl)
}

// curl -v http://localhost:8080/pipelines
func (h *handler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pipelines, err := GetAllPipeline(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pipelines)
}

// curl -v http://localhost:8080/pipelines/1
func (h *handler) show(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X DELETE http://localhost:8080/pipelines/1
func (h *handler) destroy(c echo.Context, pl *Pipeline) error {
	ctx := c.Get("aecontext").(context.Context)
	if err := pl.destroy(ctx); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X PUT http://localhost:8080/pipelines/refresh
func (h *handler) refresh(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	ids, err := GetAllActivePipelineIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/%s/refresh_task", id), map[string][]string{})
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, ids)
}
