package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable, app.authenticate)
	router.Handler(http.MethodGet, "/", dynamicMiddleware.ThenFunc(app.home))
	//router.Handler(http.MethodGet, "/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRoom))
	router.Handler(http.MethodPost, "/v1/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRoom))
	router.Handler(http.MethodGet, "/v1/room/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser, app.requireAccessRoom).ThenFunc(app.showRoom))
	router.Handler(http.MethodPost, "/v1/task", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createTask))
	router.Handler(http.MethodGet, "/v1/task/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.updateTask))
	router.Handler(http.MethodPost, "/v1/addUser", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.AddUser))
	router.Handler(http.MethodPost, "/v1/removeUser", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.RemoveUser))
	router.Handler(http.MethodPost, "/v1/removeTask", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.RemoveTask))

	router.Handler(http.MethodGet, "/v1/myrooms", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showUserRooms))
	router.Handler(http.MethodGet, "/v1/mytasks", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showUserTasks))

	router.Handler(http.MethodGet, "/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	router.Handler(http.MethodPost, "/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	router.Handler(http.MethodGet, "/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	router.Handler(http.MethodPost, "/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	router.Handler(http.MethodPost, "/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	//router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))
	router.ServeFiles("/static/*filepath", http.Dir("ui/static"))
	return standardMiddleware.Then(router)
}
