// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": Application Resource Href Factories
//
// Command:
// $ goagen
// --design=github.com/groovenauts/blocks-concurrent-batch-server/design
// --out=$(GOPATH)/src/github.com/groovenauts/blocks-concurrent-batch-server
// --version=v1.3.1

package app

import (
	"fmt"
	"strings"
)

// InstanceGroupHref returns the resource href.
func InstanceGroupHref(name interface{}) string {
	paramname := strings.TrimLeftFunc(fmt.Sprintf("%v", name), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/instance_groups/%v", paramname)
}

// JobHref returns the resource href.
func JobHref(id interface{}) string {
	paramid := strings.TrimLeftFunc(fmt.Sprintf("%v", id), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/jobs/%v", paramid)
}

// PipelineHref returns the resource href.
func PipelineHref(name interface{}) string {
	paramname := strings.TrimLeftFunc(fmt.Sprintf("%v", name), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/pipelines/%v", paramname)
}

// PipelineBaseHref returns the resource href.
func PipelineBaseHref(name interface{}) string {
	paramname := strings.TrimLeftFunc(fmt.Sprintf("%v", name), func(r rune) bool { return r == '/' })
	return fmt.Sprintf("/pipeline_bases/%v", paramname)
}
