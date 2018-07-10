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
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// StartPipelineBaseOpeningTaskPath computes a request path to the start action of PipelineBaseOpeningTask.
func StartPipelineBaseOpeningTaskPath(name string) string {
	param0 := name

	return fmt.Sprintf("/pipeline_bases/%s/opening_tasks", param0)
}

// Start operation
func (c *Client) StartPipelineBaseOpeningTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewStartPipelineBaseOpeningTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStartPipelineBaseOpeningTaskRequest create the request corresponding to the start action endpoint of the PipelineBaseOpeningTask resource.
func (c *Client) NewStartPipelineBaseOpeningTaskRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
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

// WatchPipelineBaseOpeningTaskPath computes a request path to the watch action of PipelineBaseOpeningTask.
func WatchPipelineBaseOpeningTaskPath(name string, id string) string {
	param0 := name
	param1 := id

	return fmt.Sprintf("/pipeline_bases/%s/opening_tasks/%s", param0, param1)
}

// Watch
func (c *Client) WatchPipelineBaseOpeningTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewWatchPipelineBaseOpeningTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewWatchPipelineBaseOpeningTaskRequest create the request corresponding to the watch action endpoint of the PipelineBaseOpeningTask resource.
func (c *Client) NewWatchPipelineBaseOpeningTaskRequest(ctx context.Context, path string) (*http.Request, error) {
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
