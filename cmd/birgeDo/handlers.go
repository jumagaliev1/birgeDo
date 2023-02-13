package main

import (
	"errors"
	"fmt"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"github.com/jumagaliev1/birgeDo/internal/validator"
	"github.com/jumagaliev1/birgeDo/pkg/forms"
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "home.page.go.html", &templateData{})
}

func (app *application) showRoom(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	room, err := app.models.Room.GetByID(id)
	if err == data.ErrRecordNotFound {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	tasks, err := app.models.Task.GetByRoomID(room.ID)
	if err == data.ErrRecordNotFound {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	usersTasks, err := app.models.Users.GetUserTask(room.ID)
	if err == data.ErrRecordNotFound {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	users, err := app.models.Users.GetAll()
	if err == data.ErrRecordNotFound {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	var tpdata = make(map[string]data.UserTasks)
	for _, ut := range usersTasks {
		if tpdata[ut.User].Task == nil {
			var dataTask []data.Task
			dataTask = append(dataTask, data.Task{Title: ut.Task, Done: ut.Done})
			tpdata[ut.User] = data.UserTasks{UserID: ut.UserID, User: ut.User, Task: &dataTask}
		} else {
			*tpdata[ut.User].Task = append(*tpdata[ut.User].Task, data.Task{Title: ut.Task, Done: ut.Done})

		}
	}
	var userTasks []data.UserTasks
	for _, tp := range tpdata {
		userTasks = append(userTasks, tp)

	}

	//var data []UserTask
	//for i := 0; i < len(tasks); i++ {
	//	for
	//	userTask := UserTask{Task: tasks[i],
	//		Done: }
	//}
	app.writeJSON(w, http.StatusOK, envelope{"room": room, "tasks": tasks, "userTasks": userTasks, "users": users}, nil)
	//app.render(w, r, "showRoom.page.go.html", &templateData{
	//	Room:     room,
	//	Tasks:    tasks,
	//	UserTask: userTasks,
	//	Users:    users,
	//})
}

func (app *application) createRoom(w http.ResponseWriter, r *http.Request) {
	app.logger.PrintInfo("Proverka", nil)
	if r.Method == "POST" {
		var input struct {
			Title string `json:"title"`
		}
		err := app.readJSON(w, r, &input)
		app.logger.PrintInfo("Proverka", nil)
		if err != nil {
			app.logError(r, err)
			app.badRequestResponse(w, r, err)
			return
		}
		room := &data.Room{
			Title: input.Title,
		}
		v := validator.New()

		if data.ValidateRoom(v, room); v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		user := app.contextGetUser(r)
		roomID, err := app.models.Room.Insert(room)
		if err != nil {
			app.logger.PrintError(err, nil)
			app.serverErrorResponse(w, r, err)
			return
		}
		err = app.models.Users.InsertRoomUser(user.ID, roomID)
		if err != nil {
			app.logger.PrintError(err, nil)
			app.serverErrorResponse(w, r, err)
			return
		}
		headers := make(http.Header)
		headers.Set("Location", fmt.Sprintf("/v1/room/%d", room.ID))

		app.session.Put(r, "flash", "Room successfully created!")
		err = app.writeJSON(w, http.StatusCreated, envelope{"room": room, "data": app.addDefaultData(&templateData{}, r)}, headers)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		//http.Redirect(w, r, fmt.Sprintf("/room/%d", roomID), http.StatusSeeOther)

	} else {
		//app.writeJSON(w, http.StatusOK, envelope{"form": forms.New(nil)}, nil)
		//app.render(w, r, "createRoom.page.go.html", &templateData{Form: forms.New(nil)})
	}

}

func (app *application) createTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string `json:"title"`
		RoomID int64  `json:"roomID"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	taskID, err := app.models.Task.Insert(&data.Task{Title: input.Title, RoomID: input.RoomID})
	usersID, err := app.models.Users.GetUsersByRoom(int(input.RoomID))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	for _, uID := range usersID {
		err = app.models.Users.InsertUserTask(uID, taskID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/room/%d", input.RoomID))
	app.writeJSON(w, http.StatusCreated, envelope{"task": input}, headers)

	//http.Redirect(w, r, fmt.Sprintf("/room/%d", input.RoomID), http.StatusSeeOther)
}
func (app *application) updateTask(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	task, err := app.models.Users.GetUserTaskByBothID(user.ID, id)
	if err == data.ErrRecordNotFound {
		app.notFoundResponse(w, r)
		return
	} else if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if task.Done == false {
		err = app.models.Task.UpdateUserTaskByBothIDTrue(user.ID, int(id))
	} else {
		err = app.models.Task.UpdateUserTaskByBothIDFalse(user.ID, int(id))
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprint("/v1/mytasks"))

	app.writeJSON(w, http.StatusSeeOther, envelope{"task": task}, headers)
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
		app.errorResponse(w, r, http.StatusOK, "No yet Tasks. You can create")
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.session.Put(r, "flash", "No yet Rooms. You can create")
			//app.render(w, r, "myRooms.page.go.html", &templateData{})
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
			//app.serverError(w, err)
		}
		return

	}
	app.writeJSON(w, http.StatusOK, envelope{"rooms": rooms}, nil)
	//app.render(w, r, "myRooms.page.go.html", &templateData{Rooms: rooms})

}

func (app *application) showUserTasks(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	tasks, err := app.models.Users.GetTasksByUser(user.ID)
	if len(tasks) == 0 {
		//TO-DO fix this
		app.session.Put(r, "flash", "No yet Tasks. You can create")
	}

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.session.Put(r, "flash", "No yet Tasks. You can create")
			//app.render(w, r, "myTasks.page.go.html", &templateData{})
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	app.writeJSON(w, http.StatusOK, envelope{"tasks": tasks}, nil)
	//app.render(w, r, "myTasks.page.go.html", &templateData{Tasks: tasks})

}

func (app *application) AddUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoomID int `json:"roomID"`
		UserID int `json:"userID"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/room/%d", input.RoomID))
	err = app.models.Users.InsertRoomUser(input.UserID, input.RoomID)
	if err != nil {
		switch {
		case err == data.ErrDuplicateKey:
			app.session.Put(r, "flash", "User almost exists")
			//http.Redirect(w, r, fmt.Sprintf("/v1/room/%d", input.RoomID), http.StatusSeeOther)
			app.writeJSON(w, http.StatusSeeOther, envelope{"input": input}, headers)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	app.writeJSON(w, http.StatusSeeOther, envelope{"input": input}, headers)
	//http.Redirect(w, r, fmt.Sprintf("/room/%d", input.RoomID), http.StatusSeeOther)
}

func (app *application) RemoveUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int `json:"userID"`
		RoomID int `json:"roomID"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.models.Users.RemoveRoomUser(input.UserID, input.RoomID)
	if err != nil {
		switch {
		case err == data.ErrDuplicateKey:
			app.session.Put(r, "flash", "User almost exists removed")
		default:
			app.serverError(w, err)
		}
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/room/%d", input.RoomID))
	err = app.writeJSON(w, http.StatusOK, envelope{"input": input}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) RemoveTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TaskID int `json:"taskID"`
		RoomID int `json:"roomID"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.models.Users.RemoveUserTask(input.TaskID, input.RoomID)
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/room/%d", input.RoomID))
	if err != nil {
		switch {
		case err == data.ErrDuplicateKey:
			app.session.Put(r, "flash", "Task almost exists removed")
			//http.Redirect(w, r, fmt.Sprintf("/room/%d", input.RoomID), http.StatusSeeOther)
			err = app.writeJSON(w, http.StatusOK, envelope{"input": input}, headers)
		default:
			app.serverError(w, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"input": input}, headers)
}
