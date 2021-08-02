package model

import "github.com/JungBin-Eom/DevEnvMaker-Envi/data"

type DBHandler interface {
	Close()
	CheckIdDup(string) bool
	RegisterUser(data.RegUser, int) error
	IsUser(int) bool
	AuthUser(data.Login) (bool, int)
	CreateProject(data.NewProject, int) error
}

func NewDBHandler(filepath string) DBHandler {
	return newSqliteHandler(filepath)
}
