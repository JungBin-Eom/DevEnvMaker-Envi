package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteHandler struct {
	db *sql.DB
}

func (s *sqliteHandler) Close() {
	s.db.Close()
}

func newSqliteHandler(filepath string) DBHandler {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	statement, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS users (
			id	INTEGER PRIMARY KEY AUTOINCREMENT,
			password STRING,
			name STRING,
			github STRING
		);`)
	statement.Exec()
	return &sqliteHandler{database}
}
