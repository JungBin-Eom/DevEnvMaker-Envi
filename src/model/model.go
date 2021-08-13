package model

import "github.com/JungBin-Eom/DevEnvMaker-Envi/data"

type DBHandler interface {
	Close()
	CheckIdDup(string) bool
	RegisterUser(data.User, int) error
	IsUser(int) bool
	AuthUser(data.Login) (bool, int, string)
	UserInfo(int) (*data.User, error)
	RegisterToken(int, string) error

	GetProject(string, int) (*data.Project, error)
	CreateProject(data.Project, int) error
	RemoveProject(data.Project, int) bool
	GetProjectList(int) []*data.Project

	GetAppList(int) []*data.Application
}

func NewDBHandler(filepath string) DBHandler {
	return newSqliteHandler(filepath)
}
