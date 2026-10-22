package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestWithRetryRetriableError(t *testing.T) {
	old := retryDelays
	retryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	defer func() { retryDelays = old }()

	calls := 0
	err := withRetry(context.Background(), func() error {
		calls++
		return &pgconn.PgError{Code: pgerrcode.ConnectionFailure}
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 4 {
		t.Errorf("calls = %d, want 4", calls)
	}
}

func TestWithRetryNonRetriableError(t *testing.T) {
	calls := 0
	err := withRetry(context.Background(), func() error {
		calls++
		return &pgconn.PgError{Code: pgerrcode.UniqueViolation}
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestWithRetrySuccessOnLastAttempt(t *testing.T) {
	old := retryDelays
	retryDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	defer func() { retryDelays = old }()

	calls := 0
	err := withRetry(context.Background(), func() error {
		calls++
		if calls < 4 {
			return &pgconn.PgError{Code: pgerrcode.ConnectionFailure}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 4 {
		t.Errorf("calls = %d, want 4", calls)
	}
}
