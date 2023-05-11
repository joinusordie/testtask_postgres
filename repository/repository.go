package repository

import (
	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
)

type Log interface {
	RecordLog(event fsnotify.Event) error
}

type Repository struct {
	Log
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Log: NewLogDB(db),
	}
}
