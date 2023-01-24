package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)
	router.Handler(http.MethodGet, "/", dynamicMiddleware.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRoom))
	router.Handler(http.MethodPost, "/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRoom))
	router.Handler(http.MethodGet, "/room/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showRoom))
	router.Handler(http.MethodPost, "/task", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createTask))
	router.Handler(http.MethodGet, "/task/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.updateTask))
	router.Handler(http.MethodPost, "/addUser", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.AddUser))

	router.Handler(http.MethodGet, "/myrooms", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showUserRooms))
	router.Handler(http.MethodGet, "/mytasks", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showUserTasks))

	router.Handler(http.MethodGet, "/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	router.Handler(http.MethodPost, "/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	router.Handler(http.MethodGet, "/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	router.Handler(http.MethodPost, "/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	router.Handler(http.MethodPost, "/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))

	//router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))
	return standardMiddleware.Then(router)
}
