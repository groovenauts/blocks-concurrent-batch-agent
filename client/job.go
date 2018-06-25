// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": Job Resource Client
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

// ActivateJobPath computes a request path to the activate action of Job.
func ActivateJobPath(id string) string {
	param0 := id

	return fmt.Sprintf("/jobs/%s/activate", param0)
}

// Activate job
func (c *Client) ActivateJob(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewActivateJobRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewActivateJobRequest create the request corresponding to the activate action endpoint of the Job resource.
func (c *Client) NewActivateJobRequest(ctx context.Context, path string) (*http.Request, error) {
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

// CreateJobPath computes a request path to the create action of Job.
func CreateJobPath() string {

	return fmt.Sprintf("/jobs")
}

// create
func (c *Client) CreateJob(ctx context.Context, path string, payload *JobPayload, active *string, pipelineBaseID *string, pipelineID *string, contentType string) (*http.Response, error) {
	req, err := c.NewCreateJobRequest(ctx, path, payload, active, pipelineBaseID, pipelineID, contentType)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreateJobRequest create the request corresponding to the create action endpoint of the Job resource.
func (c *Client) NewCreateJobRequest(ctx context.Context, path string, payload *JobPayload, active *string, pipelineBaseID *string, pipelineID *string, contentType string) (*http.Request, error) {
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
	if active != nil {
		values.Set("active", *active)
	}
	if pipelineBaseID != nil {
		values.Set("pipeline_base_id", *pipelineBaseID)
	}
	if pipelineID != nil {
		values.Set("pipeline_id", *pipelineID)
	}
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

// DeleteJobPath computes a request path to the delete action of Job.
func DeleteJobPath(id string) string {
	param0 := id

	return fmt.Sprintf("/jobs/%s", param0)
}

// delete
func (c *Client) DeleteJob(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewDeleteJobRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeleteJobRequest create the request corresponding to the delete action endpoint of the Job resource.
func (c *Client) NewDeleteJobRequest(ctx context.Context, path string) (*http.Request, error) {
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

// InactivateJobPath computes a request path to the inactivate action of Job.
func InactivateJobPath(id string) string {
	param0 := id

	return fmt.Sprintf("/jobs/%s/inactivate", param0)
}

// Inactivate job
func (c *Client) InactivateJob(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewInactivateJobRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewInactivateJobRequest create the request corresponding to the inactivate action endpoint of the Job resource.
func (c *Client) NewInactivateJobRequest(ctx context.Context, path string) (*http.Request, error) {
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

// PublishingTaskJobPath computes a request path to the publishing_task action of Job.
func PublishingTaskJobPath(id string) string {
	param0 := id

	return fmt.Sprintf("/jobs/%s/publishing_task", param0)
}

// Publishing job task
func (c *Client) PublishingTaskJob(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewPublishingTaskJobRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewPublishingTaskJobRequest create the request corresponding to the publishing_task action endpoint of the Job resource.
func (c *Client) NewPublishingTaskJobRequest(ctx context.Context, path string) (*http.Request, error) {
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

// ShowJobPath computes a request path to the show action of Job.
func ShowJobPath(id string) string {
	param0 := id

	return fmt.Sprintf("/jobs/%s", param0)
}

// show
func (c *Client) ShowJob(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowJobRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowJobRequest create the request corresponding to the show action endpoint of the Job resource.
func (c *Client) NewShowJobRequest(ctx context.Context, path string) (*http.Request, error) {
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
