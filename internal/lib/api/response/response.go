package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"` // "ok" or "error"
	Error  string `json:"error,omitempty"`
	Info   string `json:"info,omitempty"`
}

const (
	StatusOK            = "200 OK"
	BadRequest          = "400 Bad Request"
	NotFound            = "404 Not Found"
	InternalServerError = "501 Internal Server Error"
	// Forbidden           = "403 Forbidden"
	NotAcceptable = "406 Not Acceptable"
	Unauthorized  = "401 Unauthorized"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Info(msg string) Response {
	return Response{
		Status: StatusOK,
		Info:   msg,
	}
}

func Error(code, msg string) Response {
	return Response{
		Status: code,
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
