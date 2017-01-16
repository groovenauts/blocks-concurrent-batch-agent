package pipeline

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
)

type (
	apiHandler struct {}
	taskHandler struct {}
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

func init() {
	ah := &apiHandler{}
	th := &taskHandler{}

	g := e.Group("/pipelines")
	g.Use(middleware.CORS())

	g.GET("", withAEContext(ah.index))
	g.GET("/:id", withPipeline(ah.show))
	g.DELETE("/:id", withPipeline(ah.destroy))

	g.POST(""               , withAEContext(ah.create))
	g.POST("/:id/build_task", th.build)

	g.POST("/:id/close"     , withPipeline(ah.close))
	g.POST("/:id/close_task", th.close)

	g.PUT( "/:id/update"     , withPipeline(ah.update))
	g.POST("/:id/update_task", th.update)

	g.PUT( "/:id/resize"     , withPipeline(ah.resize))
	g.POST("/:id/resize_task", th.resize)

	g.GET( "/refresh"         , withAEContext(ah.refresh)) // from cron
	g.POST("/:id/refresh_task", th.refresh)
}

// curl -v -X POST http://localhost:8080/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *apiHandler) create(c echo.Context) error {
	return c.JSON(http.StatusCreated, map[string]string{})
}
func (t *taskHandler) build(c echo.Context) error {
	return nil
}

// curl -v http://localhost:8080/pipelines
func (h *apiHandler) index(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{})
}

// curl -v http://localhost:8080/pipelines/1
func (h *apiHandler) show(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, map[string]string{})
}

// curl -v -X DELETE http://localhost:8080/pipelines/1
func (h *apiHandler) destroy(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, map[string]string{})
}

// curl -v -X POST http://localhost:8080/pipelines/1/close
func (h *apiHandler) close(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, map[string]string{})
}
func (t *taskHandler) close(c echo.Context) error {
	return nil
}

// curl -v -X PUT http://localhost:8080/pipelines/1
func (h *apiHandler) update(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, map[string]string{})
}
func (t *taskHandler) update(c echo.Context) error {
	return nil
}

// curl -v -X PUT http://localhost:8080/pipelines/1/resize
func (h *apiHandler) resize(c echo.Context, pl *Pipeline) error {
	return c.JSON(http.StatusOK, map[string]string{})
}
func (t *taskHandler) resize(c echo.Context) error {
	return nil
}

// curl -v -X PUT http://localhost:8080/pipelines/refresh
func (h *apiHandler) refresh(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{})
}
func (t *taskHandler) refresh(c echo.Context) error {
	return nil
}
