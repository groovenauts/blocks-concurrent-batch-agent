// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": InstanceGroupResizingTask Resource Client
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

// StartInstanceGroupResizingTaskPath computes a request path to the start action of InstanceGroupResizingTask.
func StartInstanceGroupResizingTaskPath(name string) string {
	param0 := name

	return fmt.Sprintf("/instance_groups/%s/resizing_tasks", param0)
}

// Start operation
func (c *Client) StartInstanceGroupResizingTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewStartInstanceGroupResizingTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStartInstanceGroupResizingTaskRequest create the request corresponding to the start action endpoint of the InstanceGroupResizingTask resource.
func (c *Client) NewStartInstanceGroupResizingTaskRequest(ctx context.Context, path string) (*http.Request, error) {
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

// WatchInstanceGroupResizingTaskPath computes a request path to the watch action of InstanceGroupResizingTask.
func WatchInstanceGroupResizingTaskPath(name string, id string) string {
	param0 := name
	param1 := id

	return fmt.Sprintf("/instance_groups/%s/resizing_tasks/%s", param0, param1)
}

// Watch
func (c *Client) WatchInstanceGroupResizingTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewWatchInstanceGroupResizingTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewWatchInstanceGroupResizingTaskRequest create the request corresponding to the watch action endpoint of the InstanceGroupResizingTask resource.
func (c *Client) NewWatchInstanceGroupResizingTaskRequest(ctx context.Context, path string) (*http.Request, error) {
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
