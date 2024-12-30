package svc

import "context"

// Runnable is an interface for any service that maintains long living task(s)
type Runnable interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// ErrorReporter is an interface for any service that can return errors
type ErrorReporter interface {
	Errors() <-chan error
}
