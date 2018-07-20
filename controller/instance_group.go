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
		err := datastore.RunInTransaction(appCtx, func(appCtx context.Context) error {
			store := &model.InstanceGroupStore{ParentKey: orgKey}
			_, err := store.Create(appCtx, &m)
			if err != nil {
				return ctx.BadRequest(goa.ErrBadRequest(err))
			}

			if err := PostTask(appCtx, pathToInstanceGroupAction(ctx.OrgID, m.Name, "construction_tasks"), 0); err != nil {
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
				case model.Constructed, model.HealthCheckError:
					m.Status = model.DestructionStarting
					if _, err := store.Update(appCtx, m); err != nil {
						return err
					}
					if err := PostTask(appCtx, pathToInstanceGroupAction(ctx.OrgID, m.Name, "destruction_tasks"), 0); err != nil {
						return err
					}
					return ctx.Created(InstanceGroupModelToMediaType(m))
				case model.ResizeStarting, model.ResizeRunning:
					m.Status = model.ResizeWaiting
					if _, err := store.Update(appCtx, m); err != nil {
						return err
					}
					return ctx.OK(InstanceGroupModelToMediaType(m))
				case model.ResizeWaiting:
					return ctx.OK(InstanceGroupModelToMediaType(m))
				default:
					return ctx.Conflict(fmt.Errorf("Can't destruct because the InstanceGroup %q is %s", m.Name, m.Status))
				}
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

				var response func(r *app.InstanceGroup) error
				startResizing := (m.Status == model.Constructed)
				if startResizing {
					m.Status = model.ResizeStarting
					response = ctx.Created
				} else {
					response = ctx.OK
				}
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}
				if startResizing {
					if err := PostTask(appCtx, pathToInstanceGroupAction(ctx.OrgID, m.Name, "resizing_tasks"), 0); err != nil {
						return err
					}
				}
				return response(InstanceGroupModelToMediaType(m))
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

				if m.HealthCheckId != 0 {
					log.Infof(appCtx, "No new health check started because Health check %q is already running", m.HealthCheckId)
					return ctx.OK(InstanceGroupModelToMediaType(m))
				}

				hc := &model.InstanceGroupHealthCheck{}
				g := goon.FromContext(appCtx)
				hcStore := &model.InstanceGroupHealthCheckStore{ParentKey: g.Key(m)}

				if _, err := hcStore.Create(appCtx, hc); err != nil {
					return err
				}

				m.HealthCheckId = hc.Id
				if _, err := store.Update(appCtx, m); err != nil {
					return err
				}

				if err := PostTask(appCtx, pathToInstanceGroupAction(ctx.OrgID, m.Name, fmt.Sprintf("health_checks/%d", hc.Id)), 0); err != nil {
					return err
				}
				return ctx.Created(InstanceGroupModelToMediaType(m))
			})
		}, nil)
	})

	// InstanceGroupController_StartHealthCheck: end_implement
}
