package storage

import (
	"database/sql"
	"errors"
)

type Storage interface {
	Connect() error
	Close() error
	GetDB() *sql.DB
	New() error
	SaveURL(urlToSave, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
	GenerateAlias(int) (string, error)
}

var (
	ErrAliasExists   = errors.New("alias exists")
	ErrAliasNotFound = errors.New("alias not found")
)
