package main

import (
	"html/template"
	"net/url"
	"path/filepath"
	"time"

	"github.com/vandit1604/snipshot/pkg/models"
)

type templateData struct {
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
	CurrentYear int
	FormErrors  map[string]string
	FormData    url.Values
}

func NewTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// the order of template files parsed matters. Parse page files first, then layout files and then partials or blocks.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// we created an empty template set and registered the functions and parsed the files into the new template set
		ts, err := template.New(name).Funcs(templateFunctions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// adding the layout after that
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// adding the partials after that
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// the order is important since page depends on layout and after that partials are added in the page in the same order
		cache[name] = ts
	}

	return cache, nil
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var templateFunctions = template.FuncMap{
	"humanDate": humanDate,
}
