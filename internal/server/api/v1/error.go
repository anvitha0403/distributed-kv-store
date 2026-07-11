package v1

import (
	"fmt"

	"google.golang.org/grpc/status"
)

type ErrKeyDoesntExist struct {
}

func (e *ErrKeyDoesntExist) GRPCStatus() *status.Status {
	st := status.New(
		5,
		"key doesnt exist",
	)
	return st
}

func (e *ErrKeyDoesntExist) Error() string {
	return e.GRPCStatus().Err().Error()
}

type InternalServerError struct {
	msg string
}

func (e *InternalServerError) Message(msg string) {
	e.msg = msg
}
func (e *InternalServerError) GRPCStatus() *status.Status {
	st := status.New(
		2,
		fmt.Sprintf("InternalServerError: %s", e.msg),
	)

	return st
}

func (e *InternalServerError) Error() string {
	return e.GRPCStatus().Err().Error()
}
