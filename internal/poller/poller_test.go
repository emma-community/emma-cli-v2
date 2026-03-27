package poller

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWaitFor_ImmediateDone(t *testing.T) {
	calls := 0
	err := WaitFor(context.Background(), 10*time.Millisecond, 1*time.Second, func() (bool, error) {
		calls++
		return true, nil
	})
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWaitFor_EventuallyDone(t *testing.T) {
	calls := 0
	err := WaitFor(context.Background(), 10*time.Millisecond, 2*time.Second, func() (bool, error) {
		calls++
		return calls >= 3, nil
	})
	if err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
	if calls < 3 {
		t.Errorf("expected at least 3 calls, got %d", calls)
	}
}

func TestWaitFor_Timeout(t *testing.T) {
	err := WaitFor(context.Background(), 10*time.Millisecond, 50*time.Millisecond, func() (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWaitFor_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	done := make(chan error, 1)
	go func() {
		done <- WaitFor(ctx, 10*time.Millisecond, 10*time.Second, func() (bool, error) {
			calls++
			return false, nil
		})
	}()

	// Cancel after a brief delay
	time.Sleep(30 * time.Millisecond)
	cancel()

	err := <-done
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
}

func TestWaitFor_CheckReturnsError(t *testing.T) {
	checkErr := fmt.Errorf("check failed with internal error")
	err := WaitFor(context.Background(), 10*time.Millisecond, 1*time.Second, func() (bool, error) {
		return false, checkErr
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
