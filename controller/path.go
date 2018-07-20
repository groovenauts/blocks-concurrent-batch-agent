package controller

import (
	"fmt"
)

func pathToInstanceGroupAction(orgID, name, action string) string {
	return fmt.Sprintf("/orgs/%s/instance_groups/%s/%s", orgID, name, action)
}

func pathToInstanceGroupTask(orgID, name, task string, taskId int64) string {
	return pathToInstanceGroupAction(orgID, name, fmt.Sprintf("%s/%d", task, taskId))
}
