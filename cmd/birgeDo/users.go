package main

import (
	"errors"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"github.com/jumagaliev1/birgeDo/internal/validator"
	"net/http"
)

// @Summary      Register User
// @Description  Registaration user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 input body data.InputRegisterUser true "Input for remove user"
// @Success      201 {object}  data.User
// @Failure      400  {object}  Error
// @Failure      422  {object}  Error
// @Failure      500  {object}  Error
// @Router       /users [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	input := data.InputRegisterUser{}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
