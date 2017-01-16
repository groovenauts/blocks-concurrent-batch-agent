package pipeline

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	apiHandler struct {}
	taskHandler struct {}
)

func init() {
	ah := &apiHandler{}
	th := &taskHandler{}

	g := e.Group("/pipelines")
	g.Use(middleware.CORS())

	g.GET("", ah.index)
	g.GET("/:id", ah.show)
	g.DELETE("/:id", ah.destroy)

	g.POST(""          , ah.create)
	g.POST("/:id/build", th.build)

	g.POST("/:id/close"     , ah.close)
	g.POST("/:id/close_task", th.close)

	g.PUT( "/:id/update"     , ah.update)
	g.POST("/:id/update_task", th.update)

	g.PUT( "/:id/resize"     , ah.resize)
	g.POST("/:id/resize_task", th.resize)

	g.GET( "/refresh"         , ah.refresh) // from cron
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
func (h *apiHandler) show(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{})
}

// curl -v http://localhost:8080/pipelines/1
func (h *apiHandler) destroy(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{})
}

// curl -v -X POST http://localhost:8080/pipelines/1/close
func (h *apiHandler) close(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{})
}
func (t *taskHandler) close(c echo.Context) error {
	return nil
}

// curl -v -X PUT http://localhost:8080/pipelines/1
func (h *apiHandler) update(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{})
}
func (t *taskHandler) update(c echo.Context) error {
	return nil
}

// curl -v -X PUT http://localhost:8080/pipelines/1/resize
func (h *apiHandler) resize(c echo.Context) error {
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
