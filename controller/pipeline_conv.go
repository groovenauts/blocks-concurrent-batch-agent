package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func PipelinePayloadToModel(src *app.PipelinePayload) model.Pipeline {
	if src == nil {
		return model.Pipeline{}
	}
	return model.Pipeline{
		Name:             src.Name,
		InstanceGroup:    InstanceGroupBodyPayloadToModel(src.InstanceGroup),
		Container:        PipelineContainerPayloadToModel(src.Container),
		HibernationDelay: IntPointerToInt(src.HibernationDelay),
		// Status no payload field
		// IntanceGroupID no payload field
		// CreatedAt no payload field
		// UpdatedAt no payload field
	}
}

func PipelineModelToMediaType(src *model.Pipeline) *app.Pipeline {
	if src == nil {
		return nil
	}
	return &app.Pipeline{
		Name:             src.Name,
		InstanceGroup:    InstanceGroupBodyModelToMediaType(&src.InstanceGroup),
		Container:        PipelineContainerModelToMediaType(&src.Container),
		HibernationDelay: src.HibernationDelay,
		Status:           string(src.Status),
		CreatedAt:        src.CreatedAt,
		UpdatedAt:        src.UpdatedAt,
		// IntanceGroupID no media type field
		// No field for media type field "curr_base_id"
		// No field for media type field "id"
		// No field for media type field "next_base_id"
		// No field for media type field "prev_base_id"
	}
}
