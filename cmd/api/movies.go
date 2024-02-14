package main

import (
	"fmt"
	"net/http"
	"time"

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

	fmt.Fprintf(response, "%+v\n", input)
}

func (app *application) showMovieHandler(response http.ResponseWriter, request *http.Request) {

	id, err := app.readIDParam(request)
	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	err = app.writeJSON(response, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(response, request, err)
	}

}
