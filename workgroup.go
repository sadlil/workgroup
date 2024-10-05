// Package workgroup provides a mechanism for synchronizing goroutines,
// propagating errors, and context cancellation signals.
// It is designed for managing collections of goroutines working on
// subtasks of a common task. This package is a fork of the
// errgroup.Group library available in `x/sync`, but with modified
// behavior in how it handles goroutine errors and cancellation.
//
// This package offers two different failure modes:
//
//   - Collect - All goroutines are allowed to complete, and all errors
//     encountered across different goroutines are collected. Wait()
//     returns a joined error that combines all errors from the
//     individual goroutines.
//
//   - FailFast - The first error encountered immediately cancels the
//     context of all remaining goroutines and causes Wait() to return
//     that error.
//
// `workgroup.Group` also provides options to set a retry policy for
// individual goroutines within the group. A zero-value `Group` will
// collect all errors and return them as a single error.
package workgroup

import (
	"context"
	"errors"
	"sync"
)

// FailureMode defines how the workgroup handles errors encountered
// in its goroutines.
type FailureMode int

const (
	// Collect instructs the workgroup to collect all errors from
	// its goroutines and return them as a single error from `Wait()`.
	Collect FailureMode = iota
	// FailFast instructs the workgroup to halt execution and cancel
	// all remaining goroutines upon the first error encountered.
	FailFast
)

// Option is a function that configures a workgroup.
type Option func(*Group)

// WithLimit sets the maximum number of goroutines that can execute
// concurrently within the workgroup.
func WithLimit(n int) Option {
	return func(g *Group) {
		g.sem = make(chan struct{}, n)
	}
}

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero-value `workgroup.Group` is valid and has the following default behavior:
//   - No limit on the number of concurrently executing goroutines.
//   - Does not cancel on error (uses `Collect` failure mode).
//   - Does not retry on error.
type Group struct {
	cancel func()
	wg     sync.WaitGroup

	err     error
	errLock sync.Mutex
	errOnce sync.Once

	sem         chan struct{}
	failureMode FailureMode
}

// New creates a new workgroup with the specified failure mode and options.
// It returns a context that is derived from `ctx`.
// The derived context is canceled when the workgroup finishes
// or is canceled explicitly.
func New(ctx context.Context, mode FailureMode, opts ...Option) (context.Context, *Group) {
	ctx, cancel := context.WithCancel(ctx)

	g := &Group{
		cancel:      cancel,
		failureMode: mode,
	}
	for _, opt := range opts {
		opt(g)
	}
	return ctx, g
}

// Go launches a new goroutine within the workgroup to execute the
// provided function. The function may be retried according to the
// workgroup's retry policy.
// It blocks until the new goroutine can be added without exceeding the
// configured concurrency limit.
func (g *Group) Go(ctx context.Context, fn func() error) {
	g.add()
	go func() {
		defer g.done()

		err := fn()
		if err != nil {
			g.errLock.Lock()
			defer g.errLock.Unlock()

			if g.failureMode == FailFast {
				// In FailFast mode, cancel the workgroup context and
				// store the first error encountered.
				g.errOnce.Do(func() {
					g.err = err
					// Signal cancellation to all goroutines.
					g.Cancel()
				})
				return
			}

			// In Collect mode, aggregate errors from all goroutines.
			g.err = errors.Join(g.err, err)
		}
	}()
}

// Wait blocks until all goroutines in the workgroup have completed.
// It returns nil if all goroutines were successful, or an error
// aggregating the errors encountered, depending on the configured
// failure mode.
func (g *Group) Wait() error {
	g.wg.Wait()
	// Ensure context is canceled after all goroutines finish.
	g.Cancel()
	return g.err
}

// Cancel cancels the workgroup context, signaling all running
// goroutines to stop.
func (g *Group) Cancel() {
	if g.cancel != nil {
		g.cancel()
	}
}

// Len returns the number of goroutines currently active in the workgroup.
func (g *Group) Len() int {
	return len(g.sem)
}

func (g *Group) add() {
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	g.wg.Add(1)
}

func (g *Group) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}
