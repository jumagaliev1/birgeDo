package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	standardMiddleware := alice.New(app.enableCORS, app.recoverPanic, app.logRequest, secureHeaders)
	dynamicMiddleware := alice.New(app.session.Enable, app.authenticate)
	router.Handler(http.MethodPost, "/v1/room", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createRoom))
	router.Handler(http.MethodGet, "/v1/room/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser, app.requireAccessRoom).ThenFunc(app.showRoom))
	router.Handler(http.MethodPost, "/v1/task", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createTask))
	router.Handler(http.MethodGet, "/v1/task/:id", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.updateTask))
	router.Handler(http.MethodPost, "/v1/addUser", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.AddUser))
	router.Handler(http.MethodPost, "/v1/removeUser", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.RemoveUser))
	router.Handler(http.MethodPost, "/v1/removeTask", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.RemoveTask))

	router.Handler(http.MethodGet, "/v1/myrooms", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showUserRooms))
	router.Handler(http.MethodGet, "/v1/mytasks", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.showUserTasks))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	//router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))
	router.HandlerFunc(http.MethodGet, "/swagger/*any", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4000/static/swagger.json")))

	router.ServeFiles("/static/*filepath", http.Dir("docs"))
	return standardMiddleware.Then(router)
}
