// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": PipelineBaseClosingTask Resource Client
//
// Command:
// $ goagen
// --design=github.com/groovenauts/blocks-concurrent-batch-server/design
// --out=$(GOPATH)/src/github.com/groovenauts/blocks-concurrent-batch-server
// --version=v1.3.1

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// StartPipelineBaseClosingTaskPath computes a request path to the start action of PipelineBaseClosingTask.
func StartPipelineBaseClosingTaskPath() string {

	return fmt.Sprintf("/closing_tasks")
}

// Start operation
func (c *Client) StartPipelineBaseClosingTask(ctx context.Context, path string, resourceID string) (*http.Response, error) {
	req, err := c.NewStartPipelineBaseClosingTaskRequest(ctx, path, resourceID)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStartPipelineBaseClosingTaskRequest create the request corresponding to the start action endpoint of the PipelineBaseClosingTask resource.
func (c *Client) NewStartPipelineBaseClosingTaskRequest(ctx context.Context, path string, resourceID string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	values.Set("resource_id", resourceID)
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("POST", u.String(), nil)
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

// WatchPipelineBaseClosingTaskPath computes a request path to the watch action of PipelineBaseClosingTask.
func WatchPipelineBaseClosingTaskPath(id string) string {
	param0 := id

	return fmt.Sprintf("/closing_tasks/%s", param0)
}

// Watch
func (c *Client) WatchPipelineBaseClosingTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewWatchPipelineBaseClosingTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewWatchPipelineBaseClosingTaskRequest create the request corresponding to the watch action endpoint of the PipelineBaseClosingTask resource.
func (c *Client) NewWatchPipelineBaseClosingTaskRequest(ctx context.Context, path string) (*http.Request, error) {
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