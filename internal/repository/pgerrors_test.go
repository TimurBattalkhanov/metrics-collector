package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want pgErrorClassification
	}{
		{"nil", nil, nonRetriable},
		{"connection exception", &pgconn.PgError{Code: pgerrcode.ConnectionException}, retriable},
		{"connection failure", &pgconn.PgError{Code: pgerrcode.ConnectionFailure}, retriable},
		{"connection does not exist", &pgconn.PgError{Code: pgerrcode.ConnectionDoesNotExist}, retriable},
		{"unable to establish connection", &pgconn.PgError{Code: pgerrcode.SQLClientUnableToEstablishSQLConnection}, retriable},
		{"wrapped connection failure", fmt.Errorf("query: %w", &pgconn.PgError{Code: pgerrcode.ConnectionFailure}), retriable},
		{"wrapped serialization failure", fmt.Errorf("query: %w", &pgconn.PgError{Code: pgerrcode.SerializationFailure}), retriable},
		{"wrapped deadlock failure", fmt.Errorf("query: %w", &pgconn.PgError{Code: pgerrcode.DeadlockDetected}), retriable},
		{"unique violation", &pgconn.PgError{Code: pgerrcode.UniqueViolation}, nonRetriable},
		{"syntax error", &pgconn.PgError{Code: pgerrcode.SyntaxError}, nonRetriable},
		{"unknown metric type", ErrUnknownMetricType, nonRetriable},
		{"plain error", errors.New("boom"), nonRetriable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classify(tt.err); got != tt.want {
				t.Errorf("Classify(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
