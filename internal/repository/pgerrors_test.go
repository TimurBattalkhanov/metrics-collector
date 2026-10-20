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
		want PGErrorClassification
	}{
		{"nil", nil, NonRetriable},
		{"connection exception", &pgconn.PgError{Code: pgerrcode.ConnectionException}, Retriable},
		{"connection failure", &pgconn.PgError{Code: pgerrcode.ConnectionFailure}, Retriable},
		{"connection does not exist", &pgconn.PgError{Code: pgerrcode.ConnectionDoesNotExist}, Retriable},
		{"unable to establish connection", &pgconn.PgError{Code: pgerrcode.SQLClientUnableToEstablishSQLConnection}, Retriable},
		{"wrapped connection failure", fmt.Errorf("query: %w", &pgconn.PgError{Code: pgerrcode.ConnectionFailure}), Retriable},
		{"unique violation", &pgconn.PgError{Code: pgerrcode.UniqueViolation}, NonRetriable},
		{"syntax error", &pgconn.PgError{Code: pgerrcode.SyntaxError}, NonRetriable},
		{"unknown metric type", ErrUnknownMetricType, NonRetriable},
		{"plain error", errors.New("boom"), NonRetriable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.err); got != tt.want {
				t.Errorf("Classify(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
