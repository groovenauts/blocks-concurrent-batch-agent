package controller

import (
	"github.com/groovenauts/blocks-concurrent-batch-server/app"
	"github.com/groovenauts/blocks-concurrent-batch-server/model"
)

func JobMessagePayloadToModel(src *app.JobMessage) model.JobMessage {
	if src == nil {
		return model.JobMessage{}
	}
	return model.JobMessage{
		Data: StringPointerToString(src.Data),
		// AttributeEntries no payload field
		// No model field for payload field "attributes"
	}
}

func JobMessageModelToMediaType(src *model.JobMessage) *app.JobMessage {
	if src == nil {
		return nil
	}
	return &app.JobMessage{
		Data: &src.Data,
		// AttributeEntries no media type field
		// No field for media type field "attributes"
	}
}

func JobPayloadToModel(src *app.JobPayload) model.Job {
	if src == nil {
		return model.Job{}
	}
	return model.Job{
		IDByClient: StringPointerToString(src.IDByClient),
		Message:    JobMessagePayloadToModel(src.Message),
		// Status no payload field
		// Zone no payload field
		// Hostname no payload field
		// MessageID no payload field
		// Output no payload field
		// PipelineId no payload field
		// PipelineBaseId no payload field
		// PublishedAt no payload field
		// StartedAt no payload field
		// FinishedAt no payload field
		// CreatedAt no payload field
		// UpdatedAt no payload field
	}
}

func JobModelToMediaType(src *model.Job) *app.Job {
	if src == nil {
		return nil
	}
	return &app.Job{
		Status:         string(src.Status),
		Hostname:       &src.Hostname,
		Message:        JobMessageModelToMediaType(&src.Message),
		PipelineID:     &src.PipelineId,
		PipelineBaseID: src.PipelineBaseId,
		PublishedAt:    &src.PublishedAt,
		StartedAt:      &src.StartedAt,
		FinishedAt:     &src.FinishedAt,
		CreatedAt:      &src.CreatedAt,
		UpdatedAt:      &src.UpdatedAt,
		// IDByClient no media type field
		// Zone no media type field
		// MessageID no media type field
		// Output no media type field
		// No field for media type field "id"
	}
}
