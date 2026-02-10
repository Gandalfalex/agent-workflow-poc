package store

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeQuerier struct {
	rows pgx.Rows
	row  pgx.Row

	queryErr error
	execErr  error
	execTag  pgconn.CommandTag
}

func (f *fakeQuerier) Query(context.Context, string, ...any) (pgx.Rows, error) {
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	return f.rows, nil
}

func (f *fakeQuerier) QueryRow(context.Context, string, ...any) pgx.Row {
	return f.row
}

func (f *fakeQuerier) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	if f.execErr != nil {
		return pgconn.CommandTag{}, f.execErr
	}
	if f.execTag.String() != "" {
		return f.execTag, nil
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

type fakeRow struct {
	value int
	err   error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != 1 {
		return fmt.Errorf("expected one destination, got %d", len(dest))
	}
	ptr, ok := dest[0].(*int)
	if !ok {
		return errors.New("expected *int destination")
	}
	*ptr = r.value
	return nil
}

type fakeRows struct {
	values []int
	idx    int
	closed bool
	err    error
}

func (r *fakeRows) Close() {
	r.closed = true
}

func (r *fakeRows) Err() error {
	return r.err
}

func (r *fakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("SELECT")
}

func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (r *fakeRows) Next() bool {
	if r.idx >= len(r.values) {
		r.closed = true
		return false
	}
	r.idx++
	return true
}

func (r *fakeRows) Scan(dest ...any) error {
	if len(dest) != 1 {
		return fmt.Errorf("expected one destination, got %d", len(dest))
	}
	ptr, ok := dest[0].(*int)
	if !ok {
		return errors.New("expected *int destination")
	}
	*ptr = r.values[r.idx-1]
	return nil
}

func (r *fakeRows) Values() ([]any, error) {
	if r.idx == 0 || r.idx > len(r.values) {
		return nil, errors.New("no current row")
	}
	return []any{r.values[r.idx-1]}, nil
}

func (r *fakeRows) RawValues() [][]byte {
	return nil
}

func (r *fakeRows) Conn() *pgx.Conn {
	return nil
}

type fakeTx struct {
	commitErr   error
	rollbackErr error

	commits   int
	rollbacks int
}

func (tx *fakeTx) Begin(context.Context) (pgx.Tx, error) {
	return tx, nil
}

func (tx *fakeTx) Commit(context.Context) error {
	tx.commits++
	return tx.commitErr
}

func (tx *fakeTx) Rollback(context.Context) error {
	tx.rollbacks++
	return tx.rollbackErr
}

func (tx *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (tx *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (tx *fakeTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (tx *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (tx *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag(""), nil
}

func (tx *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (tx *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row {
	return &fakeRow{}
}

func (tx *fakeTx) Conn() *pgx.Conn {
	return nil
}

type fakeBeginner struct {
	tx       pgx.Tx
	beginErr error
}

func (b *fakeBeginner) Begin(context.Context) (pgx.Tx, error) {
	if b.beginErr != nil {
		return nil, b.beginErr
	}
	return b.tx, nil
}

func TestQueryOne(t *testing.T) {
	q := &fakeQuerier{row: &fakeRow{value: 42}}

	value, err := queryOne(context.Background(), q, "ignored", func(row pgx.Row) (int, error) {
		var out int
		if err := row.Scan(&out); err != nil {
			return 0, err
		}
		return out, nil
	})
	if err != nil {
		t.Fatalf("queryOne returned error: %v", err)
	}
	if value != 42 {
		t.Fatalf("expected 42, got %d", value)
	}
}

func TestQueryMany(t *testing.T) {
	rows := &fakeRows{values: []int{2, 4, 8}}
	q := &fakeQuerier{rows: rows}

	values, err := queryMany(context.Background(), q, "ignored", func(row pgx.Row) (int, error) {
		var out int
		if err := row.Scan(&out); err != nil {
			return 0, err
		}
		return out, nil
	})
	if err != nil {
		t.Fatalf("queryMany returned error: %v", err)
	}
	if len(values) != 3 || values[0] != 2 || values[1] != 4 || values[2] != 8 {
		t.Fatalf("unexpected values: %#v", values)
	}
	if !rows.closed {
		t.Fatalf("expected rows to be closed")
	}
}

func TestWithTxSuccessCommits(t *testing.T) {
	tx := &fakeTx{}
	b := &fakeBeginner{tx: tx}

	value, err := withTx(context.Background(), b, func(tx pgx.Tx) (int, error) {
		return 7, nil
	})
	if err != nil {
		t.Fatalf("withTx returned error: %v", err)
	}
	if value != 7 {
		t.Fatalf("expected 7, got %d", value)
	}
	if tx.commits != 1 {
		t.Fatalf("expected 1 commit, got %d", tx.commits)
	}
}

func TestWithTxErrorRollsBack(t *testing.T) {
	tx := &fakeTx{}
	b := &fakeBeginner{tx: tx}

	_, err := withTx(context.Background(), b, func(tx pgx.Tx) (int, error) {
		return 0, errors.New("boom")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if tx.commits != 0 {
		t.Fatalf("expected no commits, got %d", tx.commits)
	}
	if tx.rollbacks == 0 {
		t.Fatalf("expected rollback to be called")
	}
}

func TestExecOneSuccess(t *testing.T) {
	q := &fakeQuerier{execTag: pgconn.NewCommandTag("DELETE 1")}

	if err := execOne(context.Background(), q, "ignored", pgx.ErrNoRows); err != nil {
		t.Fatalf("execOne returned error: %v", err)
	}
}

func TestExecOneNotFound(t *testing.T) {
	q := &fakeQuerier{execTag: pgconn.NewCommandTag("DELETE 0")}

	err := execOne(context.Background(), q, "ignored", pgx.ErrNoRows)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestExecOnePassesExecError(t *testing.T) {
	q := &fakeQuerier{execErr: errors.New("exec failed")}

	err := execOne(context.Background(), q, "ignored", pgx.ErrNoRows)
	if err == nil || err.Error() != "exec failed" {
		t.Fatalf("expected exec error, got %v", err)
	}
}
