package dto

import (
	"fmt"
	"net/http"
	"strings"
)

type Error struct {
	Reason []string `json:"reason"`
	Code   int      `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[reason: %s, code: %d]", strings.Join(e.Reason, ", "), e.Code)
}

func NewError(reason string) error {
	var err Error
	err.Code = http.StatusInternalServerError
	err.AddReason(reason)
	return &err
}

func NewErrorWithStatus(status int, reason string) error {
	var err Error
	err.Code = status
	err.AddReason(reason)
	return &err
}

func NewErrors(reasons []error) error {
	var err Error
	err.Code = http.StatusInternalServerError
	for _, r := range reasons {
		err.AddReason(r.Error())
	}
	return &err
}

func (d *Error) AddReason(reason string) {
	d.Reason = append(d.Reason, reason)
}
