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

func (s *sqliteHandler) CheckIdDup(id string) bool {
	row, err := s.db.Query("SELECT COUNT(*) FROM users WHERE id=?", id)
	if err != nil {
		panic(err)
	}
	defer row.Close()

	row.Next()
	var count int
	row.Scan(&count)
	if count != 0 {
		return true
	} else {
		return false
	}
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
			githubId STRING,
			sessionId STRING
		);`)
	statement.Exec()
	return &sqliteHandler{database}
}
