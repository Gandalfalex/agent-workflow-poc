package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type dbQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type txBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

func queryMany[T any](ctx context.Context, q dbQuerier, query string, scan func(pgx.Row) (T, error), args ...any) ([]T, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []T{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func queryOne[T any](ctx context.Context, q dbQuerier, query string, scan func(pgx.Row) (T, error), args ...any) (T, error) {
	return scan(q.QueryRow(ctx, query, args...))
}

func execOne(ctx context.Context, q dbQuerier, query string, notFoundErr error, args ...any) error {
	tag, err := q.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if notFoundErr != nil && tag.RowsAffected() == 0 {
		return notFoundErr
	}
	return nil
}

func withTx[T any](ctx context.Context, b txBeginner, fn func(pgx.Tx) (T, error)) (T, error) {
	tx, err := b.Begin(ctx)
	if err != nil {
		var zero T
		return zero, err
	}
	defer tx.Rollback(ctx)

	result, err := fn(tx)
	if err != nil {
		var zero T
		return zero, err
	}

	if err := tx.Commit(ctx); err != nil {
		var zero T
		return zero, err
	}

	return result, nil
}
