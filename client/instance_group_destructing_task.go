// Code generated by goagen v1.3.1, DO NOT EDIT.
//
// API "appengine": InstanceGroupDestructingTask Resource Client
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

// RefreshInstanceGroupDestructingTaskPath computes a request path to the refresh action of InstanceGroupDestructingTask.
func RefreshInstanceGroupDestructingTaskPath(id string) string {
	param0 := id

	return fmt.Sprintf("/destructing_tasks/%s", param0)
}

// Refresh
func (c *Client) RefreshInstanceGroupDestructingTask(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewRefreshInstanceGroupDestructingTaskRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewRefreshInstanceGroupDestructingTaskRequest create the request corresponding to the refresh action endpoint of the InstanceGroupDestructingTask resource.
func (c *Client) NewRefreshInstanceGroupDestructingTaskRequest(ctx context.Context, path string) (*http.Request, error) {
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
