package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type AppHandler struct {
	http.Handler
	db model.DBHandler
}

func MakeHandler() *AppHandler {
	r := mux.NewRouter()
	neg := negroni.Classic()
	neg.UseHandler(r)

	a := &AppHandler{
		Handler: neg,
		db:      model.NewDBHandler(),
	}

	r.HandleFunc("/", a.IndexHandler)

	// Swagger Handlers
	opts := middleware.RedocOpts{SpecURL: "/swagger.yml"}
	sh := middleware.Redoc(opts, nil)
	r.Handle("/docs", sh)
	r.Handle("/swagger.yml", http.FileServer(http.Dir("./")))

	return a
}
