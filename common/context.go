package common

import (
	"context"
	"net/http"
)

// Define a custom contextKey type with the underlying type string
type contextKey string

// Convert the string "user" to a contextKey type and assign it to the userContextKey
// constant. We'll use this constant as the key for getting and setting user information
// in the request context.
const userContextKey = contextKey("user")

// The contextSetUser() method returns a new copy of the request with the provided
// User struct added to the context
func (app *App) contextSetUser(r *http.Request, user user) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

// The contextGetUser() retrieves the User struct from the request context. The only
// time that we'll use this helper is when we logically expect there to be User struct
// value in the context, and if it doesn't exist it will firmly be an 'unexpected' error.
func (app *App) contextGetUser(r *http.Request) user {
	user, ok := r.Context().Value(userContextKey).(user)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
