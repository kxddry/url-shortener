package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"` // "ok" or "error"
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK            = "OK"
	BadRequest          = "Bad Request"
	NotFound            = "Not Found"
	InternalServerError = "Internal Server Error"
	Forbidden           = "Forbidden"
	NotAcceptable       = "Not Acceptable"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(code, msg string) Response {
	var StatusError string
	switch code {
	case "400":
		StatusError = BadRequest
	case "404":
		StatusError = NotFound
	case "500":
		StatusError = InternalServerError
	case "403":
		StatusError = Forbidden
	default:
		StatusError = NotAcceptable
	}
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: BadRequest,
		Error:  strings.Join(errMsgs, ", "),
	}
}
