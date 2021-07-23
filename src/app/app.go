package app

import (
	"net/http"

	"github.com/JungBin-Eom/DevEnvMaker-Envi/model"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var rd *render.Render

type AppHandler struct {
	http.Handler
	db model.DBHandler
}

func (a *AppHandler) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/html/signIn.html", http.StatusTemporaryRedirect)
}

func MakeHandler(filepath string) *AppHandler {
	r := mux.NewRouter()
	neg := negroni.New(negroni.NewRecovery(), negroni.NewLogger(), negroni.NewStatic(http.Dir("public")))
	neg.UseHandler(r)

	a := &AppHandler{
		Handler: neg,
		db:      model.NewDBHandler(filepath),
	}

	r.HandleFunc("/signin", a.LoginHandler)
	r.HandleFunc("/auth/github/login", a.GithubLoginHandler)
	r.HandleFunc("/auth/github/callback", a.GithubAuthCallback)

	// r.HandleFunc("/auth/google/login", a.GoogleLoginHandler)
	// r.HandleFunc("/auth/google/callback", a.GoogleAuthCallback)

	// Swagger Handlers
	opts := middleware.RedocOpts{SpecURL: "/swagger.yml"}
	sh := middleware.Redoc(opts, nil)
	r.Handle("/docs", sh)
	r.Handle("/swagger.yml", http.FileServer(http.Dir("./")))

	return a
}
