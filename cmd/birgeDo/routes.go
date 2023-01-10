package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	_ "github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable)
	router.Handler(http.MethodGet, "/", dynamicMiddleware.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/room", dynamicMiddleware.ThenFunc(app.createRoom))
	router.Handler(http.MethodPost, "/room", dynamicMiddleware.ThenFunc(app.createRoom))
	router.Handler(http.MethodGet, "/room/:id", dynamicMiddleware.ThenFunc(app.showRoom))
	router.Handler(http.MethodPost, "/task", dynamicMiddleware.ThenFunc(app.createTask))

	//router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))
	return standardMiddleware.Then(router)
}
