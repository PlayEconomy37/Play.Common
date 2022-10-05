package common

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PlayEconomy37/Play.Common/database"
	"github.com/PlayEconomy37/Play.Common/opentelemetry"
	"github.com/PlayEconomy37/Play.Common/permissions"
	"github.com/felixge/httpsnoop"
	"github.com/pascaldekloe/jwt"
)

// RecoverPanic is a middleware used to make sure that any panics are handled properly in our application
func (app *App) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")

				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helpers
				app.ServerErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// HTTPMetrics is a middleware used to set HTTP metrics for every HTTP request
func (app *App) HTTPMetrics(appName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Create HTTP  metrics
		httpMetrics := opentelemetry.CreateHTTPMetrics(appName)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment the number of requests received by 1
			httpMetrics.TotalRequestsCounter.WithLabelValues(r.Method, r.URL.Path).Inc()

			// This function wraps a http.Handler (in this case, the next function), executes the handler and then returns a Metrics struct
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			// On the way back up the middleware chain, increment the number of responses sent by 1
			httpMetrics.TotalResponsesCounter.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(metrics.Code)).Inc()

			// Get the request processing time in microseconds from httpsnoop and increment
			// the cumulative processing time
			httpMetrics.TotalProcessingTimeCounter.WithLabelValues(r.Method, r.URL.Path).Observe(float64(metrics.Duration.Microseconds()))
		})
	}
}

// SecureHeaders is a middleware used to instruct the user’s web browser to implement some
// additional security measures to help prevent XSS and Clickjacking attacks
func (app *App) SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

// LogRequest is a middleware used to log every HTTP request that comes to our application
func (app *App) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		properties := map[string]string{
			"ipAddress": r.RemoteAddr,
			"protocol":  r.Proto,
			"method":    r.Method,
			"url":       r.URL.RequestURI(),
		}

		app.Logger.Info(fmt.Sprintf("%s - %s %s %s", properties["ipAddress"], properties["protocol"], properties["method"], properties["url"]), properties)
		next.ServeHTTP(w, r)
	})
}

// Interface for user struct
type user interface {
	getID() int64
	getPermissions() permissions.Permissions
}

// AuthRepository is an interface that defines the repository needed for authentication and authorization
type AuthRepository interface {
	GetByID(ctx context.Context, userID int64) (user, error)
}

// Authenticate is a middleware used to authenticate a user before acessing a certain route.
// It extracts a JWT access token from the Authorization header and validates it.
func (app *App) Authenticate(repository AuthRepository, fileSystem embed.FS, next http.Handler) http.Handler {
	publicKey, err := app.loadRsaPublicKey(fileSystem)
	if err != nil {
		app.Logger.Fatal(err, nil)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no such header found.
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header found, send back a 401 Unauthorized response
		if authorizationHeader == "" {
			app.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into its constituent parts, and if the
		// header isn't in the expected format we return a 401 Unauthorized response
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the actual authentication token from the header parts
		token := headerParts[1]

		// Parse the JWT and extract the claims. This will return an error if the JWT
		// contents doesn't match the signature (i.e. the token has been tampered with)
		// or the algorithm isn't valid.

		claims, err := jwt.RSACheck([]byte(token), publicKey)
		if err != nil {
			app.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Check if the JWT is still valid at this moment in time
		if !claims.Valid(time.Now()) {
			app.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Check that the issuer is our identity service
		if claims.Issuer != app.Config.Authority {
			app.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// Check that the catalog service is in the expected audiences for the JWT
		if !claims.AcceptAudience("http://localhost:4445") {
			app.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		// At this point, we know that the JWT is all OK and we can trust the data in
		// it. We extract the user ID from the claims subject and convert it from a
		// string into an int64.
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		// Retrieve the details of the user associated with the authentication token
		user, err := repository.GetByID(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, database.ErrRecordNotFound):
				app.InvalidAuthenticationTokenResponse(w, r)
			default:
				app.ServerErrorResponse(w, r, err)
			}

			return
		}

		// Call the contextSetUser() helper to add the user information to the request context
		r = app.ContextSetUser(r, user)

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// RequirePermission is a middleware used to check if user has the right permissions to access a certain route
func (app *App) RequirePermission(repository AuthRepository, code string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context
		user := app.ContextGetUser(r)

		// Get the slice of permissions for the user
		user, err := repository.GetByID(r.Context(), user.getID())
		if err != nil {
			app.ServerErrorResponse(w, r, err)
			return
		}

		// Check if the slice includes the required permission. If it doesn't, then
		// return a 403 Forbidden response.
		if !user.getPermissions().Include(code) {
			app.NotPermittedResponse(w, r)
			return
		}

		// Otherwise they have the required permission so we call the next handler in
		// the chain.
		next.ServeHTTP(w, r)
	})
}
