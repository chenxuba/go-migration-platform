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

	"go-migration-platform/services/education/internal/model"
)

type channelQueryExpectation struct {
	query   string
	args    []any
	columns []string
	rows    [][]driver.Value
}

type channelScriptState struct {
	expectations []channelQueryExpectation
	index        int
}

type channelScriptDriver struct {
	state *channelScriptState
}

type channelScriptConn struct {
	state *channelScriptState
}

type channelScriptRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

var channelScriptDriverCounter uint64

func normalizeChannelSQL(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func (d *channelScriptDriver) Open(name string) (driver.Conn, error) {
	return &channelScriptConn{state: d.state}, nil
}

func (c *channelScriptConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare not supported: %s", normalizeChannelSQL(query))
}

func (c *channelScriptConn) Close() error {
	return nil
}

func (c *channelScriptConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("transactions not supported")
}

func (c *channelScriptConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.state.index >= len(c.state.expectations) {
		return nil, fmt.Errorf("unexpected query: %s", normalizeChannelSQL(query))
	}
	expectation := c.state.expectations[c.state.index]
	actualQuery := normalizeChannelSQL(query)
	expectedQuery := normalizeChannelSQL(expectation.query)
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
	return &channelScriptRows{
		columns: expectation.columns,
		rows:    expectation.rows,
	}, nil
}

func (c *channelScriptConn) CheckNamedValue(_ *driver.NamedValue) error {
	return nil
}

func (r *channelScriptRows) Columns() []string {
	return r.columns
}

func (r *channelScriptRows) Close() error {
	return nil
}

func (r *channelScriptRows) Next(dest []driver.Value) error {
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

func newChannelScriptDB(t *testing.T, expectations []channelQueryExpectation) (*sql.DB, func()) {
	t.Helper()

	driverName := fmt.Sprintf("channel_script_%d", atomic.AddUint64(&channelScriptDriverCounter, 1))
	state := &channelScriptState{expectations: expectations}
	sql.Register(driverName, &channelScriptDriver{state: state})

	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open scripted db: %v", err)
	}

	return db, func() {
		_ = db.Close()
		if state.index != len(state.expectations) {
			t.Fatalf("not all expectations were used: used %d of %d", state.index, len(state.expectations))
		}
	}
}

func TestPageChannelPCUsesCorrectArgumentOrder(t *testing.T) {
	instID := int64(10048)
	categoryID := int64(49)

	db, cleanup := newChannelScriptDB(t, []channelQueryExpectation{
		{
			query: `
				SELECT COUNT(*) FROM inst_channel c WHERE c.del_flag = 0 AND (c.inst_id = ? OR c.inst_id IS NULL) AND c.category_id IN (?)
			`,
			args:    []any{instID, categoryID},
			columns: []string{"count"},
			rows:    [][]driver.Value{{int64(2)}},
		},
		{
			query: `
				SELECT c.id, IFNULL(c.uuid, ''), IFNULL(c.version, 0), IFNULL(c.channel_name, ''), c.category_id,
				       IFNULL(cc.category_name, ''), IFNULL(c.is_disabled, 0), IFNULL(c.is_default, 0), IFNULL(c.remark, ''), c.create_time,
				       (
						   SELECT COUNT(*)
						   FROM inst_student s
						   WHERE s.channel_id = c.id AND s.del_flag = 0
				       ) AS invalid_count,
				       (
						   SELECT COUNT(DISTINCT so.student_id)
						   FROM sale_order so
						   INNER JOIN inst_student s ON so.student_id = s.id AND s.channel_id = c.id
						   WHERE so.del_flag = 0
						     AND so.inst_id = c.inst_id
						     AND so.order_status = ?
				       ) AS deal_transform_count
				FROM inst_channel c
				LEFT JOIN inst_channel_category cc ON cc.id = c.category_id
				WHERE c.del_flag = 0 AND (c.inst_id = ? OR c.inst_id IS NULL) AND c.category_id IN (?)
				ORDER BY c.create_time DESC
				LIMIT ? OFFSET ?
			`,
			args: []any{model.OrderStatusCompleted, instID, categoryID, 50, 0},
			columns: []string{
				"id", "uuid", "version", "channel_name", "category_id",
				"category_name", "is_disabled", "is_default", "remark", "create_time",
				"invalid_count", "deal_transform_count",
			},
			rows: [][]driver.Value{
				{int64(62), "uuid-62", int64(0), "新增渠道", int64(49), "嬢嬢", int64(0), int64(0), "", nil, int64(0), int64(0)},
				{int64(61), "uuid-61", int64(0), "快手", int64(49), "嬢嬢", int64(0), int64(0), "", nil, int64(0), int64(0)},
			},
		},
	})
	defer cleanup()

	repo := &Repository{db: db}
	result, err := repo.PageChannelPC(context.Background(), instID, model.ChannelPCQueryDTO{
		PageRequestModel: model.PageRequestModel{
			PageIndex: 1,
			PageSize:  50,
		},
		QueryModel: model.ChannelPCQueryModel{
			ChannelTypeIDs: []int64{categoryID},
		},
		SortModel: model.ChannelPCSortModel{},
	})
	if err != nil {
		t.Fatalf("PageChannelPC returned error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].ChannelName != "新增渠道" || result.Items[1].ChannelName != "快手" {
		t.Fatalf("unexpected items: %+v", result.Items)
	}
}
