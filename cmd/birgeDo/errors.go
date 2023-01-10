package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func (app *application) notFound(w http.ResponseWriter) {
	message := "the requested resource could not be found"
	http.Error(w, message, 404)
}
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.logger.PrintError(err, map[string]string{"trace": trace})
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
