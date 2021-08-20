package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/JungBin-Eom/DevEnvMaker-Envi/data"
	"github.com/JungBin-Eom/DevEnvMaker-Envi/model"
	"github.com/bndr/gojenkins"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSEION_KEY")))
var rd *render.Render = render.New()

type AppHandler struct {
	http.Handler
	db model.DBHandler
}

type Success struct {
	Success bool `json:"success"`
}

type CreateSuccess struct {
	Success bool `json:"success"`
	Count   int  `json:"count"`
}

var getSessionID = func(r *http.Request) int {
	session, err := store.Get(r, "session")
	if err != nil {
		return 0
	}

	val := session.Values["id"]
	if val == nil {
		return 0
	}
	return int(val.(int64))
}

var getSessionName = func(r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}

	val := session.Values["login"]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (a *AppHandler) IndexHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/index.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) SignInHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/signin.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) SignUpHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/signup.html", http.StatusTemporaryRedirect)
}

// USER APIs
func (a *AppHandler) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	var user data.Login
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
		// http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	login, sessionId, githubName := a.db.AuthUser(user)
	if login == true {
		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		session.Values["id"] = int64(sessionId)
		session.Values["login"] = githubName
		err = session.Save(r, rw)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		rd.JSON(rw, http.StatusOK, Success{login})
	} else {
		rd.JSON(rw, http.StatusBadRequest, Success{login})
	}
}

type Duplicated struct {
	Duplicated bool `json:"duplicated"`
}

func (a *AppHandler) DupCheckHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := vars["id"]
	dup := a.db.CheckIdDup(id) // T: Dup, F: Not Dup
	rd.JSON(rw, http.StatusOK, Duplicated{dup})
}

func (a *AppHandler) UserInfoHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	user, err := a.db.UserInfo(sessionId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rd.JSON(rw, http.StatusOK, user)
}

func (a *AppHandler) UserRegisterHandler(rw http.ResponseWriter, r *http.Request) {
	var user data.User
	sessionId := getSessionID(r)
	githubName := getSessionName(r)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
		// http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	user.GithubName = githubName
	err = a.db.RegisterUser(user, sessionId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rd.JSON(rw, http.StatusOK, Success{true})
}

func (a *AppHandler) RegisterTokenHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	type Token struct {
		Token string `json:"token"`
	}
	var token Token
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&token)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	err = a.db.RegisterToken(sessionId, token.Token)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	rd.JSON(rw, http.StatusOK, Success{Success: true})
}

// PROJECT APIs
func (a *AppHandler) GetProjectsHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	list := a.db.GetProjectList(sessionId)
	rd.JSON(rw, http.StatusOK, list)
}

func (a *AppHandler) CreateProjectHandler(rw http.ResponseWriter, r *http.Request) {
	var newprj data.Project
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	if user.GithubToken == "NULL" {
		rd.JSON(rw, http.StatusOK, CreateSuccess{false, 0})
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newprj)
		if err != nil {
			rd.JSON(rw, http.StatusOK, CreateSuccess{false, 0})
		}

		count, err := a.db.CreateProject(newprj, sessionId)
		if count != 0 {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, count})
			return
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		// ctx := context.Background()
		pw := os.Getenv("JENKINS_PW")
		jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
		// Provide CA certificate if server is using self-signed certificate
		// caCert, _ := ioutil.ReadFile("/tmp/ca.crt")
		// jenkins.Requester.CACert = caCert
		_, err = jenkins.Init(r.Context())

		if err != nil {
			panic("Something Went Wrong")
		}

		jenkins.CreateFolder(r.Context(), newprj.Name)

		// pbytes, _ := json.Marshal(newprj)
		// buff := bytes.NewBuffer(pbytes)

		// req, err := http.NewRequest("POST", "https://api.github.com/user/repos", buff)
		// if err != nil {
		// 	rd.JSON(rw, http.StatusOK, Success{false})
		// }
		// req.Header.Set("content-type", "application/json")
		// req.Header.Set("authorization", "token "+user.GithubToken)
		// req.Header.Set("user-agent", user.GithubName)

		// res, err := http.DefaultClient.Do(req)
		// if err != nil {
		// 	rd.JSON(rw, http.StatusOK, Success{false})
		// }
		// defer res.Body.Close()
		// _, err := ioutil.ReadAll(res.Body)
		// if err != nil {
		// 	http.Error(rw, "Unable to read body", http.StatusBadRequest)
		// }

		rd.JSON(rw, http.StatusOK, CreateSuccess{true, 0})
	}
}

func (a *AppHandler) RemoveProjectHandler(rw http.ResponseWriter, r *http.Request) {
	var project data.Project
	sessionId := getSessionID(r)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&project)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	ok := a.db.RemoveProject(project, sessionId)
	if ok {
		rd.JSON(rw, http.StatusOK, Success{true})
	} else {
		rd.JSON(rw, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) GetProjectDetailHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name, _ := vars["name"]
	sessionId := getSessionID(r)
	project, err := a.db.GetProject(name, sessionId)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	rd.JSON(rw, http.StatusOK, project)
}

// APPLICATION APIs
func (a *AppHandler) GetAppsHandler(rw http.ResponseWriter, r *http.Request) {
	sessionId := getSessionID(r)
	list := a.db.GetAppList(sessionId)
	rd.JSON(rw, http.StatusOK, list)
}

func (a *AppHandler) CreateAppHandler(rw http.ResponseWriter, r *http.Request) {
	var newapp data.Application
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	if user.GithubToken == "NULL" {
		rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, 0})
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newapp)
		if err != nil {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, 0})
		}

		count, err := a.db.CreateApp(newapp, sessionId)
		if count != 0 {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, count})
			return
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		req, err := http.NewRequest("POST", "https://api.github.com/repos/Ricky-Envi/Envi-React/forks", nil)
		if err != nil {
			rd.JSON(rw, http.StatusInternalServerError, CreateSuccess{false, 0})
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, 0})
		}
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(rw, "Unable to read body", http.StatusBadRequest)
		}
		res.Body.Close()

		pbytes, _ := json.Marshal(newapp)
		buff := bytes.NewBuffer(pbytes)
		req, err = http.NewRequest("PATCH", "https://api.github.com/repos/"+user.GithubName+"/Envi-"+newapp.Runtime, buff)
		if err != nil {
			rd.JSON(rw, http.StatusInternalServerError, CreateSuccess{false, 0})
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		res, err = http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusBadRequest, CreateSuccess{false, 0})
		}
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(rw, "Unable to read body", http.StatusBadRequest)
		}
		res.Body.Close()

		rd.JSON(rw, http.StatusOK, CreateSuccess{true, 0})
	}
}

func (a *AppHandler) RemoveAppHandler(rw http.ResponseWriter, r *http.Request) {
	var app data.Application
	sessionId := getSessionID(r)
	user, _ := a.db.UserInfo(sessionId)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&app)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	ok := a.db.RemoveApp(app, sessionId)
	if ok {
		req, err := http.NewRequest("DELETE", "https://api.github.com/repos/"+user.GithubName+"/"+app.Name, nil)
		if err != nil {
			rd.JSON(rw, http.StatusInternalServerError, Success{false})
			return
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("authorization", "token "+user.GithubToken)
		req.Header.Set("user-agent", user.GithubName)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			rd.JSON(rw, http.StatusBadRequest, Success{false})
			return
		}
		_, err = ioutil.ReadAll(res.Body)
		if err != nil {
			http.Error(rw, "Unable to read body", http.StatusBadRequest)
		}
		res.Body.Close()
		rd.JSON(rw, http.StatusOK, Success{true})
	} else {
		rd.JSON(rw, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) GetAppDetailHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project, _ := vars["projname"]
	name, _ := vars["appname"]
	sessionId := getSessionID(r)
	app, err := a.db.GetApp(project, name, sessionId)
	if err != nil {
		http.Redirect(rw, r, "/html/404.html", http.StatusBadRequest)
	}
	rd.JSON(rw, http.StatusOK, app)
}

func (a *AppHandler) BuildAppHandler(rw http.ResponseWriter, r *http.Request) {
	var app data.Application
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&app)
	if err != nil {
		http.Redirect(rw, r, "../html/404.html", http.StatusTemporaryRedirect)
	}
	ctx := context.Background()
	pw := os.Getenv("JENKINS_PW")
	jenkins := gojenkins.CreateJenkins(nil, "http://jenkins.3.35.25.64.sslip.io", "admin", pw)
	// Provide CA certificate if server is using self-signed certificate
	// caCert, _ := ioutil.ReadFile("/tmp/ca.crt")
	// jenkins.Requester.CACert = caCert
	_, err = jenkins.Init(ctx)

	if err != nil {
		panic("Something Went Wrong")
	}

	_, _ = jenkins.CreateFolder(ctx, "newFolder")
	// 1. Jenkins Pipeline 생성
	// 2. Jenkins Pipeline 실행

	rd.JSON(rw, http.StatusOK, Success{true})
}

// GITHUB APIs
func (a *AppHandler) GetGitHubNameHandler(rw http.ResponseWriter, r *http.Request) {
	githubName := getSessionName(r)
	rd.Text(rw, http.StatusOK, githubName)
}

func CheckSignin(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if strings.Contains(r.URL.Path, "/sign") || strings.Contains(r.URL.Path, "/auth") {
		next(rw, r)
		return
	}

	// if user already signed in
	sessionID := getSessionID(r)
	if sessionID != 0 {
		next(rw, r)
		return
	}

	// redirect signin.html
	http.Redirect(rw, r, "/html/signin.html", http.StatusTemporaryRedirect)
}

func MakeHandler(filepath string) *AppHandler {
	r := mux.NewRouter()

	neg := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.HandlerFunc(CheckSignin), negroni.NewStatic(http.Dir("public")))
	neg.UseHandler(r)

	a := &AppHandler{
		Handler: neg,
		db:      model.NewDBHandler(filepath),
	}

	r.HandleFunc("/", a.IndexHandler)
	r.HandleFunc("/signin", a.SignInHandler).Methods("GET")
	r.HandleFunc("/signin", a.LoginHandler).Methods("POST")
	r.HandleFunc("/signup", a.SignUpHandler)
	r.HandleFunc("/signup/idcheck/{id:[a-zA-Z0-9]+}", a.DupCheckHandler).Methods("GET")
	r.HandleFunc("/signup/register", a.UserRegisterHandler).Methods("POST")

	r.HandleFunc("/user", a.UserInfoHandler).Methods("GET")
	r.HandleFunc("/user/token", a.RegisterTokenHandler).Methods("POST")

	r.HandleFunc("/project", a.GetProjectsHandler).Methods("GET")
	r.HandleFunc("/project", a.RemoveProjectHandler).Methods("DELETE")
	r.HandleFunc("/project", a.CreateProjectHandler).Methods("POST")
	r.HandleFunc("/project/{name:[a-zA-Z0-9]+}", a.GetProjectDetailHandler).Methods("GET")

	r.HandleFunc("/app", a.GetAppsHandler).Methods("GET")
	r.HandleFunc("/app", a.CreateAppHandler).Methods("POST")
	r.HandleFunc("/app", a.RemoveAppHandler).Methods("DELETE")
	r.HandleFunc("/app/{projname:[a-zA-Z0-9]+}/{appname:[a-zA-Z0-9]+}", a.GetAppDetailHandler).Methods("GET")
	r.HandleFunc("/app/build", a.BuildAppHandler).Methods("POST")

	// r.HandleFunc("/repos", a.Repository).Methods("GET")

	r.HandleFunc("/github/name", a.GetGitHubNameHandler).Methods("GET")

	r.HandleFunc("/auth/github/login", a.GithubLoginHandler)
	r.HandleFunc("/auth/github/callback", a.GithubAuthCallback)

	// Swagger Handlers
	opts := middleware.RedocOpts{SpecURL: "/swagger.yml"}
	sh := middleware.Redoc(opts, nil)
	r.Handle("/docs", sh)
	r.Handle("/swagger.yml", http.FileServer(http.Dir("./")))

	return a
}
