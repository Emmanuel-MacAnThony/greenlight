package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(request *http.Request, err error) {
	app.logger.Println(err)
}

func (app *application) errorResponse(response http.ResponseWriter, request *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	err := app.writeJSON(response, status, env, nil)
	if err != nil {
		app.logError(request, err)
		response.WriteHeader(500)
	}
}

func (app *application) badRequestResponse(response http.ResponseWriter, request *http.Request, err error) {
	app.errorResponse(response, request, http.StatusBadRequest, err.Error())
}

func (app *application) serverErrorResponse(response http.ResponseWriter, request *http.Request, err error) {

	app.logError(request, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(response, request, http.StatusInternalServerError, message)

}

func (app *application) notFoundResponse(response http.ResponseWriter, request *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(response, request, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(response http.ResponseWriter, request *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", request.Method)
	app.errorResponse(response, request, http.StatusMethodNotAllowed, message)
}

func (app *application) failedValidationResponse(response http.ResponseWriter, request *http.Request, errors map[string]string) {
	app.errorResponse(response, request, http.StatusUnprocessableEntity, errors)
}
