package service

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"go-migration-platform/services/education/internal/repository"
)

type queryExpectation struct {
	query   string
	args    []any
	columns []string
	rows    [][]driver.Value
}

type scriptedState struct {
	expectations []queryExpectation
	index        int
}

type scriptedDriver struct {
	state *scriptedState
}

type scriptedConn struct {
	state *scriptedState
}

type scriptedRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

var scriptedDriverCounter uint64

func normalizeSQL(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func (d *scriptedDriver) Open(name string) (driver.Conn, error) {
	return &scriptedConn{state: d.state}, nil
}

func (c *scriptedConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported in scripted driver")
}

func (c *scriptedConn) Close() error {
	return nil
}

func (c *scriptedConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported in scripted driver")
}

func (c *scriptedConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.state.index >= len(c.state.expectations) {
		return nil, fmt.Errorf("unexpected query: %s", normalizeSQL(query))
	}
	expectation := c.state.expectations[c.state.index]
	c.state.index++

	actualQuery := normalizeSQL(query)
	expectedQuery := normalizeSQL(expectation.query)
	if actualQuery != expectedQuery {
		return nil, fmt.Errorf("unexpected query\nexpected: %s\nactual:   %s", expectedQuery, actualQuery)
	}
	if len(args) != len(expectation.args) {
		return nil, fmt.Errorf("unexpected args length for query %s: expected %d, got %d", expectedQuery, len(expectation.args), len(args))
	}
	for idx, arg := range args {
		if !reflect.DeepEqual(arg.Value, expectation.args[idx]) {
			return nil, fmt.Errorf("unexpected arg %d for query %s: expected %#v, got %#v", idx, expectedQuery, expectation.args[idx], arg.Value)
		}
	}

	return &scriptedRows{
		columns: expectation.columns,
		rows:    expectation.rows,
		index:   0,
	}, nil
}

func (c *scriptedConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	return nil, fmt.Errorf("unexpected exec: %s", normalizeSQL(query))
}

func (c *scriptedConn) CheckNamedValue(_ *driver.NamedValue) error {
	return nil
}

func (r *scriptedRows) Columns() []string {
	return r.columns
}

func (r *scriptedRows) Close() error {
	return nil
}

func (r *scriptedRows) Next(dest []driver.Value) error {
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

func newScriptedService(t *testing.T, expectations []queryExpectation) (*Service, func()) {
	t.Helper()

	driverName := fmt.Sprintf("scripted_service_%d", atomic.AddUint64(&scriptedDriverCounter, 1))
	state := &scriptedState{expectations: expectations}
	sql.Register(driverName, &scriptedDriver{state: state})

	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open scripted db: %v", err)
	}

	svc := New(nil, repository.New(db), nil, nil, nil, nil)
	return svc, func() {
		_ = db.Close()
		if state.index != len(state.expectations) {
			t.Fatalf("not all expectations were used: used %d of %d", state.index, len(state.expectations))
		}
	}
}

func findInstIDExpectation(userID int64, instID int64) queryExpectation {
	return queryExpectation{
		query: `
			SELECT u.inst_id
			FROM inst_user u
			LEFT JOIN org_institution i ON u.inst_id = i.id
			WHERE u.del_flag = 0 AND u.disabled = 0
			  AND i.del_flag = 0 AND i.enabled = 1
			  AND i.expire_end_time > NOW()
			  AND u.user_id = ?
			  AND i.organ_type != 2 AND i.organ_type != 10 AND i.organ_type != 11
			ORDER BY u.id
			LIMIT 1
		`,
		args:    []any{userID},
		columns: []string{"inst_id"},
		rows:    [][]driver.Value{{instID}},
	}
}
