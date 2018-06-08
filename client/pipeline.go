// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": Pipeline Resource Client
//
// Command:
// $ goagen
// --design=github.com/groovenauts/blocks-concurrent-batch-server/design
// --out=$(GOPATH)/src/github.com/groovenauts/blocks-concurrent-batch-server
// --version=v1.3.1

package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CreatePipelinePath computes a request path to the create action of Pipeline.
func CreatePipelinePath() string {

	return fmt.Sprintf("/pipelines")
}

// create
func (c *Client) CreatePipeline(ctx context.Context, path string, payload *PipelinePayload, orgID string, contentType string) (*http.Response, error) {
	req, err := c.NewCreatePipelineRequest(ctx, path, payload, orgID, contentType)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreatePipelineRequest create the request corresponding to the create action endpoint of the Pipeline resource.
func (c *Client) NewCreatePipelineRequest(ctx context.Context, path string, payload *PipelinePayload, orgID string, contentType string) (*http.Request, error) {
	var body bytes.Buffer
	if contentType == "" {
		contentType = "*/*" // Use default encoder
	}
	err := c.Encoder.Encode(payload, &body, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	values.Set("org_id", orgID)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("POST", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	if contentType == "*/*" {
		header.Set("Content-Type", "application/json")
	} else {
		header.Set("Content-Type", contentType)
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// CurrentPipelinePath computes a request path to the current action of Pipeline.
func CurrentPipelinePath(id string) string {
	param0 := id

	return fmt.Sprintf("/pipelines/%s/current", param0)
}

// Update current pipeline base
func (c *Client) CurrentPipeline(ctx context.Context, path string, pipelineBaseID string) (*http.Response, error) {
	req, err := c.NewCurrentPipelineRequest(ctx, path, pipelineBaseID)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCurrentPipelineRequest create the request corresponding to the current action endpoint of the Pipeline resource.
func (c *Client) NewCurrentPipelineRequest(ctx context.Context, path string, pipelineBaseID string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	values.Set("pipeline_base_id", pipelineBaseID)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// DeletePipelinePath computes a request path to the delete action of Pipeline.
func DeletePipelinePath(id string) string {
	param0 := id

	return fmt.Sprintf("/pipelines/%s", param0)
}

// delete
func (c *Client) DeletePipeline(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewDeletePipelineRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeletePipelineRequest create the request corresponding to the delete action endpoint of the Pipeline resource.
func (c *Client) NewDeletePipelineRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// ListPipelinePath computes a request path to the list action of Pipeline.
func ListPipelinePath() string {

	return fmt.Sprintf("/pipelines")
}

// list
func (c *Client) ListPipeline(ctx context.Context, path string, orgID string) (*http.Response, error) {
	req, err := c.NewListPipelineRequest(ctx, path, orgID)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListPipelineRequest create the request corresponding to the list action endpoint of the Pipeline resource.
func (c *Client) NewListPipelineRequest(ctx context.Context, path string, orgID string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	values.Set("org_id", orgID)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// PreparingFinalizeTaskPipelinePath computes a request path to the preparing_finalize_task action of Pipeline.
func PreparingFinalizeTaskPipelinePath(id string) string {
	param0 := id

	return fmt.Sprintf("/pipelines/%s/preparing_finalize_task", param0)
}

// Task to finalize current_preparing or next_preparing status
func (c *Client) PreparingFinalizeTaskPipeline(ctx context.Context, path string, error *string, operationID *string) (*http.Response, error) {
	req, err := c.NewPreparingFinalizeTaskPipelineRequest(ctx, path, error, operationID)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewPreparingFinalizeTaskPipelineRequest create the request corresponding to the preparing_finalize_task action endpoint of the Pipeline resource.
func (c *Client) NewPreparingFinalizeTaskPipelineRequest(ctx context.Context, path string, error *string, operationID *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if error != nil {
		values.Set("error", *error)
	}
	if operationID != nil {
		values.Set("operation_id", *operationID)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// ShowPipelinePath computes a request path to the show action of Pipeline.
func ShowPipelinePath(id string) string {
	param0 := id

	return fmt.Sprintf("/pipelines/%s", param0)
}

// show
func (c *Client) ShowPipeline(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowPipelineRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowPipelineRequest create the request corresponding to the show action endpoint of the Pipeline resource.
func (c *Client) NewShowPipelineRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// StopPipelinePath computes a request path to the stop action of Pipeline.
func StopPipelinePath(id string) string {
	param0 := id

	return fmt.Sprintf("/pipelines/%s/stop", param0)
}

// Stop pipeline
func (c *Client) StopPipeline(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewStopPipelineRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStopPipelineRequest create the request corresponding to the stop action endpoint of the Pipeline resource.
func (c *Client) NewStopPipelineRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.APIKeySigner != nil {
		if err := c.APIKeySigner.Sign(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}
