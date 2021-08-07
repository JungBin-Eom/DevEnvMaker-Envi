package data

type User struct {
	Id       string `json:"id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Login struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

type Project struct {
	Name string `json:"name"`
}

type Application struct {
	Name string `json:"name"`
}
