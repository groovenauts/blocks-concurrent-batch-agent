package controller

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/taskqueue"
)

func PostTask(c context.Context, path string, delay time.Duration) error {
	return RequestTask(c, "POST", path, delay)
}

func PutTask(c context.Context, path string, delay time.Duration) error {
	return RequestTask(c, "PUT", path, delay)
}

func RequestTask(c context.Context, method string, path string, delay time.Duration) error {
	task := &taskqueue.Task{
		Method: method,
		Path:   path,
		ETA:    time.Now().Add(delay),
	}
	if _, err := taskqueue.Add(c, task, ""); err != nil {
		return err
	}
	return nil
}
