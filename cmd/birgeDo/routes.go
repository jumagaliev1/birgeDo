package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable, noSurf)
	router.Handler(http.MethodGet, "/", dynamicMiddleware.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	router.Handler(http.MethodPost, "/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	router.Handler(http.MethodGet, "/room/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	router.Handler(http.MethodPost, "/task", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))

	router.Handler(http.MethodGet, "/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	router.Handler(http.MethodPost, "/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	router.Handler(http.MethodGet, "/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	router.Handler(http.MethodPost, "/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	router.Handler(http.MethodPost, "/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))

	//router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))
	return standardMiddleware.Then(router)
}
