package common

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PlayEconomy37/Play.Common/logger"
)

func TestRecoverPanic(t *testing.T) {
	// Initialize a new httptest.ResponseRecorder and dummy http.Request
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP handler that we can pass to our RecoverPanic
	// middleware, which panics
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("error")
	})

	// Pass the mock HTTP handler to our RecoverPanic middleware. Because
	// RecoverPanic *returns* a http.Handler we can call its ServeHTTP()
	// method, passing in the http.ResponseRecorder and dummy http.Request to
	// execute it.
	app := &App{Logger: logger.New(os.Stdout, logger.LevelInfo)}
	app.RecoverPanic(next).ServeHTTP(rr, r)

	// Call the Result() method on the http.ResponseRecorder to get the results
	// of the test
	rs := rr.Result()

	// Check that the middleware has correctly called the next handler in line
	// and the response status code and body are as expected
	if rs.StatusCode != http.StatusInternalServerError {
		t.Errorf("want %d; got %d", http.StatusInternalServerError, rs.StatusCode)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	errorMsg := "The server encountered a problem and could not process your request"
	if !bytes.Contains(body, []byte(errorMsg)) {
		t.Errorf("want body to contain %q", errorMsg)
	}
}

func TestSecureHeaders(t *testing.T) {
	// Initialize a new httptest.ResponseRecorder and dummy http.Request
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP handler that we can pass to our SecureHeaders
	// middleware, which writes a 200 status code and "OK" response body
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Pass the mock HTTP handler to our SecureHeaders middleware. Because
	// SecureHeaders *returns* a http.Handler we can call its ServeHTTP()
	// method, passing in the http.ResponseRecorder and dummy http.Request to
	// execute it.
	app := &App{}
	app.SecureHeaders(next).ServeHTTP(rr, r)

	// Call the Result() method on the http.ResponseRecorder to get the results
	// of the test
	rs := rr.Result()

	// Check that the middleware has correctly set the X-Frame-Options header
	// on the response
	frameOptions := rs.Header.Get("X-Frame-Options")
	if frameOptions != "deny" {
		t.Errorf("want %q; got %q", "deny", frameOptions)
	}

	// Check that the middleware has correctly set the X-XSS-Protection header
	// on the response
	xssProtection := rs.Header.Get("X-XSS-Protection")
	if xssProtection != "1; mode=block" {
		t.Errorf("want %q; got %q", "1; mode=block", xssProtection)
	}

	// Check that the middleware has correctly called the next handler in line
	// and the response status code and body are as expected
	if rs.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q, got %q", "OK", string(body))
	}
}

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}

	// Redirect STDOUT to a buffer
	rFile, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Failed to redirect STDOUT")
	}

	os.Stdout = w

	go func() {
		scanner := bufio.NewScanner(rFile)

		for {
			if !scanner.Scan() {
				break
			}

			buf.WriteString(scanner.Text())
		}
	}()

	// Initialize a new httptest.ResponseRecorder and dummy http.Request
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock HTTP handler that we can pass to our secureHeaders
	// middleware, which writes a 200 status code and "OK" response body
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Pass the mock HTTP handler to our LogRequest middleware. Because
	// LogRequest *returns* a http.Handler we can call its ServeHTTP()
	// method, passing in the http.ResponseRecorder and dummy http.Request to
	// execute it.
	app := &App{Logger: logger.New(os.Stdout, logger.LevelInfo)}
	app.LogRequest(next).ServeHTTP(rr, r)

	// Call the Result() method on the http.ResponseRecorder to get the results
	// of the test
	rs := rr.Result()

	// Check that the middleware has correctly called the next handler in line
	// and the response status code and body are as expected
	if rs.StatusCode != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q, got %q", "OK", string(body))
	}

	// Reset output
	w.Close()

	if buf.Len() == 0 {
		t.Error("No information logged to STDOUT")
	}

	if strings.Count(buf.String(), "\n") > 1 {
		t.Error("Expected only a single line of log output")
	}

	if !strings.Contains(buf.String(), "/users") {
		t.Error("Expected url path to be in log")
	}
}
