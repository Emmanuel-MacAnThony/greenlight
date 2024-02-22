package main

import (
	"context"
	"net/http"

	"github.com/Emmanuel-MacAnThony/greenlight/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(request *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(request.Context(), userContextKey, user)
	return request.WithContext(ctx)
}

func (app *application) contextGetUser(request *http.Request) *data.User {

	user, ok := request.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
