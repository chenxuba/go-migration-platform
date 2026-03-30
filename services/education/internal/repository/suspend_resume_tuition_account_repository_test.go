package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go-migration-platform/services/education/internal/model"
)

type scriptedRepoStep struct {
	kind     string
	query    string
	args     []any
	columns  []string
	rows     [][]driver.Value
	affected int64
}

type scriptedRepoState struct {
	steps []scriptedRepoStep
	index int
}

type scriptedRepoDriver struct {
	state *scriptedRepoState
}

type scriptedRepoConn struct {
	state *scriptedRepoState
}

type scriptedRepoTx struct {
	state *scriptedRepoState
}

type scriptedRepoRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

var scriptedRepoDriverCounter uint64

func normalizeRepoSQL(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func (d *scriptedRepoDriver) Open(name string) (driver.Conn, error) {
	return &scriptedRepoConn{state: d.state}, nil
}

func (c *scriptedRepoConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported: %s", normalizeRepoSQL(query))
}

func (c *scriptedRepoConn) Close() error {
	return nil
}

func (c *scriptedRepoConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *scriptedRepoConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if err := c.expectStep("begin", "", nil); err != nil {
		return nil, err
	}
	return &scriptedRepoTx{state: c.state}, nil
}

func (c *scriptedRepoConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	step, err := c.expectSQLStep("query", query, args)
	if err != nil {
		return nil, err
	}
	return &scriptedRepoRows{
		columns: step.columns,
		rows:    step.rows,
		index:   0,
	}, nil
}

func (c *scriptedRepoConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	step, err := c.expectSQLStep("exec", query, args)
	if err != nil {
		return nil, err
	}
	return driver.RowsAffected(step.affected), nil
}

func (c *scriptedRepoConn) CheckNamedValue(_ *driver.NamedValue) error {
	return nil
}

func (tx *scriptedRepoTx) Commit() error {
	return tx.expectStep("commit", "", nil)
}

func (tx *scriptedRepoTx) Rollback() error {
	return tx.expectStep("rollback", "", nil)
}

func (c *scriptedRepoConn) expectStep(kind, query string, args []driver.NamedValue) error {
	_, err := c.expectSQLStep(kind, query, args)
	return err
}

func (tx *scriptedRepoTx) expectStep(kind, query string, args []driver.NamedValue) error {
	conn := scriptedRepoConn{state: tx.state}
	_, err := conn.expectSQLStep(kind, query, args)
	return err
}

func (c *scriptedRepoConn) expectSQLStep(kind, query string, args []driver.NamedValue) (scriptedRepoStep, error) {
	if c.state.index >= len(c.state.steps) {
		return scriptedRepoStep{}, fmt.Errorf("unexpected %s: %s", kind, normalizeRepoSQL(query))
	}
	step := c.state.steps[c.state.index]
	c.state.index++
	if step.kind != kind {
		return scriptedRepoStep{}, fmt.Errorf("unexpected step kind: expected %s, got %s", step.kind, kind)
	}
	if kind == "query" || kind == "exec" {
		actualQuery := normalizeRepoSQL(query)
		expectedQuery := normalizeRepoSQL(step.query)
		if actualQuery != expectedQuery {
			return scriptedRepoStep{}, fmt.Errorf("unexpected query\nexpected: %s\nactual:   %s", expectedQuery, actualQuery)
		}
		if len(args) != len(step.args) {
			return scriptedRepoStep{}, fmt.Errorf("unexpected args length for %s: expected %d, got %d", expectedQuery, len(step.args), len(args))
		}
		for idx, arg := range args {
			if !reflect.DeepEqual(arg.Value, step.args[idx]) {
				return scriptedRepoStep{}, fmt.Errorf("unexpected arg %d for %s: expected %#v, got %#v", idx, expectedQuery, step.args[idx], arg.Value)
			}
		}
	}
	return step, nil
}

func (r *scriptedRepoRows) Columns() []string {
	return r.columns
}

func (r *scriptedRepoRows) Close() error {
	return nil
}

func (r *scriptedRepoRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.index]
	r.index++
	for idx := range dest {
		if idx < len(row) {
			dest[idx] = row[idx]
		} else {
			dest[idx] = nil
		}
	}
	return nil
}

func newScriptedRepo(t *testing.T, steps []scriptedRepoStep) (*Repository, func()) {
	t.Helper()

	driverName := fmt.Sprintf("scripted_repo_%d", atomic.AddUint64(&scriptedRepoDriverCounter, 1))
	state := &scriptedRepoState{steps: steps}
	sql.Register(driverName, &scriptedRepoDriver{state: state})

	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open scripted db: %v", err)
	}

	return New(db), func() {
		_ = db.Close()
		if state.index != len(state.steps) {
			t.Fatalf("not all scripted repo steps were used: used %d of %d", state.index, len(state.steps))
		}
	}
}

func TestSyncScheduledSuspendResumeTuitionAccounts(t *testing.T) {
	now := time.Date(2026, 3, 31, 0, 0, 1, 0, time.FixedZone("CST", 8*3600))
	repo, cleanup := newScriptedRepo(t, []scriptedRepoStep{
		{kind: "begin"},
		{
			kind: "query",
			query: `
				SELECT id
				FROM tuition_account
				WHERE del_flag = 0
				  AND IFNULL(status, 0) = ?
				  AND plan_suspend_time IS NOT NULL
				  AND plan_suspend_time <= ?
				  AND (plan_resume_time IS NULL OR plan_resume_time > ?)
				ORDER BY id ASC
			`,
			args:    []any{model.TuitionAccountStatusActive, now, now},
			columns: []string{"id"},
			rows:    [][]driver.Value{{int64(101)}},
		},
		{
			kind: "exec",
			query: `
				UPDATE tuition_account
				SET status = ?,
				    suspended_time = COALESCE(plan_suspend_time, ?),
				    status_change_time = COALESCE(plan_suspend_time, ?),
				    update_id = 0,
				    update_time = NOW()
				WHERE id IN (?)
				  AND del_flag = 0
			`,
			args:     []any{model.TuitionAccountStatusSuspended, now, now, int64(101)},
			affected: 1,
		},
		{
			kind: "exec",
			query: `
				UPDATE teaching_class_student tcs
				INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id AND tc.inst_id = tcs.inst_id AND tc.del_flag = 0
				SET tcs.class_student_status = ?,
				    tcs.update_id = ?,
				    tcs.update_time = NOW()
				WHERE tcs.del_flag = 0
				  AND tc.class_type = ?
				  AND tcs.primary_tuition_account_id IN (?)
				  AND tcs.class_student_status <> ?
			`,
			args:     []any{model.TeachingClassStudentStatusStopped, int64(0), model.TeachingClassTypeOneToOne, int64(101), model.TeachingClassStudentStatusClosed},
			affected: 1,
		},
		{
			kind: "query",
			query: `
				SELECT id
				FROM tuition_account
				WHERE del_flag = 0
				  AND IFNULL(status, 0) = ?
				  AND plan_resume_time IS NOT NULL
				  AND plan_resume_time <= ?
				ORDER BY id ASC
			`,
			args:    []any{model.TuitionAccountStatusSuspended, now},
			columns: []string{"id"},
			rows:    [][]driver.Value{{int64(202)}},
		},
		{
			kind: "exec",
			query: `
				UPDATE tuition_account
				SET status = ?,
				    suspended_time = NULL,
				    status_change_time = COALESCE(plan_resume_time, ?),
				    plan_suspend_time = NULL,
				    plan_resume_time = NULL,
				    update_id = 0,
				    update_time = NOW()
				WHERE id IN (?)
				  AND del_flag = 0
			`,
			args:     []any{model.TuitionAccountStatusActive, now, int64(202)},
			affected: 1,
		},
		{
			kind: "exec",
			query: `
				UPDATE teaching_class_student tcs
				INNER JOIN teaching_class tc ON tc.id = tcs.teaching_class_id AND tc.inst_id = tcs.inst_id AND tc.del_flag = 0
				SET tcs.class_student_status = ?,
				    tcs.update_id = ?,
				    tcs.update_time = NOW()
				WHERE tcs.del_flag = 0
				  AND tc.class_type = ?
				  AND tcs.primary_tuition_account_id IN (?)
				  AND tcs.class_student_status = ?
			`,
			args:     []any{model.TeachingClassStudentStatusStudying, int64(0), model.TeachingClassTypeOneToOne, int64(202), model.TeachingClassStudentStatusStopped},
			affected: 1,
		},
		{kind: "commit"},
	})
	defer cleanup()

	count, err := repo.SyncScheduledSuspendResumeTuitionAccounts(context.Background(), now)
	if err != nil {
		t.Fatalf("sync scheduled suspend/resume tuition accounts: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 transitioned accounts, got %d", count)
	}
}
