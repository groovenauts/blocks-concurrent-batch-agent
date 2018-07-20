package controller

import (
	"encoding/json"
	"math"
	"strconv"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	"github.com/goadesign/goa"
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

// InstanceGroupHealthCheckController implements the InstanceGroupHealthCheck resource.
type InstanceGroupHealthCheckController struct {
	*goa.Controller
}

// NewInstanceGroupHealthCheckController creates a InstanceGroupHealthCheck controller.
func NewInstanceGroupHealthCheckController(service *goa.Service) *InstanceGroupHealthCheckController {
	return &InstanceGroupHealthCheckController{Controller: service.NewController("InstanceGroupHealthCheckController")}
}

// Execute runs the execute action.
func (c *InstanceGroupHealthCheckController) Execute(ctx *app.ExecuteInstanceGroupHealthCheckContext) error {
	// InstanceGroupHealthCheckController_Execute: start_implement

	// Put your logic here
	appCtx := appengine.NewContext(ctx.Request)
	orgIdString := ctx.OrgID
	name := ctx.Name
	hcIdString := ctx.ID

	orgId, err := strconv.ParseInt(orgIdString, 10, 64)
	if err != nil {
		log.Errorf(appCtx, "Invalid Organization ID: %v because of %v\n", orgIdString, err)
		return ctx.NoContent(nil)
	}

	goon := model.GoonFromContext(ctx)

	orgKey, err := goon.KeyError(&model.Organization{ID: orgId})
	if err != nil {
		log.Errorf(appCtx, "Can't get Organization key from %v because of %v\n", orgId, err)
		return ctx.NoContent(nil)
	}

	return datastore.RunInTransaction(appCtx, func(c context.Context) error {
		igStore := &model.InstanceGroupStore{ParentKey: orgKey}
		ig, err := igStore.ByID(appCtx, name)
		if err != nil {
			log.Errorf(appCtx, "Can't get InstanceGroup key from %d / %v because of %v \n", orgId, name, err)
			return ctx.NoContent(nil)
		}

		switch ig.Status {
		case model.Constructed,
			model.HealthCheckError,
			model.ResizeStarting,
			model.ResizeRunning:
		case model.ResizeWaiting,
			model.DestructionStarting,
			model.DestructionRunning,
			model.DestructionError,
			model.Destructed:
			log.Infof(appCtx, "Quit Health Check because instance group %s/%s is %s\n", orgId, name, ig.Status)
			return ctx.NoContent(nil)
		default:
			log.Errorf(appCtx, "Unexpected Instance Group status %q of %s/%s\n", ig.Status, orgId, name)
			return ctx.NoContent(nil)
		}

		igKey, err := goon.KeyError(ig)
		if err != nil {
			log.Errorf(appCtx, "Can't get InstanceGroup key from %v because of %v\n", ig, err)
			return ctx.OK(nil)
		}

		hcId, err := strconv.ParseInt(hcIdString, 10, 64)
		if err != nil {
			log.Errorf(appCtx, "Failed to parse Instance Group Health Check ID %q because of %s/%s\n", hcIdString, err)
			return ctx.NoContent(nil)
		}

		hcStore := &model.InstanceGroupHealthCheckStore{ParentKey: igKey}
		hc, err := hcStore.ByID(appCtx, hcId)
		if err != nil {
			log.Errorf(appCtx, "Can't get InstanceGrouphealthCheck %s/%s/%s because of %v \n", orgId, name, hcId, err)
			return ctx.NoContent(nil)
		}

		if hc.Id != ig.HealthCheckId {
			log.Infof(appCtx, "Quit Health Check %s because instance group %s/%s has another health check\n", hc.Id, orgId, name, ig.HealthCheckId)
			return ctx.OK(InstanceGroupHealthCheckModelToMediaType(hc))
		}

		servicer, err := model.DefaultInstanceGroupServicer(appCtx)
		if err != nil {
			log.Errorf(appCtx, "Can't get DefaultInstanceGroupServicer because of %v\n", err)
			return err // Retry
		}

		resp, err := servicer.ListManagedInstances(ig.ProjectID, ig.Zone, ig.DeploymentName+"-igm")
		if err != nil {
			log.Errorf(appCtx, "Can't get DefaultInstanceGroupServicer because of %v\n", err)
			return err // Retry
		}

		// https://godoc.org/google.golang.org/api/compute/v1#ManagedInstance
		bytes, err := json.MarshalIndent(resp.ManagedInstances, "", "  ")
		if err != nil {
			return err // Retry
		}

		instancesJson := string(bytes)
		log.Debugf(appCtx, "Managed Instances\n", instancesJson)
		hc.LastResult = instancesJson

		var response func(r *app.InstanceGroupHealthCheck) error

		total := len(resp.ManagedInstances)
		if total > 0 {
			counts := map[string]int{}
			for _, instance := range resp.ManagedInstances {
				counts[instance.InstanceStatus] += 1
			}

			// Instance status transition
			// See https://cloud.google.com/compute/docs/instances/checking-instance-status
			//
			//   (START) ==> PROVISIONING ==> STAGING ==> RUNNING ==> STOPPING ==> TERMINATED ==> (END)
			//                    ^                                                    |
			//                    |                                                    |
			//                    |----------------------------------------------------|
			//
			// ? How about these statuses ?
			//   STOPPED
			//   SUSPENDED
			//   SUSPENDING
			working := counts["PROVISIONING"] + counts["STAGING"] + counts["RUNNING"]
			workingRate := int(math.Ceil(float64(working) / float64(total)))

			if ig.Status == model.HealthCheckError {
				if (working >= ig.HealthCheck.MinimumRunningSize) && (workingRate >= ig.HealthCheck.MinimumRunningPercentage) {
					log.Infof(appCtx, "Working Instances %d is less than MinimumRunningSize %d\n", working, ig.HealthCheck.MinimumRunningSize)
					ig.Status = model.Constructed
				}
			} else {
				if working < ig.HealthCheck.MinimumRunningSize {
					log.Warningf(appCtx, "Working Instances %d is less than MinimumRunningSize %d\n", working, ig.HealthCheck.MinimumRunningSize)
					ig.Status = model.HealthCheckError
				}
				if workingRate < ig.HealthCheck.MinimumRunningPercentage {
					log.Warningf(appCtx, "Working Instance percentages %d is less than MinimumRunningPercentage %d\n", workingRate, ig.HealthCheck.MinimumRunningPercentage)
					ig.Status = model.HealthCheckError
				}
			}
			if ig.Status == model.HealthCheckError {
				response = ctx.Accepted
			} else {
				response = ctx.Created
			}
		} else {
			response = ctx.Created
		}

		_, err = hcStore.Update(appCtx, hc)
		if err != nil {
			log.Errorf(appCtx, "Failed to update InstanceGroupHealthCheck: %v because of %v\n", hc, err)
			return err // Retry
		}

		if err := PutTask(appCtx, pathToInstanceGroupTask(ctx.OrgID, ctx.Name, "health_check_tasks", hc.Id), 0); err != nil {
			return err //Retry
		}

		return response(InstanceGroupHealthCheckModelToMediaType(hc))
	}, nil)
	// InstanceGroupHealthCheckController_Execute: end_implement
}
