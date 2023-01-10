package main

import (
	"fmt"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"html/template"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/home.page.go.html",
		"./ui/html/base.layout.go.html",
		"./ui/html/footer.partial.go.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.logger.PrintError(err, nil)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	err = ts.Execute(w, nil)
	if err != nil {
		app.logger.PrintError(err, nil)
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *application) showRoom(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFound(w)
		return
	}
	room, err := app.models.Room.GetByID(id)
	if err == data.ErrRecordNotFound {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "showRoom.page.go.html", &templateData{
		Room: room,
	})
}

func (app *application) createRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		title := r.PostForm.Get("title")
		id, err := app.models.Room.Insert(&data.Room{Title: title})
		if err != nil {
			app.serverError(w, err)
			return
		}
		app.session.Put(r, "flash", "Room successfully created!")
		http.Redirect(w, r, fmt.Sprintf("/room/%d", id), http.StatusSeeOther)

	} else {
		app.render(w, r, "createRoom.page.go.html", &templateData{})
	}

}

func (app *application) createTask(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	title := r.PostForm.Get("title")
	roomID := r.PostForm.Get("room_id")
	id, err := strconv.Atoi(roomID)
	err = app.models.Task.Insert(&data.Task{Title: title, RoomID: int64(id)})
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/room/%d", id), http.StatusSeeOther)
}
