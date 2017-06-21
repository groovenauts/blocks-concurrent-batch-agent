package api

import (
	"fmt"
	"net/http"
	"regexp"

	"gae_support"
	"models"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

type handler struct {
	Actions map[string](func(c echo.Context) error)
}

var e *echo.Echo

const (
	AUTH_HEADER = "Authorization"
)

func Setup(echo *echo.Echo) {
	e = echo

	h := &handler{}
	h.buildActions()

	g := e.Group("/orgs/:org_id/pipelines")
	g.Use(middleware.CORS())

	g.GET("", h.Actions["index"])
	g.POST("", h.Actions["create"])
	g.GET("/subscriptions", h.Actions["subscriptions"])

	g.GET("/:id", h.Actions["show"])
	g.DELETE("/:id", h.Actions["destroy"])
	// g.POST("/:id/build_task", h.Actions["build"])
	g.POST("/:id/build_task", gae_support.With(h.withOrg(h.withAuth(h.Identified(h.pipelineTask("build"))))))
	g.PUT("/:id/close", h.Actions["close"])
	// g.POST("/:id/close_task", h.Actions["close_task"])
	g.POST("/:id/close_task", gae_support.With(h.withOrg(h.withAuth(h.Identified(h.pipelineTask("close"))))))

	g = e.Group("/pipelines")
	g.Use(middleware.CORS())
	g.GET("/refresh", h.Actions["refresh"])
	g.POST("/:id/refresh_task", h.Actions["refresh_task"])
}

func (h *handler) buildActions() {
	h.Actions = map[string](func(c echo.Context) error){
		"index":         gae_support.With(h.withOrg(h.withAuth(h.index))),
		"create":        gae_support.With(h.withOrg(h.withAuth(h.create))),
		"subscriptions": gae_support.With(h.withOrg(h.withAuth(h.subscriptions))),
		"show":          gae_support.With(h.withOrg(h.withAuth(h.Identified(h.show)))),
		"destroy":       gae_support.With(h.withOrg(h.withAuth(h.Identified(h.destroy)))),
		"close":         gae_support.With(h.withOrg(h.withAuth(h.Identified(h.close)))),
		"refresh":       gae_support.With(h.refresh), // Don't use withAuth because this is called from cron
		"refresh_task":  gae_support.With(h.Identified(h.pipelineTask("refresh"))),
		// "build_task": gae_support.With(h.withOrg(h.withAuth(h.Identified(h.pipelineTask("build"))))),
		// "close_task": gae_support.With(h.withOrg(h.withAuth(h.Identified(h.pipelineTask("close"))))),
	}
}

func (h *handler) withOrg(f func(c echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		org_id := c.Param("org_id")
		org, err := models.GlobalOrganizationAccessor.Find(ctx, org_id)
		if err == models.ErrNoSuchOrganization {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "No Organization found for " + org_id})
		}
		if err != nil {
			log.Errorf(ctx, "Failed to find Organization id: %v because of %v\n", org_id, err)
			return err
		}
		c.Set("organization", org)
		return f(c)
	}
}

func (h *handler) withAuth(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
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
		org := c.Get("organization").(*models.Organization)
		_, err := org.AuthAccessor().FindWithToken(ctx, token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token"})
		}
		return impl(c)
	}
}

func (h *handler) Identified(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param("id")

		var accessor *models.PipelineAccessor
		org := c.Get("organization").(*models.Organization)
		if org == nil {
			accessor = models.GlobalPipelineAccessor
		} else {
			accessor = org.PipelineAccessor()
		}

		pl, err := accessor.Find(ctx, id)
		switch {
		case err == models.ErrNoSuchPipeline:
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "@Identified %v id: %v\n", err, id)
			return err
		}
		c.Set("pipeline", pl)
		return impl(c)
	}
}

// curl -v -X PUT http://localhost:8080/orgs/2/pipelines/1/close
func (h *handler) close(c echo.Context, pl *models.Pipeline) error {
	ctx := c.Get("aecontext").(context.Context)
	org := c.Get("organization").(*models.Organization)
	pl := c.Get("pipeline").(*models.Pipeline)
	id := c.Param("id")
	req := c.Request()
	t := taskqueue.NewPOSTTask(fmt.Sprintf("/orgs/%s/pipelines/%s/close_task", org.ID, id), map[string][]string{})
	t.Header.Add(AUTH_HEADER, req.Header.Get(AUTH_HEADER))
	if _, err := taskqueue.Add(ctx, t, ""); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, pl)
}

// curl -v -X POST http://localhost:8080/orgs/2/pipelines/1/build_task
// curl -v -X	POST http://localhost:8080/orgs/2/pipelines/1/close_task
// curl -v -X	POST http://localhost:8080/pipelines/1/refresh_task
func (h *handler) pipelineTask(action string) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		pl := c.Get("pipeline").(*models.Pipeline)
		err := pl.Process(ctx, action)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, pl)
	}
}

// curl -v -X POST http://localhost:8080/orgs/2/pipelines --data '{"id":"2","name":"akm"}' -H 'Content-Type: application/json'
func (h *handler) create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	req := c.Request()
	pl := &models.Pipeline{}
	if err := c.Bind(pl); err != nil {
		log.Errorf(ctx, "err: %v\n", err)
		log.Errorf(ctx, "req: %v\n", req)
		return err
	}
	org := c.Get("organization").(*models.Organization)
	pl.Organization = org
	err := pl.Create(ctx)
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

// curl -v http://localhost:8080/orgs/2/pipelines
func (h *handler) index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	org := c.Get("organization").(*models.Organization)
	pipelines, err := org.PipelineAccessor().GetAll(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, pipelines)
}

// curl -v http://localhost:8080/orgs/2/pipelines/subscriptions
func (h *handler) subscriptions(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	org := c.Get("organization").(*models.Organization)
	subscriptions, err := org.PipelineAccessor().GetActiveSubscriptions(ctx)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, subscriptions)
}

// curl -v http://localhost:8080/orgs/2/pipelines/1
func (h *handler) show(c echo.Context) error {
	pl := c.Get("pipeline").(*models.Pipeline)
	return c.JSON(http.StatusOK, pl)
}

// curl -v -X DELETE http://localhost:8080/orgs/2/pipelines/1
func (h *handler) destroy(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	pl := c.Get("pipeline").(*models.Pipeline)
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
		ids, err := models.GlobalPipelineAccessor.GetIDsByStatus(ctx, st)
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
