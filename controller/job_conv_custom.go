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
