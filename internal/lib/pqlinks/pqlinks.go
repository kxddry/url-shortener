package pqlinks

import (
	"database/sql"
	"fmt"
	"github.com/kxddry/url-shortener/internal/config"
	"strings"

	_ "github.com/lib/pq"
)

func Link(cfg config.Storage) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}

func DataSourceName(cfg config.Storage) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

func EnsureDBexists(dbname, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec("CREATE DATABASE" + " " + dbname)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}
	return nil
}
