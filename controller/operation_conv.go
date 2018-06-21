package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func OperationModelToMediaType(src *model.Operation) *app.Operation {
	if src == nil {
		return nil
	}
	return &app.Operation{
		OwnerType:     src.OwnerType,
		Name:          src.Name,
		Service:       src.Service,
		OperationType: src.OperationType,
		Status:        src.Status,
		ProjectID:     src.ProjectId,
		Zone:          src.Zone,
		CreatedAt:     &src.CreatedAt,
		UpdatedAt:     &src.UpdatedAt,
		// OwnerID no media type field
		// No field for media type field "id"
	}
}
