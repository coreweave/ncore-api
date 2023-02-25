package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGX limited interface with high-level API for pgx methods safe to be used in high-level business logic packages.
// It is satisfied by implementations *pgx.Conn and *pgxpool.Pool (and you should probably use the second one usually).
//
// Caveat: It doesn't expose a method to acquire a *pgx.Conn or handle notifications,
// so it's not compatible with LISTEN/NOTIFY.
//
// Reference: https://pkg.go.dev/github.com/jackc/pgx/v5
type PGX interface {
	PGXQuerier
}

// PGXQuerier interface with methods used for everything, including transactions.
type PGXQuerier interface {
	// Exec executes sql. sql can be either a prepared statement name or an SQL string. arguments should be referenced
	// positionally from the sql string as $1, $2, etc.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)

	// Query executes sql with args. If there is an error the returned Rows will be returned in an error state. So it is
	// allowed to ignore the error returned from Query and handle it in Rows.
	//
	// For extra control over how the query is executed, the types QuerySimpleProtocol, QueryResultFormats, and
	// QueryResultFormatsByOID may be used as the first args to control exactly how the query is executed. This is rarely
	// needed. See the documentation for those types for details.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow is a convenience wrapper over Query. Any error that occurs while
	// querying is deferred until calling Scan on the returned Row. That Row will
	// error with ErrNoRows if no rows are returned.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Validate if the PGX interface was derived from *pgx.Conn and *pgxpool.Pool correctly.
var (
	_ PGX = (*pgx.Conn)(nil)
	_ PGX = (*pgxpool.Pool)(nil)
)
