package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/lib/pq"
)

// Storage is an interface that defines the methods for interacting with the database.
type Storage interface {
	storage.Storage
}

// PostgresStorage is a struct that implements the Storage interface for PostgreSQL.
type PostgresStorage struct {
	config.Storage
	db *sql.DB
}

func NewPostgresStorage(cfg config.Storage) *PostgresStorage {
	return &PostgresStorage{
		Storage: cfg,
		db:      nil,
	}
}

func (ps *PostgresStorage) Connect() error {
	if ps.db != nil { // don't do anything if the connection is already established
		return nil
	}
	const op = "storage.postgres.Connect"
	connStr := "host=" + ps.Host + " port=" + ps.Port + " user=" + ps.User +
		" password=" + ps.Password + " dbname=" + ps.DBName + " sslmode=" + ps.SSLMode
	var err error
	ps.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return ps.db.Ping()
}

func (ps *PostgresStorage) Close() error {
	if ps.db != nil {
		return ps.db.Close()
	}
	return nil
}

func (ps *PostgresStorage) GetDB() *sql.DB {
	if ps.db == nil {
		err := ps.Connect()
		if err != nil {
			return nil
		}
	}
	return ps.db
}

func (ps *PostgresStorage) New() error {
	const op = "storage.postgres.New"
	db := ps.GetDB()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	stmt1 := `CREATE TABLE IF NOT EXISTS url(
    	id SERIAL PRIMARY KEY,
    	alias TEXT NOT NULL UNIQUE,
    	url TEXT NOT NULL
    );`

	stmt2 := `CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`

	_, err = tx.ExecContext(context.Background(), stmt1)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.ExecContext(context.Background(), stmt2)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}

func (s *PostgresStorage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.SaveURL"
	db := s.GetDB()

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := tx.ExecContext(context.Background(), "INSERT INTO url (alias, url) VALUES ($1, $2) RETURNING id", alias, urlToSave)
	if err != nil {
		// If the alias already exists, return the ErrAliasExists error
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique violation
			_ = tx.Rollback()
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAliasExists)
		}
		// if the error is not a unique violation, rollback the transaction and return the error anyway
		_ = tx.Rollback()
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	idInt, _ := id.LastInsertId()

	return idInt, tx.Commit()
}

func (s *PostgresStorage) GetURL(alias string) (string, error) {
	// returns nil if alias was found
	db := s.GetDB()
	const op = "storage.GetURL"
	var url string
	err := db.QueryRowContext(context.Background(), "SELECT url FROM url WHERE alias = $1", alias).Scan(&url)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, storage.ErrAliasNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return url, nil
}

func (s *PostgresStorage) DeleteURL(alias string) error {
	db := s.GetDB()
	const op = "storage.DeleteURL"
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.ExecContext(context.Background(), "DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit()
}

func (s *PostgresStorage) GenerateAlias(length int) (string, error) {
	alias := random.NewRandomString(length)
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			return "", fmt.Errorf("failed to generate alias: %w", ctx.Err())
		}
		_, err = s.GetURL(alias)
		if err.Error() == "storage.GetURL: "+storage.ErrAliasNotFound.Error() {
			break
		}
		if err == nil {
			alias = random.NewRandomString(length)
			continue
		}
		return "", fmt.Errorf("failed to generate alias: %w", err)
	}
	return alias, nil
}
