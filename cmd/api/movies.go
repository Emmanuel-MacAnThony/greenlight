package main

import (
	"fmt"
	"net/http"
)

func (app *application) createMovieHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Println(response, "create a movie")
}

func (app *application) showMovieHandler(response http.ResponseWriter, request *http.Request) {

	id, err := app.readIDParam(request)
	if err != nil {
		http.NotFound(response, request)
		return
	}
	fmt.Fprintf(response, "show the details of movie %d\n", id)
}
