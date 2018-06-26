package model

import (
	"gopkg.in/go-playground/validator.v9"
)

func (m *PipelineBase) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}
