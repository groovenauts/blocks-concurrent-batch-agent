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

var ErrTimeout = errors.New("Timeout")
