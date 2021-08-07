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

func (s *sqliteHandler) RegisterUser(user data.User, sessionId int) error {
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

func (s *sqliteHandler) CreateProject(project data.Project, sessionId int) error {
	statement, err := s.db.Prepare("INSERT INTO projects (name, owner, description) VALUES (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	_, err = statement.Exec(project.Name, sessionId, project.Description)
	return err
}

func (s *sqliteHandler) UserInfo(sessionId int) (*data.User, error) {
	var user data.User
	row, err := s.db.Query("SELECT id, password, email FROM users WHERE sessionId=?", sessionId)
	if err != nil {
		return &user, err
	}
	defer row.Close()

	row.Next()
	row.Scan(&user.Id, &user.Password, &user.Email)
	return &user, nil
}

func (s *sqliteHandler) GetProjects(sessionId int) []*data.Project {
	projects := []*data.Project{}
	rows, err := s.db.Query("SELECT name FROM projects WHERE owner=?", sessionId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var project data.Project
		rows.Scan(&project.Name)
		projects = append(projects, &project)
	}
	return projects
}

func (s *sqliteHandler) GetApps(sessionId int) []*data.Application {
	apps := []*data.Application{}
	rows, err := s.db.Query("SELECT name FROM applications WHERE owner=?", sessionId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var app data.Application
		rows.Scan(&app.Name)
		apps = append(apps, &app)
	}
	return apps
}

func (s *sqliteHandler) RemoveProject(project data.Project, sessionId int) bool {
	row, err := s.db.Query("SELECT id FROM projects WHERE name=? AND owner=?", project.Name, sessionId)
	if err != nil {
		panic(err)
	}
	row.Next()
	var id int
	row.Scan(&id)
	row.Close()

	prjStatement, err := s.db.Prepare("DELETE FROM projects WHERE name=? AND owner=?")
	if err != nil {
		panic(err)
	}
	prjResult, err := prjStatement.Exec(project.Name, sessionId)
	if err != nil {
		panic(err)
	}

	appStatement, err := s.db.Prepare("DELETE FROM applications WHERE project=? AND owner=?")
	if err != nil {
		panic(err)
	}
	_, err = appStatement.Exec(id, sessionId)
	if err != nil {
		panic(err)
	}

	cnt, _ := prjResult.RowsAffected()
	return cnt > 0
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
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING,
			owner INTEGER,
			description STRING,
			FOREIGN KEY(owner) REFERENCES users(sessionId)
		);`)
	createApplication, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS applications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name STRING,
			owner INTEGER,
			project STRING,
			description STRING,
			FOREIGN KEY(owner) REFERENCES users(sessionId),
			FOREIGN KEY(project) REFERENCES projects(id)
		);`)
	createUser.Exec()
	createProject.Exec()
	createApplication.Exec()
	return &sqliteHandler{database}
}
