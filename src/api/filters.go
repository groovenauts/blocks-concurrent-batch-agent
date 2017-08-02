package api

import (
	"fmt"
	"net/http"

	"models"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func orgBy(key string, f func(c echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		org_id := c.Param(key)
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

func PlToOrg(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		pl := c.Get("pipeline").(*models.Pipeline)
		pl.LoadOrganization(ctx)
		c.Set("organization", pl.Organization)
		return impl(c)
	}
}

func PjToPl(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		pj := c.Get("job").(*models.PipelineJob)
		pj.LoadPipeline(ctx)
		c.Set("pipeline", pj.Pipeline)
		return impl(c)
	}
}

func plBy(key string, impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param(key)

		var accessor *models.PipelineAccessor
		obj := c.Get("organization")

		if obj == nil {
			accessor = models.GlobalPipelineAccessor
		} else {
			org, ok := c.Get("organization").(*models.Organization)
			if ok {
				accessor = org.PipelineAccessor()
			} else {
				msg := fmt.Sprintf("invalid organization: %v", obj)
				log.Errorf(ctx, "plBy %s\n", msg)
				panic(msg)
			}
		}

		pl, err := accessor.Find(ctx, id)
		switch {
		case err == models.ErrNoSuchPipeline:
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "plBy %v id: %v\n", err, id)
			return err
		}
		c.Set("pipeline", pl)
		return impl(c)
	}
}

func pjBy(key string, impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param(key)

		var accessor *models.PipelineJobAccessor
		obj := c.Get("pipeline")

		if obj == nil {
			accessor = models.GlobalPipelineJobAccessor
		} else {
			pl, ok := obj.(*models.Pipeline)
			if ok {
				accessor = pl.JobAccessor()
			} else {
				msg := fmt.Sprintf("invalid pipeline: %v", obj)
				log.Errorf(ctx, "pjBy %s\n", msg)
				panic(msg)
			}
		}

		pj, err := accessor.Find(ctx, id)
		switch {
		case err == models.ErrNoSuchPipelineJob:
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "plBy %v id: %v\n", err, id)
			return err
		}
		c.Set("job", pj)
		return impl(c)
	}
}
