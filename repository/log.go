package repository

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
)

type LogDB struct {
	db *sqlx.DB
}

func NewLogDB(db *sqlx.DB) *LogDB {
	return &LogDB{db: db}
}

func (r *LogDB) RecordLog(event fsnotify.Event) error {
	query := fmt.Sprintf("INSERT INTO %s (name, operation) values ($1, $2)", "log")
	_, err := r.db.Exec(query, event.Name, event.Op)

	return err
}
