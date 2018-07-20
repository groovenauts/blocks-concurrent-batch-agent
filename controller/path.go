package controller

import (
	"fmt"
)

func pathToInstanceGroupAction(orgID, name, action string) string {
	return fmt.Sprintf("/orgs/%s/instance_groups/%s/%s", orgID, name, action)
}
