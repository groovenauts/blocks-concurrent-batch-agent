package controller

import (
	"fmt"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/mjibson/goon"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// InstanceGroupController implements the InstanceGroup resource.
type InstanceGroupController struct {
	*goa.Controller
}

// NewInstanceGroupController creates a InstanceGroup controller.
func NewInstanceGroupController(service *goa.Service) *InstanceGroupController {
	return &InstanceGroupController{Controller: service.NewController("InstanceGroupController")}
}

// Create runs the create action.
func (c *InstanceGroupController) Create(ctx *app.CreateInstanceGroupContext) error {
	// InstanceGroupController_Create: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		m := InstanceGroupPayloadToModel(ctx.Payload)
		m.Status = model.ConstructionStarting
		err := datastore.RunInTransaction(appCtx, func(c context.Context) error {
			store := &model.InstanceGroupStore{ParentKey: orgKey}
			_, err := store.Put(c, &m)
			if err != nil {
				return ctx.BadRequest(goa.ErrBadRequest(err))
			}

			if err := PostTask(appCtx, fmt.Sprintf("/construction_tasks?resource_id=%d", m.Name), 0); err != nil {
				return err
			}
			return nil
		}, nil)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}
		return ctx.Created(InstanceGroupModelToMediaType(&m))
	})

	// InstanceGroupController_Create: end_implement
}

// Delete runs the delete action.
func (c *InstanceGroupController) Delete(ctx *app.DeleteInstanceGroupContext) error {
	// InstanceGroupController_Delete: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.InstanceGroup) error {
				switch m.Status {
				case model.ConstructionError, model.DestructionError, model.Destructed: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't delete because the InstanceGroup %q is %s", m.Name, m.Status))
				}

				if err := store.Delete(appCtx, m); err != nil {
					return err
				}

				return ctx.OK(InstanceGroupModelToMediaType(m))
			})
		}, nil)
	})

	// InstanceGroupController_Delete: end_implement
}

// Destruct runs the destruct action.
func (c *InstanceGroupController) Destruct(ctx *app.DestructInstanceGroupContext) error {
	// InstanceGroupController_Destruct: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}
		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.InstanceGroup) error {
				switch m.Status {
				case model.Constructed: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't destruct because the InstanceGroup %q is %s", m.Name, m.Status))
				}

				m.Status = model.DestructionStarting
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PostTask(appCtx, fmt.Sprintf("/destruction_tasks?resource_id=%d", m.Name), 0); err != nil {
					return err
				}
				return ctx.Created(InstanceGroupModelToMediaType(m))
			})
		}, nil)
	})

	// InstanceGroupController_Destruct: end_implement
}

// List runs the list action.
func (c *InstanceGroupController) List(ctx *app.ListInstanceGroupContext) error {
	// InstanceGroupController_List: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}
		models, err := store.All(appCtx)
		if err != nil {
			return ctx.BadRequest(goa.ErrBadRequest(err))
		}

		res := app.InstanceGroupCollection{}
		for _, m := range models {
			res = append(res, InstanceGroupModelToMediaType(m))
		}
		return ctx.OK(res)
	})

	// InstanceGroupController_List: end_implement
}

// Resize runs the resize action.
func (c *InstanceGroupController) Resize(ctx *app.ResizeInstanceGroupContext) error {
	// InstanceGroupController_Resize: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}

		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.InstanceGroup) error {
				switch m.Status {
				case model.Constructed, model.ResizeStarting, model.ResizeRunning: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't resize because the InstanceGroup %q is %s", m.Name, m.Status))
				}

				if m.InstanceSizeRequested >= ctx.NewSize {
					return ctx.OK(InstanceGroupModelToMediaType(m))
				}

				m.InstanceSizeRequested = ctx.NewSize
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PostTask(appCtx, fmt.Sprintf("/resizing_tasks?resource_id=%d", m.Name), 0); err != nil {
					return err
				}
				return ctx.Created(InstanceGroupModelToMediaType(m))
			})
		}, nil)
	})

	// InstanceGroupController_Resize: end_implement
}

// Show runs the show action.
func (c *InstanceGroupController) Show(ctx *app.ShowInstanceGroupContext) error {
	// InstanceGroupController_Show: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}
		return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.InstanceGroup) error {
			return ctx.OK(InstanceGroupModelToMediaType(m))
		})
	})

	// InstanceGroupController_Show: end_implement
}

// StartHealthCheck runs the start_health_check action.
func (c *InstanceGroupController) StartHealthCheck(ctx *app.StartHealthCheckInstanceGroupContext) error {
	// InstanceGroupController_StartHealthCheck: start_implement

	// Put your logic here
	return WithAuthOrgKey(ctx.Context, ctx.OrgID, func(orgKey *datastore.Key) error {
		appCtx := appengine.NewContext(ctx.Request)
		store := &model.InstanceGroupStore{ParentKey: orgKey}

		return datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			return c.member(appCtx, store, ctx.Name, ctx.BadRequest, ctx.NotFound, func(m *model.InstanceGroup) error {
				switch m.Status {
				case model.Constructed, model.HealthCheckError, model.ResizeStarting, model.ResizeRunning: // Through
				default:
					return ctx.Conflict(fmt.Errorf("Can't start health check because the InstanceGroup %q is %s", m.Name, m.Status))
				}

				if m.HealthCheckId != "" {
					log.Infof(appCtx, "No new health check started because Health check %q is already running", m.HealthCheckId)
					return ctx.OK(InstanceGroupModelToMediaType(m))
				}

				hc := &model.InstanceGroupHealthCheck{}
				g := goon.FromContext(appCtx)
				hcStore := &model.InstanceGroupHealthCheckStore{ParentKey: g.Key(m)}

				if _, err := hcStore.Create(appCtx, hc); err != nil {
					return err
				}

				m.HealthCheckId = fmt.Sprintf("%d", hc.Id)
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PutTask(appCtx, fmt.Sprintf("/instance_group_health_checks/%d", hc.Id), 0); err != nil {
					return err
				}
				return ctx.Created(InstanceGroupModelToMediaType(m))
			})
		}, nil)
	})

	// InstanceGroupController_StartHealthCheck: end_implement
}
