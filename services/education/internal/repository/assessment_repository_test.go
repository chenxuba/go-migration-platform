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
)

type assessmentRepoExpectation struct {
	query        string
	args         []any
	rowsAffected int64
	commit       bool
}

type assessmentRepoScriptedState struct {
	expectations []assessmentRepoExpectation
	index        int
}

type assessmentRepoScriptedDriver struct {
	state *assessmentRepoScriptedState
}

type assessmentRepoScriptedConn struct {
	state *assessmentRepoScriptedState
}

type assessmentRepoScriptedTx struct {
	state *assessmentRepoScriptedState
}

type assessmentRepoScriptedRows struct{}

var assessmentRepoScriptedDriverCounter uint64

func (d *assessmentRepoScriptedDriver) Open(name string) (driver.Conn, error) {
	return &assessmentRepoScriptedConn{state: d.state}, nil
}

func (c *assessmentRepoScriptedConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported in assessment repository scripted driver")
}

func (c *assessmentRepoScriptedConn) Close() error {
	return nil
}

func (c *assessmentRepoScriptedConn) Begin() (driver.Tx, error) {
	return &assessmentRepoScriptedTx{state: c.state}, nil
}

func (c *assessmentRepoScriptedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return &assessmentRepoScriptedTx{state: c.state}, nil
}

func (c *assessmentRepoScriptedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.state.index >= len(c.state.expectations) {
		return nil, fmt.Errorf("unexpected exec: %s", normalizeAssessmentRepoSQL(query))
	}
	expectation := c.state.expectations[c.state.index]
	if expectation.commit {
		return nil, fmt.Errorf("expected commit but got exec: %s", normalizeAssessmentRepoSQL(query))
	}
	actualQuery := normalizeAssessmentRepoSQL(query)
	expectedQuery := normalizeAssessmentRepoSQL(expectation.query)
	if actualQuery != expectedQuery {
		return nil, fmt.Errorf("unexpected exec\nexpected: %s\nactual:   %s", expectedQuery, actualQuery)
	}
	if len(args) != len(expectation.args) {
		return nil, fmt.Errorf("unexpected args length for exec %s: expected %d, got %d", expectedQuery, len(expectation.args), len(args))
	}
	for idx, arg := range args {
		if !reflect.DeepEqual(arg.Value, expectation.args[idx]) {
			return nil, fmt.Errorf("unexpected arg %d for exec %s: expected %#v, got %#v", idx, expectedQuery, expectation.args[idx], arg.Value)
		}
	}
	c.state.index++
	return driver.RowsAffected(expectation.rowsAffected), nil
}

func (c *assessmentRepoScriptedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return nil, fmt.Errorf("unexpected query: %s", normalizeAssessmentRepoSQL(query))
}

func (c *assessmentRepoScriptedConn) CheckNamedValue(value *driver.NamedValue) error {
	return nil
}

func (tx *assessmentRepoScriptedTx) Commit() error {
	if tx.state.index >= len(tx.state.expectations) {
		return fmt.Errorf("unexpected commit")
	}
	expectation := tx.state.expectations[tx.state.index]
	if !expectation.commit {
		return fmt.Errorf("expected exec but got commit")
	}
	tx.state.index++
	return nil
}

func (tx *assessmentRepoScriptedTx) Rollback() error {
	return nil
}

func (r *assessmentRepoScriptedRows) Columns() []string {
	return nil
}

func (r *assessmentRepoScriptedRows) Close() error {
	return nil
}

func (r *assessmentRepoScriptedRows) Next(dest []driver.Value) error {
	return io.EOF
}

func TestUpdateAssessmentDraftInputProgressAndItemDetailsPersistsOneItemInTransaction(t *testing.T) {
	score := 2
	const (
		instID     int64 = 10
		draftID    int64 = 20
		operatorID int64 = 30
	)
	expectations := []assessmentRepoExpectation{
		{
			query: `
				UPDATE assessment_draft
				SET input_json = ?,
				    progress_json = ?,
				    answered_item_count = ?,
				    raw_score_count = ?,
				    status = ?,
				    update_id = ?,
				    update_time = NOW()
				WHERE id = ? AND inst_id = ? AND del_flag = 0 AND submitted_record_id = 0
			`,
			args: []any{
				`{"itemScoreList":[{"itemNo":1,"score":2}]}`,
				`{"answeredItemCount":1}`,
				1,
				0,
				"draft",
				operatorID,
				draftID,
				instID,
			},
			rowsAffected: 1,
		},
		{
			query: `
				INSERT INTO assessment_draft_item_score (
					inst_id, draft_id, item_no, score, create_id, update_id, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					score = VALUES(score),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`,
			args:         []any{instID, draftID, 1, score, operatorID, operatorID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_record_value
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND item_no = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID, 1},
			rowsAffected: 1,
		},
		{
			query: `
				INSERT INTO assessment_draft_item_record_value (
					inst_id, draft_id, item_no, field_key, value_json, create_id, update_id, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					value_json = VALUES(value_json),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`,
			args:         []any{instID, draftID, 1, "惯用眼", `"右眼"`, operatorID, operatorID},
			rowsAffected: 1,
		},
		{commit: true},
	}
	repo, cleanup := newAssessmentRepoScriptedRepository(t, expectations)
	defer cleanup()

	err := repo.UpdateAssessmentDraftInputProgressAndItemDetails(
		context.Background(),
		instID,
		draftID,
		struct {
			ItemScoreList []map[string]int `json:"itemScoreList"`
		}{ItemScoreList: []map[string]int{{"itemNo": 1, "score": 2}}},
		struct {
			AnsweredItemCount int `json:"answeredItemCount"`
		}{AnsweredItemCount: 1},
		1,
		0,
		"draft",
		1,
		&score,
		map[string]any{"惯用眼": "右眼"},
		true,
		operatorID,
	)
	if err != nil {
		t.Fatalf("UpdateAssessmentDraftInputProgressAndItemDetails returned error: %v", err)
	}
}

func newAssessmentRepoScriptedRepository(t *testing.T, expectations []assessmentRepoExpectation) (*Repository, func()) {
	t.Helper()
	driverName := fmt.Sprintf("assessment_repo_scripted_%d", atomic.AddUint64(&assessmentRepoScriptedDriverCounter, 1))
	state := &assessmentRepoScriptedState{expectations: expectations}
	sql.Register(driverName, &assessmentRepoScriptedDriver{state: state})
	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open scripted db: %v", err)
	}
	return New(db), func() {
		_ = db.Close()
		if state.index != len(state.expectations) {
			t.Fatalf("not all expectations were used: used %d of %d", state.index, len(state.expectations))
		}
	}
}

func normalizeAssessmentRepoSQL(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}
