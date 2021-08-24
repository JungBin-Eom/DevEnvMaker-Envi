package data

type User struct {
	Id          string `json:"id"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	GithubName  string `json:"github_name"`
	GithubToken string `json:"github_token"`
}

type Login struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Application struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Project     string `json:"project"`
	Runtime     string `json:"runtime"`
}

type Values struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
}
