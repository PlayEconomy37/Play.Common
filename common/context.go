package common

import (
	"context"
	"net/http"

	"github.com/PlayEconomy37/Play.Common/database"
)

// Define a custom contextKey type with the underlying type string
type contextKey string

// Convert the string "user" to a contextKey type and assign it to the userContextKey
// constant. We'll use this constant as the key for getting and setting user information
// in the request context.
const userContextKey = contextKey("user")

// ContextSetUser returns a new copy of the request with the provided
// User struct added to the context
func (app *App) ContextSetUser(r *http.Request, user database.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

// ContextGetUser retrieves the User struct from the request context. The only
// time that we'll use this helper is when we logically expect there to be User struct
// value in the context, and if it doesn't exist it will firmly be an 'unexpected' error.
func (app *App) ContextGetUser(r *http.Request) database.User {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
