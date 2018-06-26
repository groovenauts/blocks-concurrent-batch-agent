package model

import (
	"regexp"

	"gopkg.in/go-playground/validator.v9"
)

const (
	Ubuntu16ImageFamily = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/family/ubuntu-1604-lts"
)

var ImagesRecommendedForGpu = []string{
	Ubuntu16ImageFamily,
}
var Ubuntu1604Regexp = regexp.MustCompile(`ubuntu.*1604`)

func InstanceGroupStructLevelValidation(sl validator.StructLevel) {
	pl := sl.Current().Interface().(InstanceGroup)
	bd := pl.BootDisk
	if pl.GpuAccelerators.Count > 0 {
		if !Ubuntu1604Regexp.MatchString(bd.SourceImage) {
			sl.ReportError(bd.SourceImage, "SourceImage", "", "source_image", "Invalid Image for GPU")
		}
	}
}

func (m *InstanceGroup) Validate() error {
	validator := validator.New()
	validator.RegisterStructValidation(InstanceGroupStructLevelValidation, InstanceGroup{})
	err := validator.Struct(m)
	return err
}
