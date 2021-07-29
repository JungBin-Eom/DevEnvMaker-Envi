package app

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/JungBin-Eom/DevEnvMaker-Envi/data"
	"github.com/JungBin-Eom/DevEnvMaker-Envi/model"
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

func (a *AppHandler) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	var user data.Login
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	login, sessionId := a.db.AuthUser(user)
	if login == true {
		session, err := store.Get(r, "session")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		session.Values["id"] = int64(sessionId)
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

func (a *AppHandler) UserRegisterHandler(rw http.ResponseWriter, r *http.Request) {
	var user data.RegUser
	sessionId := getSessionID(r)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	err = a.db.RegisterUser(user, sessionId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rd.JSON(rw, http.StatusOK, Success{true})
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
	// r.HandleFunc("/repos", a.Repository).Methods("GET")
	r.HandleFunc("/auth/github/login", a.GithubLoginHandler)
	r.HandleFunc("/auth/github/callback", a.GithubAuthCallback)

	// Swagger Handlers
	opts := middleware.RedocOpts{SpecURL: "/swagger.yml"}
	sh := middleware.Redoc(opts, nil)
	r.Handle("/docs", sh)
	r.Handle("/swagger.yml", http.FileServer(http.Dir("./")))

	return a
}
