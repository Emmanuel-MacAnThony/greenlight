package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(response, "status: available")
	fmt.Fprintf(response, "environment: %s\n", app.config.env)
	fmt.Fprintf(response, "version: %s\n", version)
}
