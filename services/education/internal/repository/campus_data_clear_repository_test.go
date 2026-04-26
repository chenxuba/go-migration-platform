package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type campusClearScriptState struct {
	mu           sync.Mutex
	countByQuery map[string]int64
	queryLog     []string
	execLog      []string
}

type campusClearScriptDriver struct {
	state *campusClearScriptState
}

type campusClearScriptConn struct {
	state *campusClearScriptState
}

type campusClearScriptTx struct{}

type campusClearScriptRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

var campusClearScriptDriverCounter uint64

func normalizeCampusClearSQL(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func (d *campusClearScriptDriver) Open(name string) (driver.Conn, error) {
	return &campusClearScriptConn{state: d.state}, nil
}

func (c *campusClearScriptConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported: %s", normalizeCampusClearSQL(query))
}

func (c *campusClearScriptConn) Close() error {
	return nil
}

func (c *campusClearScriptConn) Begin() (driver.Tx, error) {
	return &campusClearScriptTx{}, nil
}

func (c *campusClearScriptConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	return &campusClearScriptTx{}, nil
}

func (c *campusClearScriptConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	c.state.execLog = append(c.state.execLog, normalizeCampusClearSQL(query))
	c.state.mu.Unlock()
	return driver.RowsAffected(1), nil
}

func (c *campusClearScriptConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	normalizedQuery := normalizeCampusClearSQL(query)

	c.state.mu.Lock()
	c.state.queryLog = append(c.state.queryLog, normalizedQuery)
	count := c.state.countByQuery[normalizedQuery]
	c.state.mu.Unlock()

	return &campusClearScriptRows{
		columns: []string{"count"},
		rows:    [][]driver.Value{{count}},
	}, nil
}

func (c *campusClearScriptConn) CheckNamedValue(_ *driver.NamedValue) error {
	return nil
}

func (tx *campusClearScriptTx) Commit() error {
	return nil
}

func (tx *campusClearScriptTx) Rollback() error {
	return nil
}

func (r *campusClearScriptRows) Columns() []string {
	return r.columns
}

func (r *campusClearScriptRows) Close() error {
	return nil
}

func (r *campusClearScriptRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	row := r.rows[r.index]
	r.index++
	for idx := range dest {
		if idx < len(row) {
			dest[idx] = row[idx]
		}
	}
	return nil
}

func newCampusClearScriptDB(t *testing.T, countByQuery map[string]int64) (*sql.DB, *campusClearScriptState, func()) {
	t.Helper()

	driverName := fmt.Sprintf("campus_clear_script_%d", atomic.AddUint64(&campusClearScriptDriverCounter, 1))
	state := &campusClearScriptState{
		countByQuery: countByQuery,
		queryLog:     make([]string, 0, 32),
		execLog:      make([]string, 0, 32),
	}
	sql.Register(driverName, &campusClearScriptDriver{state: state})

	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open scripted db: %v", err)
	}

	return db, state, func() {
		_ = db.Close()
	}
}

func containsNormalizedCampusClearSQL(items []string, query string) bool {
	normalizedQuery := normalizeCampusClearSQL(query)
	for _, item := range items {
		if item == normalizedQuery {
			return true
		}
	}
	return false
}

func TestClearCampusBusinessDataClearsWeChatAndExportRecords(t *testing.T) {
	exportRecordCountQuery := `
		SELECT
			(SELECT COUNT(*) FROM enrolled_student_export_record WHERE inst_id = ? AND del_flag = 0)
			+ (SELECT COUNT(*) FROM intent_student_export_record WHERE inst_id = ? AND del_flag = 0)
			+ (SELECT COUNT(*) FROM pending_renewal_student_export_record WHERE inst_id = ? AND del_flag = 0)
			+ (SELECT COUNT(*) FROM class_record_export_record WHERE inst_id = ? AND del_flag = 0)
			+ (SELECT COUNT(*) FROM student_arrear_export_record WHERE inst_id = ? AND del_flag = 0)
	`
	templateMessageRecordCountQuery := `SELECT COUNT(*) FROM template_message_record WHERE inst_id = ? AND del_flag = 0`
	templateMessageRecordItemCountQuery := `SELECT COUNT(*) FROM template_message_record_item WHERE inst_id = ? AND del_flag = 0`
	weChatBindTicketCountQuery := `SELECT COUNT(*) FROM wechat_official_bind_ticket WHERE inst_id = ?`
	weChatStudentBindingCountQuery := `SELECT COUNT(*) FROM wechat_official_student_binding WHERE inst_id = ?`

	db, state, cleanup := newCampusClearScriptDB(t, map[string]int64{
		normalizeCampusClearSQL(exportRecordCountQuery):              5,
		normalizeCampusClearSQL(templateMessageRecordCountQuery):     2,
		normalizeCampusClearSQL(templateMessageRecordItemCountQuery): 3,
		normalizeCampusClearSQL(weChatBindTicketCountQuery):          4,
		normalizeCampusClearSQL(weChatStudentBindingCountQuery):      6,
	})
	defer cleanup()

	repo := &Repository{db: db}
	summary, err := repo.ClearCampusBusinessData(context.Background(), 10048, 9527)
	if err != nil {
		t.Fatalf("ClearCampusBusinessData returned error: %v", err)
	}

	if summary.ExportRecords != 5 {
		t.Fatalf("expected export record total 5, got %d", summary.ExportRecords)
	}
	if summary.TemplateMessageRecords != 2 {
		t.Fatalf("expected template message record total 2, got %d", summary.TemplateMessageRecords)
	}
	if summary.TemplateMessageRecordItems != 3 {
		t.Fatalf("expected template message record item total 3, got %d", summary.TemplateMessageRecordItems)
	}
	if summary.WeChatBindTickets != 4 {
		t.Fatalf("expected wechat bind ticket total 4, got %d", summary.WeChatBindTickets)
	}
	if summary.WeChatStudentBindings != 6 {
		t.Fatalf("expected wechat student binding total 6, got %d", summary.WeChatStudentBindings)
	}

	requiredQueries := []string{
		exportRecordCountQuery,
		templateMessageRecordCountQuery,
		templateMessageRecordItemCountQuery,
		weChatBindTicketCountQuery,
		weChatStudentBindingCountQuery,
	}
	for _, query := range requiredQueries {
		if !containsNormalizedCampusClearSQL(state.queryLog, query) {
			t.Fatalf("expected count query to be executed: %s", normalizeCampusClearSQL(query))
		}
	}

	requiredDeletes := []string{
		`DELETE FROM template_message_record_item WHERE inst_id = ?`,
		`DELETE FROM template_message_record WHERE inst_id = ?`,
		`DELETE FROM pending_renewal_student_export_record WHERE inst_id = ?`,
		`DELETE FROM intent_student_export_record WHERE inst_id = ?`,
		`DELETE FROM class_record_export_record WHERE inst_id = ?`,
		`DELETE FROM student_arrear_export_record WHERE inst_id = ?`,
		`DELETE FROM wechat_official_bind_ticket WHERE inst_id = ?`,
		`DELETE FROM wechat_official_student_binding WHERE inst_id = ?`,
	}
	for _, query := range requiredDeletes {
		if !containsNormalizedCampusClearSQL(state.execLog, query) {
			t.Fatalf("expected delete query to be executed: %s", normalizeCampusClearSQL(query))
		}
	}
}
