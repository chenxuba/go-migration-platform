package config

import (
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestLoadConfiguresBeijingTimeDefaults(t *testing.T) {
	oldLocal := time.Local
	defer func() {
		time.Local = oldLocal
	}()

	t.Setenv("DB_LOC", "")
	t.Setenv("DB_TIME_ZONE", "")
	t.Setenv("APP_TIMEZONE", "")

	cfg := Load("test-service", "9999")
	parsed, err := mysql.ParseDSN(cfg.MySQLDSN())
	if err != nil {
		t.Fatalf("ParseDSN returned error: %v", err)
	}
	if got := parsed.Loc.String(); got != "Asia/Shanghai" {
		t.Fatalf("loc = %q, want Asia/Shanghai", got)
	}
	if got := parsed.Params["time_zone"]; got != "'+08:00'" {
		t.Fatalf("time_zone = %q, want '+08:00'", got)
	}
	if got := time.Local.String(); got != "Asia/Shanghai" {
		t.Fatalf("time.Local = %q, want Asia/Shanghai", got)
	}
}
