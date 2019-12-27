package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	// "google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/groovenauts/blocks-concurrent-batch-agent/src/models"
)

type OrganizationsHandler struct {
	Views Views
}

// GET http://localhost:8080/admin/orgs
type ResOrgsIndex struct {
	Flash         *Flash
	Organizations []*models.Organization
}

func (h *OrganizationsHandler) Index(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	orgs, err := models.GlobalOrganizationAccessor.All(ctx)
	if err != nil {
		log.Errorf(ctx, "indexPage error: %v\n", err)
		return err
	}
	r := ResOrgsIndex{
		Organizations: orgs,
	}
	r.Flash = c.Get("flash").(*Flash)
	return h.Views.Render(c, http.StatusOK, "index", &r)
}

// GET http://localhost:8080/admin/orgs/new
type ResOrgsNew struct {
	Flash        *Flash
	Organization *models.Organization
}

func (h *OrganizationsHandler) New(c echo.Context) error {
	org := &models.Organization{}
	r := &ResOrgsNew{
		Organization: org,
		Flash:        c.Get("flash").(*Flash),
	}
	return h.Views.Render(c, http.StatusOK, "new", r)
}

// POST http://localhost:8080/admin/orgs
func (h *OrganizationsHandler) Create(c echo.Context) error {
	ctx := c.Get("aecontext").(context.Context)
	org := &models.Organization{}
	if err := c.Bind(org); err != nil {
		return err
	}

	err := org.Create(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to create Organization: %v because of %v\n", org, err)
		r := &ResOrgsNew{
			Organization: org,
			Flash: &Flash{
				Alert: fmt.Sprintf("Failed to create Organization. error: %v", err),
			},
		}
		return h.Views.Render(c, http.StatusOK, "new", r)
	}
	setFlash(c, "notice", fmt.Sprintf("Organization %q is created successfully.", org.Name))
	return c.Redirect(http.StatusFound, "/admin/orgs")
}

func (h *OrganizationsHandler) Identified(f func(c echo.Context, org *models.Organization) error) func(c echo.Context) error {
	return withFlash(func(c echo.Context) error {
		ctx := c.Get("aecontext").(context.Context)
		org, err := models.GlobalOrganizationAccessor.Find(ctx, c.Param("id"))
		if err == models.ErrNoSuchOrganization {
			setFlash(c, "alert", fmt.Sprintf("Organization not found for id: %v", c.Param("id")))
			return c.Redirect(http.StatusFound, "/admin/orgs")
		}
		if err != nil {
			setFlash(c, "alert", fmt.Sprintf("Failed to find Organization for id: %v error: %v", c.Param("id"), err))
			return c.Redirect(http.StatusFound, "/admin/orgs")
		}
		return f(c, org)
	})
}

// GET http://localhost:8080/admin/orgs/:id
type ResOrgsShow struct {
	Flash        *Flash
	Organization *models.Organization
}

func (h *OrganizationsHandler) Show(c echo.Context, org *models.Organization) error {
	r := &ResOrgsShow{
		Organization: org,
		Flash:        c.Get("flash").(*Flash),
	}
	return h.Views.Render(c, http.StatusOK, "show", r)
}

// GET http://localhost:8080/admin/orgs/:id/edit
type ResOrgsEdit struct {
	Flash        *Flash
	Organization *models.Organization
}

func (h *OrganizationsHandler) Edit(c echo.Context, org *models.Organization) error {
	r := &ResOrgsEdit{
		Organization: org,
		Flash:        c.Get("flash").(*Flash),
	}
	return h.Views.Render(c, http.StatusOK, "edit", r)
}

// POST http://localhost:8080/admin/orgs/:id/update
func (h *OrganizationsHandler) Update(c echo.Context, org *models.Organization) error {
	ctx := c.Get("aecontext").(context.Context)
	if err := c.Bind(org); err != nil {
		return err
	}
	err := org.Update(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to update Organization: %v because of %v\n", org, err)
		r := &ResOrgsEdit{
			Organization: org,
			Flash: &Flash{
				Alert: fmt.Sprintf("Failed to update Organization. error: %v", err),
			},
		}
		return h.Views.Render(c, http.StatusOK, "edit", r)
	}
	setFlash(c, "notice", fmt.Sprintf("Organization %q is updated successfully.", org.Name))
	return c.Redirect(http.StatusFound, "/admin/orgs/"+org.ID)
}

// DELETE http://localhost:8080/admin/orgs/:id
func (h *OrganizationsHandler) Destroy(c echo.Context, org *models.Organization) error {
	ctx := c.Get("aecontext").(context.Context)
	err := org.Destroy(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to destroy Organization: %v because of %v\n", org, err)
		setFlash(c, "alert", fmt.Sprintf("Failed to destroy Organization. id: %v error: %v", org.ID, err))
		return c.Redirect(http.StatusFound, "/admin/orgs")
	}
	setFlash(c, "notice", fmt.Sprintf("The Organization is deleted successfully. id: %v", org.ID))
	return c.Redirect(http.StatusFound, "/admin/orgs")
}
