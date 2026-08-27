package database

import (
	"context"
	"os"
	"testing"
	"time"

	"sharing-vision-backend-test/internal/config"
)

func TestNewMySQLConfigPreservesDriverDefaults(t *testing.T) {
	t.Parallel()

	cfg := config.Load()
	driverConfig := newMySQLConfig(cfg)

	if !driverConfig.AllowNativePasswords {
		t.Error("AllowNativePasswords = false, want true")
	}
	if !driverConfig.ParseTime {
		t.Error("ParseTime = false, want true")
	}
	if !driverConfig.ClientFoundRows {
		t.Error("ClientFoundRows = false, want true")
	}
}

func TestOpenMySQLIntegration(t *testing.T) {
	if os.Getenv("MYSQL_INTEGRATION_TEST") != "1" {
		t.Skip("set MYSQL_INTEGRATION_TEST=1 to test the local MySQL connection")
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := OpenMySQL(ctx, cfg)
	if err != nil {
		t.Fatalf("OpenMySQL() returned an error: %v", err)
	}
	defer db.Close()
}
