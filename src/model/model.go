package model

type DBHandler interface {
	Close()
}

func NewDBHandler() DBHandler {
	return newSqliteHandler()
}
