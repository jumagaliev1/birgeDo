package main

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"github.com/jumagaliev1/birgeDo/internal/validator"
	"net/http"
	"strconv"
	"time"
)

const SecretKey = "fhkjhfakjfiuhewbfbvbnsmsihiqe"

// @Summary      Authentication User
// @Description  Authentication user
// @Tags 		 User
// @Accept       json
// @Produce      json
// @Param		 input body data.InputAuthUser true "Input for Auth user"
// @Success      201 {object}  data.Token
// @Failure      400  {object}  Error
// @Failure      401  {object}  Error
// @Failure      422  {object}  Error
// @Failure      500  {object}  Error
// @Router       /tokens/authentication [post]
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	input := data.InputAuthUser{}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    strconv.Itoa(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	})

	token, err := claims.SignedString([]byte(SecretKey))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token, "user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
