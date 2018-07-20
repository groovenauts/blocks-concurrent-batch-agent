package controller

import (
	"fmt"
)

func pathToInstanceGroupAction(orgID, name, action string) string {
	return fmt.Sprintf("/orgs/%s/instance_groups/%s/%s", orgID, name, action)
}

func pathToInstanceGroupTask(orgID, name, task string, taskId int64) string {
	var action string
	if taskId == 0 {
		action = task
	} else {
		action = fmt.Sprintf("%s/%d", task, taskId)
	}
	return pathToInstanceGroupAction(orgID, name, action)
}

func pathToInstanceGroupConstructionTask(orgID, name string, taskId int64) string {
	return pathToInstanceGroupTask(orgID, name, "construction_tasks", taskId)
}

func pathToInstanceGroupDestructionTask(orgID, name string, taskId int64) string {
	return pathToInstanceGroupTask(orgID, name, "destruction_tasks", taskId)
}

func pathToInstanceGroupResizingTask(orgID, name string, taskId int64) string {
	return pathToInstanceGroupTask(orgID, name, "resizing_tasks", taskId)
}

func pathToInstanceGroupHealthCheckTask(orgID, name string, taskId int64) string {
	return pathToInstanceGroupTask(orgID, name, "health_check_tasks", taskId)
}
