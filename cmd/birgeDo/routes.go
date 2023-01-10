package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	//fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/room", app.createRoom)
	router.HandlerFunc(http.MethodPost, "/room", app.createRoom)
	router.HandlerFunc(http.MethodGet, "/room/:id", app.showRoom)
	router.HandlerFunc(http.MethodPost, "/task", app.createTask)

	//router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))
	return router
}
