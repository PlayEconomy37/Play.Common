package common

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

// Acts as the holding structure for any dynamic data that we want to pass to our HTML templates
// This should be defined by the application using the templates and not here.
type TemplateData struct {
	CurrentYear int
}

// Create a new template cache
func NewTemplateCache(dir string, functions template.FuncMap) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache
	cache := map[string]*template.Template{}

	// Use the `filepath.Glob` function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the `name` variable
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before
		// calling the `ParseFiles()` method. This means we have to use template.New() to
		// create an empty template set, use the Funcs() method to register the
		// template.FuncMap, and then parse the file into the template set
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Use the `ParseGlob` method to add any 'layout' templates to the
		// template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// Use the `ParseGlob` method to add any 'partial' templates to the
		// template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key
		cache[name] = ts
	}

	return cache, nil
}

// Adds default data to a template
// This should be defined by the application using the templates and not here.
func (app *App) AddDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	if td == nil {
		td = &TemplateData{}
	}

	td.CurrentYear = time.Now().Year()

	return td
}

// Renders the specified template
func (app *App) Render(w http.ResponseWriter, r *http.Request, templateCache map[string]*template.Template, name string, td *TemplateData) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl')
	ts, ok := templateCache[name]
	if !ok {
		app.ServerErrorResponse(w, r, fmt.Errorf("the template %s does not exist", name))
		return
	}

	buf := new(bytes.Buffer)

	// Write the template set to the buffer, instead of straight to the http.ResponseWriter.
	// This prevents certain runtime errors caused by mistakes when passing dynamic data from ocurring.
	err := ts.Execute(buf, app.AddDefaultData(td, r))
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}

	// Write the contents of the buffer to the http.ResponseWriter
	buf.WriteTo(w)
}
