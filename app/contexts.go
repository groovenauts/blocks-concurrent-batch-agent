// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": Application Contexts
//
// Command:
// $ goagen
// --design=github.com/groovenauts/blocks-concurrent-batch-server/design
// --out=$(GOPATH)/src/github.com/groovenauts/blocks-concurrent-batch-server
// --version=v1.3.1

package app

import (
	"context"
	"github.com/goadesign/goa"
	"net/http"
)

// CreateIntanceGroupContext provides the IntanceGroup create action context.
type CreateIntanceGroupContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	Payload *InstanceGroupPayload
}

// NewCreateIntanceGroupContext parses the incoming request URL and body, performs validations and creates the
// context used by the IntanceGroup controller create action.
func NewCreateIntanceGroupContext(ctx context.Context, r *http.Request, service *goa.Service) (*CreateIntanceGroupContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := CreateIntanceGroupContext{Context: ctx, ResponseData: resp, RequestData: req}
	return &rctx, err
}

// Created sends a HTTP response with status code 201.
func (ctx *CreateIntanceGroupContext) Created(r *InstanceGroup) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.instance-group+json")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 201, r)
}

// BadRequest sends a HTTP response with status code 400.
func (ctx *CreateIntanceGroupContext) BadRequest(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 400, r)
}

// Unauthorized sends a HTTP response with status code 401.
func (ctx *CreateIntanceGroupContext) Unauthorized(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 401, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *CreateIntanceGroupContext) NotFound(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 404, r)
}

// InternalServerError sends a HTTP response with status code 500.
func (ctx *CreateIntanceGroupContext) InternalServerError(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 500, r)
}

// DeleteIntanceGroupContext provides the IntanceGroup delete action context.
type DeleteIntanceGroupContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	ID string
}

// NewDeleteIntanceGroupContext parses the incoming request URL and body, performs validations and creates the
// context used by the IntanceGroup controller delete action.
func NewDeleteIntanceGroupContext(ctx context.Context, r *http.Request, service *goa.Service) (*DeleteIntanceGroupContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := DeleteIntanceGroupContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramID := req.Params["id"]
	if len(paramID) > 0 {
		rawID := paramID[0]
		rctx.ID = rawID
	}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *DeleteIntanceGroupContext) OK(r *InstanceGroup) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.instance-group+json")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// BadRequest sends a HTTP response with status code 400.
func (ctx *DeleteIntanceGroupContext) BadRequest(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 400, r)
}

// Unauthorized sends a HTTP response with status code 401.
func (ctx *DeleteIntanceGroupContext) Unauthorized(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 401, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *DeleteIntanceGroupContext) NotFound(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 404, r)
}

// InternalServerError sends a HTTP response with status code 500.
func (ctx *DeleteIntanceGroupContext) InternalServerError(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 500, r)
}

// DestructIntanceGroupContext provides the IntanceGroup destruct action context.
type DestructIntanceGroupContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	ID string
}

// NewDestructIntanceGroupContext parses the incoming request URL and body, performs validations and creates the
// context used by the IntanceGroup controller destruct action.
func NewDestructIntanceGroupContext(ctx context.Context, r *http.Request, service *goa.Service) (*DestructIntanceGroupContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := DestructIntanceGroupContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramID := req.Params["id"]
	if len(paramID) > 0 {
		rawID := paramID[0]
		rctx.ID = rawID
	}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *DestructIntanceGroupContext) OK(r *InstanceGroup) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.instance-group+json")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// BadRequest sends a HTTP response with status code 400.
func (ctx *DestructIntanceGroupContext) BadRequest(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 400, r)
}

// Unauthorized sends a HTTP response with status code 401.
func (ctx *DestructIntanceGroupContext) Unauthorized(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 401, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *DestructIntanceGroupContext) NotFound(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 404, r)
}

// InternalServerError sends a HTTP response with status code 500.
func (ctx *DestructIntanceGroupContext) InternalServerError(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 500, r)
}

// ListIntanceGroupContext provides the IntanceGroup list action context.
type ListIntanceGroupContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
}

// NewListIntanceGroupContext parses the incoming request URL and body, performs validations and creates the
// context used by the IntanceGroup controller list action.
func NewListIntanceGroupContext(ctx context.Context, r *http.Request, service *goa.Service) (*ListIntanceGroupContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ListIntanceGroupContext{Context: ctx, ResponseData: resp, RequestData: req}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ListIntanceGroupContext) OK(r InstanceGroupCollection) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.instance-group+json; type=collection")
	}
	if r == nil {
		r = InstanceGroupCollection{}
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// BadRequest sends a HTTP response with status code 400.
func (ctx *ListIntanceGroupContext) BadRequest(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 400, r)
}

// Unauthorized sends a HTTP response with status code 401.
func (ctx *ListIntanceGroupContext) Unauthorized(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 401, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ListIntanceGroupContext) NotFound(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 404, r)
}

// InternalServerError sends a HTTP response with status code 500.
func (ctx *ListIntanceGroupContext) InternalServerError(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 500, r)
}

// ResizeIntanceGroupContext provides the IntanceGroup resize action context.
type ResizeIntanceGroupContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	ID      string
	Payload *ResizeIntanceGroupPayload
}

// NewResizeIntanceGroupContext parses the incoming request URL and body, performs validations and creates the
// context used by the IntanceGroup controller resize action.
func NewResizeIntanceGroupContext(ctx context.Context, r *http.Request, service *goa.Service) (*ResizeIntanceGroupContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ResizeIntanceGroupContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramID := req.Params["id"]
	if len(paramID) > 0 {
		rawID := paramID[0]
		rctx.ID = rawID
	}
	return &rctx, err
}

// resizeIntanceGroupPayload is the IntanceGroup resize action payload.
type resizeIntanceGroupPayload struct {
	// New Instance Size
	NewSize *int `form:"new_size,omitempty" json:"new_size,omitempty" yaml:"new_size,omitempty" xml:"new_size,omitempty"`
}

// Validate runs the validation rules defined in the design.
func (payload *resizeIntanceGroupPayload) Validate() (err error) {
	if payload.NewSize == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`raw`, "new_size"))
	}
	return
}

// Publicize creates ResizeIntanceGroupPayload from resizeIntanceGroupPayload
func (payload *resizeIntanceGroupPayload) Publicize() *ResizeIntanceGroupPayload {
	var pub ResizeIntanceGroupPayload
	if payload.NewSize != nil {
		pub.NewSize = *payload.NewSize
	}
	return &pub
}

// ResizeIntanceGroupPayload is the IntanceGroup resize action payload.
type ResizeIntanceGroupPayload struct {
	// New Instance Size
	NewSize int `form:"new_size" json:"new_size" yaml:"new_size" xml:"new_size"`
}

// OK sends a HTTP response with status code 200.
func (ctx *ResizeIntanceGroupContext) OK(r *InstanceGroup) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.instance-group+json")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// BadRequest sends a HTTP response with status code 400.
func (ctx *ResizeIntanceGroupContext) BadRequest(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 400, r)
}

// Unauthorized sends a HTTP response with status code 401.
func (ctx *ResizeIntanceGroupContext) Unauthorized(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 401, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ResizeIntanceGroupContext) NotFound(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 404, r)
}

// InternalServerError sends a HTTP response with status code 500.
func (ctx *ResizeIntanceGroupContext) InternalServerError(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 500, r)
}

// ShowIntanceGroupContext provides the IntanceGroup show action context.
type ShowIntanceGroupContext struct {
	context.Context
	*goa.ResponseData
	*goa.RequestData
	ID string
}

// NewShowIntanceGroupContext parses the incoming request URL and body, performs validations and creates the
// context used by the IntanceGroup controller show action.
func NewShowIntanceGroupContext(ctx context.Context, r *http.Request, service *goa.Service) (*ShowIntanceGroupContext, error) {
	var err error
	resp := goa.ContextResponse(ctx)
	resp.Service = service
	req := goa.ContextRequest(ctx)
	req.Request = r
	rctx := ShowIntanceGroupContext{Context: ctx, ResponseData: resp, RequestData: req}
	paramID := req.Params["id"]
	if len(paramID) > 0 {
		rawID := paramID[0]
		rctx.ID = rawID
	}
	return &rctx, err
}

// OK sends a HTTP response with status code 200.
func (ctx *ShowIntanceGroupContext) OK(r *InstanceGroup) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.instance-group+json")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 200, r)
}

// BadRequest sends a HTTP response with status code 400.
func (ctx *ShowIntanceGroupContext) BadRequest(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 400, r)
}

// Unauthorized sends a HTTP response with status code 401.
func (ctx *ShowIntanceGroupContext) Unauthorized(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 401, r)
}

// NotFound sends a HTTP response with status code 404.
func (ctx *ShowIntanceGroupContext) NotFound(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 404, r)
}

// InternalServerError sends a HTTP response with status code 500.
func (ctx *ShowIntanceGroupContext) InternalServerError(r error) error {
	if ctx.ResponseData.Header().Get("Content-Type") == "" {
		ctx.ResponseData.Header().Set("Content-Type", "application/vnd.goa.error")
	}
	return ctx.ResponseData.Service.Send(ctx.Context, 500, r)
}
