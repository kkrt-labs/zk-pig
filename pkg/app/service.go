package app

import (
	"context"
)

// Runnable is an interface for any service that maintains long living task(s)
type Runnable interface {
	// Start long living task(s)
	// It SHOULD return an error if the service can not start successfully
	// In case the context is canceled or times out, the service SHOULD return an error ASAP
	//
	// App ensures that Start is called only once
	// App ensures that all service's dependencies have been successfully started before calling Start
	Start(context.Context) error

	// Stop long living task(s)
	// It SHOULD attempt to gracefully stop and clean its internal state and return an error if it can not do so
	// In case the context is canceled or times out, the service should return an error ASAP
	//
	// App ensures that Stop is called only once
	Stop(context.Context) error
}

// Checkable is a service that can expose its health status
type Checkable interface {
	// Ready should return nil if the service is ready to accept traffic
	// Otherwise, it should return an error
	//
	// Ready is called by the App only when the service is Running (after successful Start() and before calling Stop())
	Ready(ctx context.Context) error
}
