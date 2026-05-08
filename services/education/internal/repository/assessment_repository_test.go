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
)

type assessmentRepoExpectation struct {
	query        string
	args         []any
	rows         []driver.Value
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

type assessmentRepoScriptedRows struct {
	values []driver.Value
	read   bool
}

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
	if c.state.index >= len(c.state.expectations) {
		return nil, fmt.Errorf("unexpected query: %s", normalizeAssessmentRepoSQL(query))
	}
	expectation := c.state.expectations[c.state.index]
	if expectation.commit {
		return nil, fmt.Errorf("expected commit but got query: %s", normalizeAssessmentRepoSQL(query))
	}
	actualQuery := normalizeAssessmentRepoSQL(query)
	expectedQuery := normalizeAssessmentRepoSQL(expectation.query)
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
	c.state.index++
	return &assessmentRepoScriptedRows{values: expectation.rows}, nil
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
	return []string{"id"}
}

func (r *assessmentRepoScriptedRows) Close() error {
	return nil
}

func (r *assessmentRepoScriptedRows) Next(dest []driver.Value) error {
	if r.read || len(r.values) == 0 {
		return io.EOF
	}
	copy(dest, r.values)
	r.read = true
	return nil
}

func TestSaveAssessmentDraftPersistsMainAndDetailRowsInTransaction(t *testing.T) {
	const (
		instID     int64 = 10
		draftID    int64 = 20
		operatorID int64 = 30
	)
	expectations := []assessmentRepoExpectation{
		{
			query: `
				UPDATE assessment_draft
				SET student_id = ?,
				    student_name = ?,
				    assessment_code = ?,
				    assessment_name = ?,
				    scale_version = ?,
				    birth_date = ?,
				    assessment_date = ?,
				    examiner_id = ?,
				    examiner_name = ?,
				    input_json = ?,
				    progress_json = ?,
				    answered_item_count = ?,
				    raw_score_count = ?,
				    status = ?,
				    submitted_record_id = 0,
				    remark = ?,
				    update_id = ?,
				    update_time = NOW()
				WHERE id = ? AND inst_id = ? AND del_flag = 0
			`,
			args: []any{
				int64(3),
				"张一鸣",
				"PEP3",
				"PEP-3儿童心理教育评核",
				"2025-92题版",
				nil,
				nil,
				operatorID,
				"陈老师",
				`{"studentName":"张一鸣","itemScoreList":[{"itemNo":1,"score":2}]}`,
				`{"answeredItemCount":1}`,
				1,
				1,
				"draft",
				"",
				operatorID,
				draftID,
				instID,
			},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_score
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID},
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
			args:         []any{instID, draftID, 1, 2, operatorID, operatorID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_raw_score
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID},
			rowsAffected: 1,
		},
		{
			query: `
				INSERT INTO assessment_draft_raw_score (
					inst_id, draft_id, scale_code, raw_score, create_id, update_id, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					raw_score = VALUES(raw_score),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`,
			args:         []any{instID, draftID, "DEV", 3, operatorID, operatorID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_record_value
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID},
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

	entity := AssessmentDraftEntity{
		ID:             draftID,
		InstID:         instID,
		StudentID:      3,
		StudentName:    "张一鸣",
		AssessmentCode: "PEP3",
		AssessmentName: "PEP-3儿童心理教育评核",
		ScaleVersion:   "2025-92题版",
		ExaminerID:     operatorID,
		ExaminerName:   "陈老师",
		Input: struct {
			StudentName   string           `json:"studentName"`
			ItemScoreList []map[string]int `json:"itemScoreList"`
		}{
			StudentName:   "张一鸣",
			ItemScoreList: []map[string]int{{"itemNo": 1, "score": 2}},
		},
		Progress: struct {
			AnsweredItemCount int `json:"answeredItemCount"`
		}{
			AnsweredItemCount: 1,
		},
		AnsweredItemCount: 1,
		RawScoreCount:     1,
		Status:            "draft",
		CreatedBy:         operatorID,
		UpdatedBy:         operatorID,
	}
	draftIDResult, err := repo.SaveAssessmentDraft(
		context.Background(),
		entity,
		map[int]int{1: 2},
		map[string]int{"dev": 3},
		map[int]map[string]any{1: {"惯用眼": "右眼"}},
		operatorID,
	)
	if err != nil {
		t.Fatalf("SaveAssessmentDraft returned error: %v", err)
	}
	if draftIDResult != draftID {
		t.Fatalf("expected draft id %d, got %d", draftID, draftIDResult)
	}
}

func TestSaveAssessmentDraftReusesOpenDraftWhenIdempotent(t *testing.T) {
	const (
		instID          int64 = 10
		existingDraftID int64 = 82
		operatorID      int64 = 30
	)
	assessmentDate := time.Date(2026, 5, 8, 0, 0, 0, 0, time.UTC)
	entity := AssessmentDraftEntity{
		InstID:         instID,
		StudentID:      3,
		StudentName:    "张一鸣",
		AssessmentCode: "ERXIN2",
		AssessmentName: "儿心量表-II",
		ScaleVersion:   "WS-T-580-2017",
		AssessmentDate: &assessmentDate,
		ExaminerID:     operatorID,
		ExaminerName:   "陈老师",
		Input: struct {
			ItemPassList []struct {
				ItemNo int  `json:"itemNo"`
				Passed bool `json:"passed"`
			} `json:"itemPassList"`
		}{
			ItemPassList: []struct {
				ItemNo int  `json:"itemNo"`
				Passed bool `json:"passed"`
			}{{ItemNo: 136, Passed: true}},
		},
		Progress: struct {
			AnsweredItemCount int `json:"answeredItemCount"`
		}{
			AnsweredItemCount: 1,
		},
		AnsweredItemCount: 1,
		Status:            "draft",
		CreatedBy:         operatorID,
		UpdatedBy:         operatorID,
		ReuseOpenDraft:    true,
	}
	lockKey := assessmentDraftReuseLockKey(entity)
	expectations := []assessmentRepoExpectation{
		{
			query: `SELECT GET_LOCK(?, 5)`,
			args:  []any{lockKey},
			rows:  []driver.Value{int64(1)},
		},
		{
			query: `
				SELECT id
				FROM assessment_draft
				WHERE inst_id = ?
				  AND student_id = ?
				  AND assessment_code = ?
				  AND assessment_date = ?
				  AND submitted_record_id = 0
				  AND del_flag = 0
				  AND status <> 'submitted'
				ORDER BY update_time DESC, id DESC
				LIMIT 1
				FOR UPDATE
			`,
			args: []any{instID, int64(3), "ERXIN2", assessmentDate},
			rows: []driver.Value{existingDraftID},
		},
		{
			query: `
				UPDATE assessment_draft
				SET student_id = ?,
				    student_name = ?,
				    assessment_code = ?,
				    assessment_name = ?,
				    scale_version = ?,
				    birth_date = ?,
				    assessment_date = ?,
				    examiner_id = ?,
				    examiner_name = ?,
				    input_json = ?,
				    progress_json = ?,
				    answered_item_count = ?,
				    raw_score_count = ?,
				    status = ?,
				    submitted_record_id = 0,
				    remark = ?,
				    update_id = ?,
				    update_time = NOW()
				WHERE id = ? AND inst_id = ? AND del_flag = 0
			`,
			args: []any{
				int64(3),
				"张一鸣",
				"ERXIN2",
				"儿心量表-II",
				"WS-T-580-2017",
				nil,
				assessmentDate,
				operatorID,
				"陈老师",
				`{"itemPassList":[{"itemNo":136,"passed":true}]}`,
				`{"answeredItemCount":1}`,
				1,
				0,
				"draft",
				"",
				operatorID,
				existingDraftID,
				instID,
			},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_score
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, existingDraftID},
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
			args:         []any{instID, existingDraftID, 136, 1, operatorID, operatorID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_raw_score
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, existingDraftID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_record_value
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, existingDraftID},
			rowsAffected: 1,
		},
		{commit: true},
		{
			query: `SELECT RELEASE_LOCK(?)`,
			args:  []any{lockKey},
			rows:  []driver.Value{int64(1)},
		},
	}
	repo, cleanup := newAssessmentRepoScriptedRepository(t, expectations)
	defer cleanup()

	draftIDResult, err := repo.SaveAssessmentDraft(
		context.Background(),
		entity,
		map[int]int{136: 1},
		nil,
		nil,
		operatorID,
	)
	if err != nil {
		t.Fatalf("SaveAssessmentDraft returned error: %v", err)
	}
	if draftIDResult != existingDraftID {
		t.Fatalf("expected reused draft id %d, got %d", existingDraftID, draftIDResult)
	}
}

func TestUpdateAssessmentDraftInputAndProgressIncludingSubmittedPersistsMainAndDetailRowsInTransaction(t *testing.T) {
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
				WHERE id = ? AND inst_id = ? AND del_flag = 0
			`,
			args: []any{
				`{"studentName":"张一鸣","itemScoreList":[{"itemNo":1,"score":2}]}`,
				`{"answeredItemCount":1}`,
				1,
				1,
				"submitted",
				operatorID,
				draftID,
				instID,
			},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_score
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID},
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
			args:         []any{instID, draftID, 1, 2, operatorID, operatorID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_raw_score
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID},
			rowsAffected: 1,
		},
		{
			query: `
				INSERT INTO assessment_draft_raw_score (
					inst_id, draft_id, scale_code, raw_score, create_id, update_id, create_time, update_time, del_flag
				) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 0)
				ON DUPLICATE KEY UPDATE
					inst_id = VALUES(inst_id),
					raw_score = VALUES(raw_score),
					update_id = VALUES(update_id),
					update_time = NOW(),
					del_flag = 0
			`,
			args:         []any{instID, draftID, "DEV", 3, operatorID, operatorID},
			rowsAffected: 1,
		},
		{
			query: `
				UPDATE assessment_draft_item_record_value
				SET del_flag = 1, update_id = ?, update_time = NOW()
				WHERE inst_id = ? AND draft_id = ? AND del_flag = 0
			`,
			args:         []any{operatorID, instID, draftID},
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

	err := repo.UpdateAssessmentDraftInputAndProgressIncludingSubmitted(
		context.Background(),
		instID,
		draftID,
		struct {
			StudentName   string           `json:"studentName"`
			ItemScoreList []map[string]int `json:"itemScoreList"`
		}{
			StudentName:   "张一鸣",
			ItemScoreList: []map[string]int{{"itemNo": 1, "score": 2}},
		},
		struct {
			AnsweredItemCount int `json:"answeredItemCount"`
		}{
			AnsweredItemCount: 1,
		},
		1,
		1,
		"submitted",
		operatorID,
		map[int]int{1: 2},
		map[string]int{"dev": 3},
		map[int]map[string]any{1: {"惯用眼": "右眼"}},
	)
	if err != nil {
		t.Fatalf("UpdateAssessmentDraftInputAndProgressIncludingSubmitted returned error: %v", err)
	}
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
				SELECT id
				FROM assessment_draft
				WHERE id = ? AND inst_id = ? AND del_flag = 0 AND submitted_record_id = 0
				FOR UPDATE
			`,
			args: []any{draftID, instID},
			rows: []driver.Value{draftID},
		},
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
