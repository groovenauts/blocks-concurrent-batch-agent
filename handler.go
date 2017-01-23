package pipeline

import (
	"fmt"
	"io"
	"net/http"
	"html/template"

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

	g.GET(".json", withAEContext(h.index))
	g.GET("/:id.json", withPipeline(h.show))
	g.DELETE("/:id.json", withPipeline(h.destroy))

	g.POST(".json", withAEContext(h.create))
	g.POST("/:id/build_task.json", pipelineTask("build"))

	actions := []string{"close", "update", "resize"}
	for _, action := range actions {
		g.PUT("/:id/"+action+".json", callPipelineTask(action))
		g.POST("/:id/"+action+"_task.json", pipelineTask(action))
	}

	g.GET("/refresh.json", withAEContext(h.refresh)) // from cron
	g.POST("/:id/refresh_task.json", pipelineTask("refresh"))

	t := &Template{
    templates: template.Must(template.ParseGlob("views/*.html")),
	}

	e.Renderer = t
	e.GET("/pipelines.html", withAEContext(h.indexPage))
	e.GET("/pipelines/new.html", withAEContext(h.newPage))
	e.POST("/pipelines.html", withAEContext(h.createPage))
}

func withAEContext(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		req := c.Request()
		ctx := appengine.NewContext(req)
		c.Set("aecontext", ctx)
		return impl(c)
	}
}

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
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

// curl -v -X PUT http://localhost:8080/pipelines/1/close.json
// curl -v -X PUT http://localhost:8080/pipelines/1/update.json
// curl -v -X PUT http://localhost:8080/pipelines/1/resize.json
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

// curl -v -X POST http://localhost:8080/pipelines/1/build_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/close_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/update_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/resize_task.json
// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task.json
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

// curl -v -X POST http://localhost:8080/pipelines.json --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *handler) create(c echo.Context) error {
	pl, err := h.createImpl(c)
	if err != nil { return err }
	return c.JSON(http.StatusCreated, pl)
}

// POST http://localhost:8080/pipelines.html
func (h *handler) createPage(c echo.Context) error {
	_, err := h.createImpl(c)
	if err != nil {
		return c.Render(http.StatusOK, "new", err)
	}
	return c.Redirect(http.StatusSeeOther, "/pipelines.html")
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
	t := taskqueue.NewPOSTTask("/pipelines/"+pl.ID+"/build_task", map[string][]string{})
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

// curl -v http://localhost:8080/pipelines.html
func (h *handler) indexPage(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	log.Debugf(ctx, "indexPage\n")
	pipelines, err := GetAllPipeline(ctx)
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	log.Debugf(ctx, "indexPage pipelines: %v\n", pipelines)
	r := c.Render(http.StatusOK, "index", pipelines)
	log.Debugf(ctx, "indexPage r: %v\n", r)
	return r
}

// curl -v http://localhost:8080/pipelines.html
func (h *handler) newPage(c echo.Context) error {
	return c.Render(http.StatusOK, "new", nil)
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
		t := taskqueue.NewPOSTTask(fmt.Sprintf("/%s/refresh_task", id), map[string][]string{})
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, ids)
}
