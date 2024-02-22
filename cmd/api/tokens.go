package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Emmanuel-MacAnThony/greenlight/internal/data"
	"github.com/Emmanuel-MacAnThony/greenlight/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(response http.ResponseWriter, request *http.Request) {

	// parse the email and password from the request body
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(response, request, &input)

	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(response, request)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	err = app.writeJSON(response, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}
