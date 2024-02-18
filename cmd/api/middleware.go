package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				response.Header().Set("Connection", "close")
				app.serverErrorResponse(response, request, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(response, request)
	})
}
