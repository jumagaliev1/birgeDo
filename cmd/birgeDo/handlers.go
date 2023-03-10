package main

import (
	"errors"
	"fmt"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"github.com/jumagaliev1/birgeDo/internal/validator"
	_ "github.com/swaggo/swag/example/celler/httputil"
	"net/http"
)

// @Summary      Show Room Data
// @Description  get room data by id
//@Security	ApiKeyAuth
// @Tags         Room
// @Accept       json
// @Produce      json
// @Param        id path  int  true  "Room ID"
// @Success      200  {object}  []data.UserTasks
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /room/{id} [get]
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

	app.writeJSON(w, http.StatusOK, envelope{"room": room, "tasks": tasks, "userTasks": userTasks, "users": users}, nil)

}

// @Summary      Create Room
// @Description  Create Room ...
//@Security	ApiKeyAuth
// @Tags         Room
// @Accept       json
// @Produce      json
// @Param InputCreteRoom body data.InputCreateRoom true "Input for create room" ""
// @Success      200  {object}  data.Room
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /room [post]
func (app *application) createRoom(w http.ResponseWriter, r *http.Request) {
	input := data.InputCreateRoom{}
	err := app.readJSON(w, r, &input)
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

}

// @Summary      Create Task
// @Description  Create Task ...
//@Security	ApiKeyAuth
// @Tags         Task
// @Accept       json
// @Produce      json
// @Success      200  {object}  data.Task
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /task [post]
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
}

// @Summary      Update Task
// @Description  Update Task if true to false otherwise
//@Security	ApiKeyAuth
// @Tags         Task
// @Accept       json
// @Produce      json
// @Param        id path  int  true  "Task ID"
// @Success      200  {object}  data.Task
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /task/{id} [get]
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

// @Summary      Show User Rooms
// @Description  ...
//@Security	ApiKeyAuth
// @Tags         Room
// @Accept       json
// @Produce      json
// @Success      200  {object}  []data.Room
// @Failure      400  {object}  Error
// @Failure 401 {object} Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /myrooms/ [get]
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

// @Summary      Show User Tasks
// @Description  ..
// @Tags         Task
//@Security	ApiKeyAuth
// @Accept       json
//@Security	ApiKeyAuth
// @Produce      json
// @Success      200  {object}  []data.Task
// @Failure      400  {object}  Error
// @Failure      404  {object}  Error
// @Failure      500  {object}  Error
// @Router       /mytasks [get]
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

// @Summary      Add User to Room
// @Description  Add User to Room
// @Tags Room
// @Security	ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param		 roomID body data.InputAddUser true "Input for adding user"
// @Success      200  {object}  data.UserTasks
// @Failure      400  {object}  Error
// @Failure      500  {object}  Error
// @Router       /addUser [post]
func (app *application) AddUser(w http.ResponseWriter, r *http.Request) {
	input := data.InputAddUser{}
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
			app.writeJSON(w, http.StatusSeeOther, envelope{"input": input}, headers)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	app.writeJSON(w, http.StatusSeeOther, envelope{"input": input}, headers)
}

// @Summary      Remove User
// @Description  Remove user from Room
// @Tags 		 Room
// @Accept       json
//@Security	ApiKeyAuth
// @Produce      json
// @Param		 input body data.InputAddUser true "Input for remove user"
// @Success      200  {object}  data.InputAddUser
// @Failure      400  {object}  Error
// @Failure      500  {object}  Error
// @Router       /removeUser [post]
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
	err = app.writeJSON(w, http.StatusSeeOther, envelope{"input": input}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary      Remove Task
// @Description  Remove task from Room
// @Tags 		 Task
//@Security		 ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param		 input body data.InputRemoveTask true "Input for remove user"
// @Success      200  {object}  data.UserTasks
// @Failure      400  {object}  Error
// @Failure      500  {object}  Error
// @Router       /removeTask [post]
func (app *application) RemoveTask(w http.ResponseWriter, r *http.Request) {
	input := data.InputRemoveTask{}
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
			err = app.writeJSON(w, http.StatusOK, envelope{"input": input}, headers)
		default:
			app.serverError(w, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusSeeOther, envelope{"input": input}, headers)
}
