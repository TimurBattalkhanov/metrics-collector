package repository

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgErrorClassification int

const (
	nonRetriable pgErrorClassification = iota
	retriable
)

func classify(err error) pgErrorClassification {
	if err == nil {
		return nonRetriable
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return classifyPgError(pgErr)
	}
	return nonRetriable
}

func classifyPgError(pgErr *pgconn.PgError) pgErrorClassification {
	if pgerrcode.IsConnectionException(pgErr.Code) ||
		pgErr.Code == pgerrcode.SerializationFailure ||
		pgErr.Code == pgerrcode.DeadlockDetected {
		return retriable
	}
	return nonRetriable
}
