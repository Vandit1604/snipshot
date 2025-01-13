package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vandit1604/snipshot/pkg/forms"
	"github.com/vandit1604/snipshot/pkg/models"
)

func (app *app) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := templateData{
		Snippets: snippets,
	}

	// using the cache to render the home page
	app.render(w, r, "home.page.tmpl", &templateData)
}

func (app *app) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	snippet, err := app.snippets.Get(id)

	if err == models.ErrRecordNotFound {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := templateData{
		Snippet: snippet,
	}
	app.render(w, r, "show.page.tmpl", &templateData)
}

func (app *app) createSnippet(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm() has limit of 10MB sent data by default
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	form := forms.New(r.PostForm)
	form.Required("title", "expires", "content")
	form.PermittedValues("expires", "365", "7", "1")
	form.MaxLength("title", 100)

	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// cookie based session management
	app.session.Put(r, "flash", "Snippet successfully created!")

	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}

func (app *app) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *app) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err == models.ErrInvalidCredenetials {
		form.Errors.Add("generic", "Email or Password is incorrect")
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// good idea to never specify what the id is
	app.session.Put(r, "userID", id)
	app.session.Put(r, "flash", "You have been logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{Form: forms.New(nil)})
}

func (app *app) signupUser(w http.ResponseWriter, r *http.Request) {
	// 1. get form data
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
	}
	// 2. validate the form data and display form with errors if any
	// Question: here we are getting the data from the post request but when are we making this post request to this handler?
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{
			Form: form,
		})
		return
	}

	// 3. add user in db if it doesn't exist
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	// Otherwise add a confirmation flash message to the session confirming tha
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")
	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *app) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *app) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "userID")
	app.session.Put(r, "flash", "You have been logged out successfully")
	http.Redirect(w, r, "/", 303)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
