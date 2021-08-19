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
	CreateProject(data.Project, int) (int, error)
	RemoveProject(data.Project, int) bool
	GetProjectList(int) []*data.Project

	GetApp(string, string, int) (*data.Application, error)
	GetAppList(int) []*data.Application
	CreateApp(data.Application, int) (int, error)
	RemoveApp(data.Application, int) bool
}

func NewDBHandler(filepath string) DBHandler {
	return newSqliteHandler(filepath)
}
