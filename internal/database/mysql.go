package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"sharing-vision-backend-test/internal/config"
)

// OpenMySQL creates and verifies a configured MySQL connection pool.
func OpenMySQL(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	connector, err := mysql.NewConnector(newMySQLConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("create MySQL connector: %w", err)
	}

	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}

	return db, nil
}

func newMySQLConfig(cfg config.Config) *mysql.Config {
	// Preserve secure driver defaults, including native password support.
	driverConfig := mysql.NewConfig()
	driverConfig.User = cfg.DBUser
	driverConfig.Passwd = cfg.DBPassword
	driverConfig.Net = "tcp"
	driverConfig.Addr = cfg.DBHost
	driverConfig.DBName = cfg.DBName
	driverConfig.ParseTime = true
	driverConfig.Collation = "utf8mb4_unicode_ci"
	driverConfig.ClientFoundRows = true
	return driverConfig
}
