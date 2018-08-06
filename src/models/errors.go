package models

import (
	"errors"
	"fmt"
)

type InvalidParent struct {
	ID string
}

func (e *InvalidParent) Error() string {
	return fmt.Sprintf("Invalid parent from ID: %q", e.ID)
}

type InvalidReference struct {
	ID string
}

func (e *InvalidReference) Error() string {
	return fmt.Sprintf("Invalid reference from ID: %q", e.ID)
}

var ErrTimeout = errors.New("Timeout")

type InvalidOperation struct {
	Msg string
}

func (e *InvalidOperation) Error() string {
	return e.Msg
}

type InvalidStateTransition struct {
	Msg string
}

func (e *InvalidStateTransition) Error() string {
	return e.Msg
}

type SubscriprionNotFound struct {
	Subscription string
}

func (e *SubscriprionNotFound) Error() string {
	return fmt.Sprintf("%q not found", e.Subscription)
}
