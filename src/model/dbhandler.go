package model

import (
	"database/sql"

	"github.com/JungBin-Eom/DevEnvMaker-Envi/data"
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

func (s *sqliteHandler) RegisterUser(user data.RegUser, sessionId int) error {
	statement, err := s.db.Prepare("INSERT INTO users (id, password, email, sessionId) VALUES (?, ?, ?, ?)")
	if err != nil {
		panic(err)
	}
	_, err = statement.Exec(user.Id, user.Password, user.Email, sessionId)
	return err
}

func (s *sqliteHandler) IsUser(sessionId int) bool {
	row, err := s.db.Query("SELECT COUNT(*) FROM users WHERE sessionId=?", sessionId)
	if err != nil {
		panic(err)
	}
	defer row.Close()

	row.Next()
	var count int
	row.Scan(&count)
	if count == 0 {
		return false
	} else {
		return true
	}
}

func (s *sqliteHandler) AuthUser(user data.Login) (bool, int) {
	row, err := s.db.Query("SELECT COUNT(*) FROM users WHERE id=? AND password=?", user.Id, user.Password)
	if err != nil {
		panic(err)
	}
	defer row.Close()

	row.Next()
	var count int
	row.Scan(&count)
	if count == 0 {
		return false, 0
	} else {
		session, err := s.db.Query("SELECT sessionId FROM users WHERE id=?", user.Id)
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.Next()
		var sessionId int
		session.Scan(&sessionId)
		return true, sessionId
	}
}

func newSqliteHandler(filepath string) DBHandler {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	createUser, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS users (
			id	STRING PRIMARY KEY,
			password STRING,
			email STRING,
			sessionId INTEGER
		);`)
	createProject, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING,
			owner STRING,
			FOREIGN KEY(owner) REFERENCES users(id)
		);`)
	createApplication, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS application (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			naem STRING,
			owner STRING,
			project STRING,
			FOREIGN KEY(owner) REFERENCES users(id),
			FOREIGN KEY(project) REFERENCES project(id)
		);`)
	createUser.Exec()
	createProject.Exec()
	createApplication.Exec()
	return &sqliteHandler{database}
}
