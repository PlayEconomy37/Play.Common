package common

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PlayEconomy37/Play.Common/opentelemetry"
	"github.com/felixge/httpsnoop"
)

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

func (app *App) Metrics(appName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Create Prometheus metrics
		prometheusMetrics := opentelemetry.CreateMetrics(appName)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Increment the number of requests received by 1
			prometheusMetrics.TotalRequestsCounter.WithLabelValues(r.Method, r.URL.Path).Inc()

			// This function wraps a http.Handler (in this case, the next function), executes the handler and then returns a Metrics struct
			metrics := httpsnoop.CaptureMetrics(next, w, r)

			// On the way back up the middleware chain, increment the number of responses sent by 1
			prometheusMetrics.TotalResponsesCounter.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(metrics.Code)).Inc()

			// Get the request processing time in microseconds from httpsnoop and increment
			// the cumulative processing time
			prometheusMetrics.TotalProcessingTimeCounter.WithLabelValues(r.Method, r.URL.Path).Observe(float64(metrics.Duration.Microseconds()))
		})
	}
}
