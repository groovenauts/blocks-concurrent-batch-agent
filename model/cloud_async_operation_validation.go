package model

import (
	"gopkg.in/go-playground/validator.v9"
)

func (m *InstanceGroupOperation) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}

func (m *PipelineBaseOperation) Validate() error {
	validator := validator.New()
	err := validator.Struct(m)
	return err
}
