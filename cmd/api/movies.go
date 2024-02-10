package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Emmanuel-MacAnThony/greenlight/internal/data"
)

func (app *application) createMovieHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Println(response, "create a movie")
}

func (app *application) showMovieHandler(response http.ResponseWriter, request *http.Request) {

	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(response, request)
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
