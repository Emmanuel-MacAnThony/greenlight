package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(request *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": request.Method,
		"request_url":    request.URL.String(),
	})
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

func (app *application) editConflictResponse(response http.ResponseWriter, request *http.Request) {

	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(response, request, http.StatusConflict, message)
}

func (app *application) rateLimitExceededResponse(response http.ResponseWriter, request *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(response, request, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(response http.ResponseWriter, request *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(response, request, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}
