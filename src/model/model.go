package model

import "github.com/JungBin-Eom/DevEnvMaker-Envi/data"

type DBHandler interface {
	Close()
	CheckIdDup(string) bool
	RegisterUser(data.User, int) error
	IsUser(int) bool
	AuthUser(data.Login) (bool, int)
	UserInfo(int) (*data.User, error)

	CreateProject(data.Project, int) error
	GetProjects(int) []*data.Project

	GetApps(int) []*data.Application
}

func NewDBHandler(filepath string) DBHandler {
	return newSqliteHandler(filepath)
}
