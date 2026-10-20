package repository

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

func Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return ClassifyPgError(pgErr)
	}
	return NonRetriable
}

func ClassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	if pgerrcode.IsConnectionException(pgErr.Code) {
		return Retriable
	}
	return NonRetriable
}
