package pipeline

import (
	"net/http"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Pipeline represents a set of Middleware to be applied, and which will be applied in the reverse order
// they are specified in the slice
type Pipeline []Middleware

// PipeHandler is a wrapper allowing us to use method calls for chaining the primary Pipe operation
type PipeHandler struct {
	h     http.HandlerFunc
	pipes []Pipeline
}

// Build takes a set of middleware functions and creates a pipeline with the functions
// sorted in the reverse of the order they were passed. Since the middleware will then be applied in
// reverse pipeline order, this lets the user of the package specify them in a more natural way, e.g.
//
// Build(withHostname, mustAuthenticate)
//
// will first fire the withHostname middleware (for logging, etc.) and then the authentication middleware
// to actually allow or deny the request.
func Build(ms ...Middleware) Pipeline {
	n := len(ms)
	p := make(Pipeline, n, n)

	// We apply the middleware in reverse order so that the package user can specify them in
	// the natural ordering
	for i := range ms {
		p[i] = ms[n-i-1]
	}

	return p
}

// Start is a convenience function to create a PipeHandler type that we can call methods on allowing us to chain
// calls to Pipe.
func Start(h http.HandlerFunc) PipeHandler {
	return PipeHandler{h: h}
}

// Pipe adds the specified Pipeline to the list of Pipelines to process  when creating the final http.HandlerFunc
func (ph PipeHandler) Pipe(ps Pipeline) PipeHandler {
	ph.pipes = append(ph.pipes, ps)
	return ph
}

func (ph PipeHandler) process() http.HandlerFunc {
	n := len(ph.pipes)

	// We step through the pipes that were added in reverse order so they are applied in the order
	// specified by the package user
	for i := range ph.pipes {
		pl := ph.pipes[n-i-1]

		for _, p := range pl {
			ph.h = p(ph.h)
		}
	}

	return ph.h
}

// Handler processes all of the pipes and returns the total object that results, and which can then be
// used with http.HandleFunc, etc.
func (ph PipeHandler) Handler() http.HandlerFunc {
	return ph.process()
}
