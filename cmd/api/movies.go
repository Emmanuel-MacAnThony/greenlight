package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Emmanuel-MacAnThony/greenlight/internal/data"
	"github.com/Emmanuel-MacAnThony/greenlight/internal/validator"
)

func (app *application) createMovieHandler(response http.ResponseWriter, request *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(response, request, &input)
	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)

	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(response, http.StatusCreated, envelope{"movie": movie}, headers)

	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}

func (app *application) showMovieHandler(response http.ResponseWriter, request *http.Request) {

	id, err := app.readIDParam(request)
	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return

	}

	err = app.writeJSON(response, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}

func (app *application) updateMovieHandler(response http.ResponseWriter, request *http.Request) {

	// Extract Movie ID from url
	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(response, request)
		return
	}
	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = app.readJSON(response, request, &input)
	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	err = app.writeJSON(response, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}

func (app *application) deleteMovieHandler(response http.ResponseWriter, request *http.Request) {

	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(response, request)
		return
	}

	err = app.models.Movies.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}

	err = app.writeJSON(response, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}
