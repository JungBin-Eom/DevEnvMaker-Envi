package data

type RegUser struct {
	Id       string `json:"id"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Login struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}
