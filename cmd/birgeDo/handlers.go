package main

import (
	"errors"
	"fmt"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"github.com/jumagaliev1/birgeDo/pkg/forms"
	"net/http"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.go.html", &templateData{})
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
		form := forms.New(r.PostForm)
		form.Required("title")
		form.MaxLength("title", 10)
		//title := r.PostForm.Get("title")
		if !form.Valid() {
			app.render(w, r, "createRoom.page.go.html", &templateData{Form: form})
			return
		}
		user := app.authenticatedUser(r)
		roomID, err := app.models.Room.Insert(&data.Room{Title: form.Get("title")})
		if err != nil {
			app.logger.PrintError(err, nil)
			app.serverError(w, err)
			return
		}
		err = app.models.Users.InsertRoomUser(user.ID, roomID)
		if err != nil {
			app.logger.PrintError(err, nil)
			app.serverError(w, err)
			return
		}
		app.session.Put(r, "flash", "Room successfully created!")
		http.Redirect(w, r, fmt.Sprintf("/room/%d", roomID), http.StatusSeeOther)

	} else {
		app.render(w, r, "createRoom.page.go.html", &templateData{Form: forms.New(nil)})
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
	taskID, err := app.models.Task.Insert(&data.Task{Title: title, RoomID: int64(id)})
	usersID, err := app.models.Users.GetUsersByRoom(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	for _, uID := range usersID {
		err = app.models.Users.InsertUserTask(uID, taskID)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/room/%d", id), http.StatusSeeOther)
}
func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {

	app.render(w, r, "signup.page.go.html", &templateData{

		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.go.html", &templateData{Form: form})
	}
	user := &data.User{
		Name:      form.Get("name"),
		Email:     form.Get("email"),
		Activated: false,
	}
	err = user.Password.Set(form.Get("password"))
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			form.Errors.Add("email", "a user with this email address already exists")
			app.render(w, r, "signup.page.go.html", &templateData{
				Form: form,
			})
			return
		default:
			app.serverError(w, err)
		}
		return
	}

	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.go.html", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("email", "password")
	user, err := app.models.Users.GetByEmail(form.Get("email"))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(w, r, "login.page.go.html", &templateData{
				Form: form,
			})
		default:
			app.serverError(w, err)
		}
		return
	}

	match, err := user.Password.Matches(form.Get("password"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !match {
		form.Errors.Add("generic", "Email or Password is incorrect")
		app.render(w, r, "login.page.go.html", &templateData{
			Form: form,
		})
		//app.invalidCredentials(w)
		return
	}
	app.session.Put(r, "userID", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "userID")

	app.session.Put(r, "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", 303)
}

func (app *application) showUserRooms(w http.ResponseWriter, r *http.Request) {
	user := app.authenticatedUser(r)

	rooms, err := app.models.Users.GetRoomsByUser(user.ID)
	if len(rooms) == 0 {
		//TO-DO fix this
		app.session.Put(r, "flash", "No yet Tasks. You can create")
	}

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.session.Put(r, "flash", "No yet Rooms. You can create")
			app.render(w, r, "myRooms.page.go.html", &templateData{})
		default:
			app.serverError(w, err)
		}
		return

	}
	app.render(w, r, "myRooms.page.go.html", &templateData{Rooms: rooms})

}

func (app *application) showUserTasks(w http.ResponseWriter, r *http.Request) {
	user := app.authenticatedUser(r)
	tasks, err := app.models.Users.GetTasksByUser(user.ID)
	if len(tasks) == 0 {
		//TO-DO fix this
		app.session.Put(r, "flash", "No yet Tasks. You can create")
	}

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.session.Put(r, "flash", "No yet Tasks. You can create")

			app.render(w, r, "myTasks.page.go.html", &templateData{})
		default:
			app.serverError(w, err)
		}
		return
	}

	app.render(w, r, "myTasks.page.go.html", &templateData{Tasks: tasks})

}

func (app *application) AddUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("roomID", "userID")
	if !form.Valid() {
		app.render(w, r, "signup.page.go.html", &templateData{Form: form})
	}
	roomID, err := strconv.Atoi(form.Get("roomID"))
	userID, err := strconv.Atoi(form.Get("userID"))
	err = app.models.Users.InsertRoomUser(userID, roomID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/room/%d", roomID), http.StatusSeeOther)
}
