package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func PipelineContainerPayloadToModel(src *app.PipelineContainer) model.PipelineContainer {
	if src == nil {
		return model.PipelineContainer{}
	}
	return model.PipelineContainer{
		Name:             src.Name,
		Size:             src.Size,
		Command:          StringPointerToString(src.Command),
		Options:          StringPointerToString(src.Options),
		StackdriverAgent: BoolPointerToBool(src.StackdriverAgent),
	}
}

func PipelineContainerModelToMediaType(src *model.PipelineContainer) *app.PipelineContainer {
	if src == nil {
		return nil
	}
	return &app.PipelineContainer{
		Name:             src.Name,
		Size:             src.Size,
		Command:          &src.Command,
		Options:          &src.Options,
		StackdriverAgent: &src.StackdriverAgent,
	}
}

func PipelineBasePayloadToModel(src *app.PipelineBasePayload) model.PipelineBase {
	if src == nil {
		return model.PipelineBase{}
	}
	return model.PipelineBase{
		Name:             src.Name,
		InstanceGroup:    InstanceGroupBodyPayloadToModel(src.InstanceGroup),
		Container:        PipelineContainerPayloadToModel(src.Container),
		HibernationDelay: IntPointerToInt(src.HibernationDelay),
		// Status no payload field
		// IntanceGroupID no payload field
		// CreatedAt no payload field
		// UpdatedAt no payload field
		// No model field for payload field "project_id"
		// No model field for payload field "zone"
	}
}

func PipelineBaseModelToMediaType(src *model.PipelineBase) *app.PipelineBase {
	if src == nil {
		return nil
	}
	return &app.PipelineBase{
		Name:             src.Name,
		InstanceGroup:    InstanceGroupBodyModelToMediaType(&src.InstanceGroup),
		Container:        PipelineContainerModelToMediaType(&src.Container),
		HibernationDelay: src.HibernationDelay,
		Status:           string(src.Status),
		CreatedAt:        src.CreatedAt,
		UpdatedAt:        src.UpdatedAt,
		// IntanceGroupID no media type field
		// No field for media type field "id"
		// No field for media type field "instance_group_id"
		// No field for media type field "project_id"
		// No field for media type field "zone"
	}
}
