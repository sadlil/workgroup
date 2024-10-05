package workgroup

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var (
	errInternal = errors.New("internal")
	errInvalid  = errors.New("invalid")
)

func TestWorkGroup_Collect(t *testing.T) {
	var (
		count       int32
		expectedErr = errors.New("Collect")
	)

	ctx, group := New(context.Background(), Collect)
	for i := 0; i < 10; i++ {
		group.Go(ctx, func() error {
			atomic.AddInt32(&count, 1)
			return expectedErr
		})
	}
	err := group.Wait()
	if err == nil {
		t.Fatal("g.Wait() = nil, want error")
	}
	if count != 10 {
		t.Errorf("expected 10 goroutines to run, but got %d", count)
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to contain %q, but got %q", expectedErr, err)
	}
}

func TestWorkGroup_Collect_MultipleErrors(t *testing.T) {
	ctx, group := New(context.Background(), Collect)

	// Start two goroutines with errors.
	group.Go(ctx, func() error {
		return fmt.Errorf("error 1: %w", errInternal)
	})
	group.Go(ctx, func() error {
		return fmt.Errorf("error 2: %w", errInvalid)
	})

	err := group.Wait()
	if err == nil {
		t.Fatal("g.Wait() = nil, want error")
	}
	if !errors.Is(err, errInternal) {
		t.Errorf("errors.Is(err, errInternal) = false, want true")
	}
	if !errors.Is(err, errInvalid) {
		t.Errorf("errors.Is(err, errInvalid) = false, want true")
	}
}

func TestWorkGroup_FailFast(t *testing.T) {
	var (
		count       int32
		expectedErr = errors.New("FailFast")
	)

	ctx, g := New(context.Background(), FailFast)
	for i := 0; i < 10; i++ {
		g.Go(ctx, func() error {
			select {
			case <-time.After(time.Duration(i) * time.Second):
				atomic.AddInt32(&count, 1)
				return fmt.Errorf("error %d: %w", i, expectedErr)
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}
	err := g.Wait()
	if err == nil {
		t.Fatal("group.Wait() = nil, want error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("errors.Is(err, expectedErr) = false, want true")
	}
	if count == 10 {
		t.Error("expected some goroutines to be cancelled, but all ran")
	}
}

// TestWorkgroup_FailFastMode tests that the workgroup stops on the first error in FailFast mode.
func TestWorkgroup_FailFast_SingleError(t *testing.T) {
	ctx := context.Background()
	ctx, group := New(ctx, FailFast)

	// Start two goroutines with errors.
	// Both of them will finish as no handler for ctx.Done() is set.
	group.Go(ctx, func() error {
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("error 1: %w", errInternal)
	})
	group.Go(ctx, func() error {
		return fmt.Errorf("error 2: %w", errInvalid)
	})

	err := group.Wait()
	if err == nil {
		t.Fatal("group.Wait() = nil, want error")
	}

	if !errors.Is(err, errInvalid) {
		t.Errorf("errors.Is(err, status.ErrInvalidArgument) = false, want true")
	}
	if errors.Is(err, errInternal) {
		t.Errorf("errors.Is(err, status.ErrInternal) = true, want false")
	}
}

func TestWorkGroup_NoError(t *testing.T) {
	ctx := context.Background()
	ctx, g := New(ctx, Collect)

	// Start two successful goroutines.
	g.Go(ctx, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	g.Go(ctx, func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	err := g.Wait()
	if err != nil {
		t.Fatalf("group.Wait() = %v, want nil", err)
	}
}

func TestGroup_WithLimit(t *testing.T) {
	var (
		current int32
		max     int32
	)

	ctx, g := New(context.Background(), Collect, WithLimit(3))
	for i := 0; i < 10; i++ {
		g.Go(ctx, func() error {
			c := atomic.AddInt32(&current, 1)
			if c > max {
				atomic.StoreInt32(&max, c)
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&current, -1)
			return nil
		})
	}
	err := g.Wait()
	if err != nil {
		t.Fatalf("group.Wait() = %v, want nil", err)
	}
	if max != 3 {
		t.Errorf("expected maximum 3 concurrent goroutines, but got %d", max)
	}
}

func TestGroup_Cancel(t *testing.T) {
	ctx, g := New(context.Background(), Collect)
	var count int32
	for i := 0; i < 10; i++ {
		g.Go(ctx, func() error {
			select {
			case <-time.After(time.Duration(i) * time.Millisecond):
				atomic.AddInt32(&count, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}
	// Allow some goroutines to run
	time.Sleep(2 * time.Millisecond)
	g.Cancel()

	err := g.Wait()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected context.Canceled, but got %v", err)
	}
	if count == 10 {
		t.Error("expected some goroutines to be cancelled, but all ran")
	}
}

func BenchmarkGo(b *testing.B) {
	ctx, g := New(context.Background(), Collect)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g.Go(ctx, func() error { return nil })
	}
	g.Wait()
}
