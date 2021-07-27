package model

type DBHandler interface {
	Close()
	CheckIdDup(string) bool
}

func NewDBHandler(filepath string) DBHandler {
	return newSqliteHandler(filepath)
}
