package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func InstanceGroupOperationModelToMediaType(src *model.InstanceGroupOperation) *app.CloudAsyncOperation {
	if src == nil {
		return nil
	}
	return &app.CloudAsyncOperation{
		Name:          src.Name,
		Service:       src.Service,
		OperationType: src.OperationType,
		Status:        src.Status,
		ProjectID:     src.ProjectId,
		Zone:          src.Zone,
		CreatedAt:     &src.CreatedAt,
		UpdatedAt:     &src.UpdatedAt,
		// Errors no media type field
		// Logs no media type field
		// No field for media type field "id"
		// No field for media type field "owner_id"
		// No field for media type field "owner_type"
	}
}

func PipelineBaseOperationModelToMediaType(src *model.PipelineBaseOperation) *app.CloudAsyncOperation {
	if src == nil {
		return nil
	}
	return &app.CloudAsyncOperation{
		Name:          src.Name,
		Service:       src.Service,
		OperationType: src.OperationType,
		Status:        src.Status,
		ProjectID:     src.ProjectId,
		Zone:          src.Zone,
		CreatedAt:     &src.CreatedAt,
		UpdatedAt:     &src.UpdatedAt,
		// Errors no media type field
		// Logs no media type field
		// No field for media type field "id"
		// No field for media type field "owner_id"
		// No field for media type field "owner_type"
	}
}
