package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *app) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *app) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *app) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *app) render(w http.ResponseWriter, r *http.Request, pageName string, data *templateData) {
	ts, ok := app.templateCache[pageName]
	if !ok {
		app.serverError(w, fmt.Errorf("template set not found in cache with name %s", pageName))
		return
	}

	// adding default data like CurrentYear to each render
	data = app.addDefaultData(data, r)

	// init a buffer
	buf := new(bytes.Buffer)

	err := ts.Execute(buf, data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// write to buffer if there's no error
	buf.WriteTo(w)
}

func (app *app) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}
