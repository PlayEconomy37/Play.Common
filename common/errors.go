package common

import (
	"fmt"
	"net/http"

	"github.com/PlayEconomy37/Play.Common/types"
)

// Generic helper for logging an error message
func (app *App) logError(r *http.Request, err error) {
	app.Logger.Error(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// Generic helper for sending JSON-formatted error
// messages to the client with a given status code
func (app *App) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := types.Envelope{"error": message}

	err := app.WriteJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// This method will be used to send a 500 Internal Server Error status code when our application encounters an unexpected problem at runtime
func (app *App) ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "The server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// This method will be used to send a 404 Not Found status code and JSON response to the client
func (app *App) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// This method will be used to send a 405 Method Not Allowed
// status code and JSON response to the client
func (app *App) MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// This method will be used to send a 400 Bad Request
// status code and JSON response to the client
func (app *App) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// This method will be used to send a 422 Unprocessable Entity status code and
// the contents of the errors map from our Validator type as a JSON response body
func (app *App) FailedValidation(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// This method will be used to send a 409 Conflict status code and
// JSON response to the client
func (app *App) EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// This method will be used to send a 429 Too Many Requests status code when our application encounters too many requests at the same time
func (app *App) RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// This method will be used to send a 401 Unauthorized status code forproviding invalid authentication credentials
func (app *App) InvalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// This method will be used to send a 401 Unauthorized status code for providing an invalid authentication token
func (app *App) InvalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// This method will be used to send a 401 Unauthorized status code due to user not being authenticated when trying to access a resource
func (app *App) AuthenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// This method will be used to send a 403 Forbidden status code for user not having necessary permissions when trying to access a resource
func (app *App) NotPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
