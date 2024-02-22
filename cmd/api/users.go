package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Emmanuel-MacAnThony/greenlight/internal/data"
	"github.com/Emmanuel-MacAnThony/greenlight/internal/validator"
)

func (app *application) registerUserHandler(response http.ResponseWriter, request *http.Request) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(response, request, &input)

	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(response, request, v.Errors)
		default:
			app.serverErrorResponse(response, request, err)

		}

		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	app.background(func() {

		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl.html", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	err = app.writeJSON(response, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}

func (app *application) activateUserHandler(response http.ResponseWriter, request *http.Request) {

	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(response, request, &input)

	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}

	// Validate the plaintext token provided by the client.
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {

		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(response, request, v.Errors)
		default:
			app.serverErrorResponse(response, request, err)

		}

		return
	}

	user.Activated = true

	err = app.models.Users.UpdateUser(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	err = app.writeJSON(response, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}

// {"name": "Alice Smith", "email": "alice@example.com", "password": "pa55word"}
// {"name": "Bob Jones", "email": "bob@example.com", "password": "pa55word"}
//{"name": "Carol Smith", "email": "carol@example.com", "password": "pa55word"}
