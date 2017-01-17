package pipeline

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/taskqueue"
)

type (
	apiHandler struct {}
	taskHandler struct {
		factory ProcessorFactory
	}
)

func withAEContext(impl func (c echo.Context) error) (func(c echo.Context) error) {
	return func(c echo.Context) error {
		req := c.Request()
		ctx := appengine.NewContext(req)
		c.Set("aecontext", ctx)
		return impl(c)
	}
}

func withPipeline(impl func (c echo.Context, pl *Pipeline) error) (func(c echo.Context) error) {
	return withAEContext(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param("id")
		pl, err := FindPipeline(ctx, id)
		if err != nil {
			return err
		}
		return impl(c, pl)
	})
}

// curl -v -X PUT http://localhost:8080/pipelines/1/close
// curl -v -X PUT http://localhost:8080/pipelines/1/update
// curl -v -X PUT http://localhost:8080/pipelines/1/resize
func (ah *apiHandler)callPipelineTask(action string) (func(c echo.Context) error) {
	return withPipeline(func(c echo.Context, pl *Pipeline) error {
		id := c.Param("id")
		ctx := c.Get("aecontext").(context.Context)
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/%s/%s_task", id, action), map[string][]string{})
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, pl)
	})
}

func (th *taskHandler)pipelineTask(action string) (func(c echo.Context) error) {
	return withPipeline(func(c echo.Context, pl *Pipeline) error {
		ctx := c.Get("aecontext").(context.Context)
		pr, err := th.factory.Create(ctx, action)
		if err != nil {
			return err
		}
		return pr.Process(ctx, pl)
	})
}

func init() {
	ah := &apiHandler{}
	th := &taskHandler{&DefaultProcessorFactory{}}

	g := e.Group("/pipelines")
	g.Use(middleware.CORS())

	g.GET("", withAEContext(ah.index))
	g.GET("/:id", withPipeline(ah.show))
	g.DELETE("/:id", withPipeline(ah.destroy))

	g.POST(""               , withAEContext(ah.create))
	g.POST("/:id/build_task", th.pipelineTask("build"))

	actions := []string{"close", "update", "resize"}
	for _, action := range actions {
		g.PUT( "/:id/" + action          , ah.callPipelineTask(action))
		g.POST("/:id/" + action + "_task", th.pipelineTask(action))
	}

	g.GET( "/refresh"         , withAEContext(ah.refresh)) // from cron
	g.POST("/:id/refresh_task", th.pipelineTask("refresh"))
}

// curl -v -X POST http://localhost:8080/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *apiHandler) create(c echo.Context) error {
	plp := &PipelineProps{}
	if err := c.Bind(plp); err != nil {
		return err
	}
	ctx := c.Get("aecontext").(context.Context)
	id, err := CreatePipeline(ctx, plp)
	if err != nil {
		return err
	}
	t := taskqueue.NewPOSTTask("/" + id + "/build_task", map[string][]string{})
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, plp)
}

// curl -v http://localhost:8080/pipelines
func (h *apiHandler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pipelines, err := GetAllPipeline(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pipelines)
}

// curl -v http://localhost:8080/pipelines/1
func (h *apiHandler) show(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X DELETE http://localhost:8080/pipelines/1
func (h *apiHandler) destroy(c echo.Context, pl *Pipeline) error {
	ctx := c.Get("aecontext").(context.Context)
	if err := pl.destroy(ctx); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X PUT http://localhost:8080/pipelines/refresh
func (h *apiHandler) refresh(c echo.Context) error {
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
