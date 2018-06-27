package controller

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/taskqueue"
)

func PutTask(c context.Context, path string, delay time.Duration) error {
	task := &taskqueue.Task{
		Method: "PUT",
		Path: path,
		ETA: time.Now().Add(delay),
	}
	if _, err := taskqueue.Add(c, task, ""); err != nil {
		return err
	}
	return nil
}
