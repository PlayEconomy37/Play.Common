package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PlayEconomy37/Play.Common/types"
	"github.com/PlayEconomy37/Play.Common/validator"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper for sending JSON responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *App) WriteJSON(w http.ResponseWriter, status int, data types.Envelope, headers http.Header) error {
	// Encode the data to JSON
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications
	js = append(js, '\n')

	// Loop through the header map and add each header to the http.ResponseWriter header map
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// Helper for reading JSON data from HTTP request to specified target
func (app *App) ReadJSON(w http.ResponseWriter, r *http.Request, target interface{}) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder and call the DisallowUnknownFields() method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return
	// an error instead of just ignoring the field.
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// Decode the request body into the target destination
	err := decoder.Decode(target)
	if err != nil {
		// If there is an error during decoding, start the triage
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// This error occurs when there are syntax errors in the JSON
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains malformed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error for syntax errors in the JSON
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains malformed JSON")

		//This error occurs when JSON value is of the wrong type for the target destination. If the error relates
		// to a specific field, then we include that in our error message to make it easier for the client to debug
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}

			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"". We check for this, extract the field name from the error
		// and interpolate it into our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")

			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the request body exceeds 1MB in size the decode will now fail with the
		// error "http: request body too large"
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		// A json.InvalidUnmarshalError error will be returned if we pass a non-nil
		// pointer to Decode(). We catch this error and panic.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For anything else, return the error message as-is
		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// Retrieve the URL parameter `id` from the current request context, then convert it to an integer
func (app *App) ReadIdParam(r *http.Request) (int64, error) {
	// Extract URL parameters from request context
	params := chi.URLParamFromCtx(r.Context(), "id")

	// `id` must be a positive integer so we try to convert it
	// If `ìd` cannot be converted to an integer or if it is smaller than 1, throw error
	id, err := strconv.ParseInt(params, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// Retrieve the URL parameter `id` from the current request context, then convert it to an ObjectID
func (app *App) ReadObjectIdParam(r *http.Request) (primitive.ObjectID, error) {
	// Extract URL parameters from request context
	params := chi.URLParamFromCtx(r.Context(), "id")

	// `id` must be an ObjectID so we try to convert it
	objectID, err := primitive.ObjectIDFromHex(params)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return objectID, nil
}

// The ReadStringFromQueryString() helper returns a string value from the query string, or the provided
// default value if no matching key could be found
func (app *App) ReadStringFromQueryString(queryString url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exists this
	// will return the empty string ""
	value := queryString.Get(key)

	// If no key exists (or the value is empty) then return the default value
	if value == "" {
		return defaultValue
	}

	return value
}

// The ReadCsvFromQueryString() helper reads a string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found, it returns
// the provided default value.
func (app *App) ReadCsvFromQueryString(queryString url.Values, key string, defaultValue []string) []string {
	// Extract the value from the query string
	csv := queryString.Get(key)

	// If no key exists (or the value is empty) then return the default value
	if csv == "" {
		return defaultValue
	}

	// Otherwise parse the value into a []string slice and return it
	return strings.Split(csv, ",")
}

// The ReadIntFromQueryString() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.
func (app *App) ReadIntFromQueryString(queryString url.Values, key string, defaultValue int, v *validator.Validator) int {
	// Extract the value from the query string
	str := queryString.Get(key)

	// If no key exists (or the value is empty) then return the default value
	if str == "" {
		return defaultValue
	}

	// Try to convert the value to an int. If this fails, add an error message to the
	// validator instance and return the default value.
	value, err := strconv.Atoi(str)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return value
}

// The ReadFloatFromQueryString() helper reads a string value from the query string and converts it to a
// float64 before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to a float64, then we record an
// error message in the provided Validator instance.
func (app *App) ReadFloatFromQueryString(queryString url.Values, key string, defaultValue float64, v *validator.Validator) float64 {
	// Extract the value from the query string
	str := queryString.Get(key)

	// If no key exists (or the value is empty) then return the default value
	if str == "" {
		return defaultValue
	}

	// Try to convert the value to a float64. If this fails, add an error message to the
	// validator instance and return the default value.
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		v.AddError(key, "must be a float64 value")
		return defaultValue
	}

	return value
}

// Helper function which runs function in a separate go routine and makes sure that we
// recover any panic that happens in the go routine.
// We pass in the context for opentelemetry tracing
func (app *App) Background(ctx context.Context, fn func(ctx context.Context)) {
	// Increment the WaitGroup counter
	app.WaitGroup.Add(1)

	// Launch a background goroutine
	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.WaitGroup.Done()

		// Recover any panic
		defer func() {
			if err := recover(); err != nil {
				app.Logger.Error(fmt.Errorf("%s", err), nil)
			}
		}()

		// Execute the arbitrary function that we passed as the parameter
		fn(ctx)
	}()
}
