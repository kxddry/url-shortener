package storage

import (
	"errors"
)

type Storage interface {
	SaveURL(urlToSave, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
	GenerateAlias(int) (string, error)
}

var (
	ErrAliasExists   = errors.New("alias exists")
	ErrAliasNotFound = errors.New("alias not found")
)
