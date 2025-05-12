package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kxddry/url-shortener/internal/config"
	"github.com/kxddry/url-shortener/internal/lib/pqlinks"
	"github.com/kxddry/url-shortener/internal/storage"
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"
	dsn := pqlinks.DataSourceName(cfg)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, db.Ping()
}

func (s *Storage) SaveURL(urlToSave, alias string, creator int64) (int64, error) {
	const op = "storage.postgres.SaveURL"
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRow(`INSERT INTO url (alias, url, createdBy) VALUES ($1, $2, $3) RETURNING id;`, alias, urlToSave, creator).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAliasExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, tx.Commit()
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	row := s.db.QueryRow(`SELECT url FROM url WHERE alias = $1;`, alias)

	var url string
	err := row.Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrAliasNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.DeleteURL"
	tx, err := s.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.ExecContext(context.Background(), "DELETE FROM url WHERE alias = $1", alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return tx.Commit()
}

func (s *Storage) Creator(alias string) (int64, error) {
	const op = "storage.Creator"

	row := s.db.QueryRow(`SELECT createdBy FROM url WHERE alias = $1;`, alias)

	var uid int64
	err := row.Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAliasNotFound)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
