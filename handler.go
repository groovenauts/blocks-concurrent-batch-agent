package pipeline

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func withAEContext(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		req := c.Request()
		ctx := appengine.NewContext(req)
		c.Set("aecontext", ctx)
		return impl(c)
	}
}

type handler struct{}

func init() {
	h := &handler{}

	g := e.Group("/pipelines")
	g.Use(middleware.CORS())

	g.GET(".json", h.withAuth(h.index))
	g.GET("/:id.json", h.withPipeline(h.show))
	g.DELETE("/:id.json", h.withPipeline(h.destroy))

	g.POST(".json", h.withAuth(h.create))
	g.POST("/:id/build_task.json", h.pipelineTask("build"))

	actions := []string{"close", "update", "resize"}
	for _, action := range actions {
		g.PUT("/:id/"+action+".json", h.callPipelineTask(action))
		g.POST("/:id/"+action+"_task.json", h.pipelineTask(action))
	}

	g.GET("/refresh.json", h.withAuth(h.refresh)) // from cron
	g.POST("/:id/refresh_task.json", h.pipelineTask("refresh"))
}

func (h *handler) withAuth(impl func(c echo.Context) error) func(c echo.Context) error {
	return withAEContext(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		req := c.Request()
		raw := req.Header.Get("Authorization")
		if raw == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		}
		re := regexp.MustCompile(`\ABearer\s+`)
		token := re.ReplaceAllString(raw, "")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		}
		_, err := FindAuthWithToken(ctx, token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
		}
		return impl(c)
	})
}

func (h *handler) withPipeline(impl func(c echo.Context, pl *Pipeline) error) func(c echo.Context) error {
	return h.withAuth(func(c echo.Context) error {
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

// curl -v -X PUT http://localhost:8080/pipelines/1/close.json
// curl -v -X PUT http://localhost:8080/pipelines/1/update.json
// curl -v -X PUT http://localhost:8080/pipelines/1/resize.json
func (h *handler) callPipelineTask(action string) func(c echo.Context) error {
	return h.withPipeline(func(c echo.Context, pl *Pipeline) error {
		id := c.Param("id")
		ctx := c.Get("aecontext").(context.Context)
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/pipelines/%s/%s_task.json", id, action), map[string][]string{})
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, pl)
	})
}

// curl -v -X POST http://localhost:8080/pipelines/1/build_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/close_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/update_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/resize_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task.json
func (h *handler) pipelineTask(action string) func(c echo.Context) error {
	return h.withPipeline(func(c echo.Context, pl *Pipeline) error {
		ctx := c.Get("aecontext").(context.Context)
		err := pl.process(ctx, action)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, pl)
	})
}

// curl -v -X POST http://localhost:8080/pipelines.json --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *handler) create(c echo.Context) error {
	pl, err := h.createImpl(c)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, pl)
}

func (h *handler) createImpl(c echo.Context) (*Pipeline, error) {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	plp := &PipelineProps{}
	if err := c.Bind(plp); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return nil, err
	}
	pl, err := CreatePipeline(ctx, plp)
	log.Debugf(ctx, "Created pipeline: %v\nProps: %v\n", pl, pl.Props)
	if err != nil {
		return nil, err
	}
	t := taskqueue.NewPOSTTask("/pipelines/"+pl.ID+"/build_task.json", map[string][]string{})
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return nil, err
	}
	return pl, nil
}

// curl -v http://localhost:8080/pipelines.json
func (h *handler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pipelines, err := GetAllPipeline(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pipelines)
}

// curl -v http://localhost:8080/pipelines/1.json
func (h *handler) show(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X DELETE http://localhost:8080/pipelines/1.json
func (h *handler) destroy(c echo.Context, pl *Pipeline) error {
	ctx := c.Get("aecontext").(context.Context)
	if err := pl.destroy(ctx); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X PUT http://localhost:8080/pipelines/refresh.json
func (h *handler) refresh(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	ids, err := GetAllActivePipelineIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/%s/refresh_task.json", id), map[string][]string{})
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, ids)
}
