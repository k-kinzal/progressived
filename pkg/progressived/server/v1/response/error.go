package response

import "fmt"

type Err error

type BadRequestError struct {
	Messsage string `json:"message"`
}

func (e BadRequestError) Error() string {
	return e.Messsage
}

type NotFoundError struct {
	Method string
	Path   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s: not found", e.Method, e.Path)
}

type MethodNotAllowedError struct {
	Method string
	Path   string
}

func (e *MethodNotAllowedError) Error() string {
	return fmt.Sprintf("%s %s: not allowd method", e.Method, e.Path)
}

type InternalServerError struct{ Err }
