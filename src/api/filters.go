package api

import (
	"context"
	"fmt"

	"github.com/labstack/echo"
	"google.golang.org/appengine/log"

	"github.com/groovenauts/blocks-concurrent-batch-server/src/models"
)

func orgBy(key string, statusNotFound int, f func(c echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		org_id := c.Param(key)
		org, err := models.GlobalOrganizationAccessor.Find(ctx, org_id)
		if err == models.ErrNoSuchOrganization {
			if statusNotFound < 400 {
				log.Errorf(ctx, "Organization not found for %v\n", org_id)
			}
			return c.JSON(statusNotFound, map[string]string{"message": "No Organization found for " + org_id})
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

func JobToPl(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		job := c.Get("job").(*models.Job)
		job.LoadPipeline(ctx)
		c.Set("pipeline", job.Pipeline)
		return impl(c)
	}
}

func OperationToPl(impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		operation := c.Get("operation").(*models.PipelineOperation)
		operation.LoadPipeline(ctx)
		c.Set("pipeline", operation.Pipeline)
		return impl(c)
	}
}

func plBy(key string, statusNotFound int, impl func(c echo.Context) error) func(c echo.Context) error {
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
			if statusNotFound < 400 {
				log.Errorf(ctx, "Pipeline not found for %v\n", id)
			}
			return c.JSON(statusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "plBy %v id: %v\n", err, id)
			return err
		}
		c.Set("pipeline", pl)
		return impl(c)
	}
}

func jobBy(key string, statusNotFound int, impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param(key)

		var accessor *models.JobAccessor
		obj := c.Get("pipeline")

		if obj == nil {
			accessor = models.GlobalJobAccessor
		} else {
			pl, ok := obj.(*models.Pipeline)
			if ok {
				accessor = pl.JobAccessor()
			} else {
				msg := fmt.Sprintf("invalid pipeline: %v", obj)
				log.Errorf(ctx, "jobBy %s\n", msg)
				panic(msg)
			}
		}

		job, err := accessor.Find(ctx, id)
		switch {
		case err == models.ErrNoSuchJob:
			if statusNotFound < 400 {
				log.Errorf(ctx, "Job not found for %v\n", id)
			}
			return c.JSON(statusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "plBy %v id: %v\n", err, id)
			return err
		}
		c.Set("job", job)
		return impl(c)
	}
}

func operationBy(idName string, statusNotFound int, impl func(c echo.Context) error) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		id := c.Param(idName)

		var accessor *models.PipelineOperationAccessor
		obj := c.Get("pipeline")

		if obj == nil {
			accessor = models.GlobalPipelineOperationAccessor
		} else {
			pl, ok := obj.(*models.Pipeline)
			if ok {
				accessor = pl.OperationAccessor()
			} else {
				msg := fmt.Sprintf("invalid pipeline: %v", obj)
				log.Errorf(ctx, "operationBy %s\n", msg)
				panic(msg)
			}
		}

		operation, err := accessor.Find(ctx, id)
		switch {
		case err == models.ErrNoSuchPipelineOperation:
			if statusNotFound < 400 {
				log.Errorf(ctx, "Operation not found for %v\n", id)
			}
			return c.JSON(statusNotFound, map[string]string{"message": "Not found for " + id})
		case err != nil:
			log.Errorf(ctx, "operationBy %v id: %v\n", err, id)
			return err
		}
		c.Set("operation", operation)
		return impl(c)
	}
}
