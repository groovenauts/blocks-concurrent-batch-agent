// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": PipelineBaseOpeningTask Resource Client
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

// RefreshPipelineBaseOpeningTaskPath computes a request path to the refresh action of PipelineBaseOpeningTask.
func RefreshPipelineBaseOpeningTaskPath(id string) string {
	param0 := id

	return fmt.Sprintf("/opening_tasks/%s", param0)
}

// Refresh
func (c *Client) RefreshPipelineBaseOpeningTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewRefreshPipelineBaseOpeningTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewRefreshPipelineBaseOpeningTaskRequest create the request corresponding to the refresh action endpoint of the PipelineBaseOpeningTask resource.
func (c *Client) NewRefreshPipelineBaseOpeningTaskRequest(ctx context.Context, path string) (*http.Request, error) {
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

// StartPipelineBaseOpeningTaskPath computes a request path to the start action of PipelineBaseOpeningTask.
func StartPipelineBaseOpeningTaskPath() string {

	return fmt.Sprintf("/opening_tasks")
}

// Start refreshing
func (c *Client) StartPipelineBaseOpeningTask(ctx context.Context, path string, payload *OperationPayload, id *string, contentType string) (*http.Response, error) {
	req, err := c.NewStartPipelineBaseOpeningTaskRequest(ctx, path, payload, id, contentType)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStartPipelineBaseOpeningTaskRequest create the request corresponding to the start action endpoint of the PipelineBaseOpeningTask resource.
func (c *Client) NewStartPipelineBaseOpeningTaskRequest(ctx context.Context, path string, payload *OperationPayload, id *string, contentType string) (*http.Request, error) {
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
	if id != nil {
		values.Set("id", *id)
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
