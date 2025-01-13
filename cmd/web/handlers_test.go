package main

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
)

// here we were testing if the server is working fine but to automate this process i have wrote another function
// func TestPing(t *testing.T) {
// 	// the ResponseRecorder is like ResponseWriter but it just records and doesn't writes on the HTTP connection
// 	// 	the httptest.ResponseRecorder type. This is
// 	// essentially an implementation of http.ResponseWriter which records
// 	// the response status code, headers and body instead of actually writing
// 	// them to a HTTP connection.
// 	rr := httptest.NewRecorder()
//
// 	// creating a dummy request to test the ping
// 	r, err := http.NewRequest("GET", "/", nil)
//
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ping(rr, r)
//
// 	if rr.Result().StatusCode != http.StatusOK {
// 		t.Fatalf("want %d, got %d", http.StatusOK, rr.Result().StatusCode)
// 	}
//
// 	defer rr.Result().Body.Close()
//
// 	body, err := io.ReadAll(rr.Result().Body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if string(body) != "OK" {
// 		t.Errorf("want body equal to %s", "OK")
// 	}
//
// }

// func TestPing(t *testing.T) {
// 	app := &app{
// 		errorLog: log.New(io.Discard, "", 0),
// 		infoLog:  log.New(io.Discard, "", 0),
// 	}
//
// 	// We then use the httptest.NewTLSServer() function to create a new test
// 	// server, passing in the value returned by our app.routes() method as the
// 	// handler for the server. This starts up a HTTPS server which listens on a
// 	// randomly-chosen port of your local machine for the duration of the test.
// 	// Notice that we defer a call to ts.Close() to shutdown the server when
// 	// the test finishes.
// 	ts := httptest.NewTLSServer(app.setupRoutes())
// 	defer ts.Close()
//
// 	// The network address that the test server is listening on is contained
// 	// in the ts.URL field. We can use this along with the ts.Client().Get()
// 	// method to make a GET /ping request against the test server. This
// 	// returns a http.Response struct containing the response.
// 	rs, err := ts.Client().Get(ts.URL + "/ping")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// We can then check the value of the response status code and body using
// 	// the same code as before.
// 	if rs.StatusCode != http.StatusOK {
// 		t.Errorf("want %d; got %d", http.StatusOK, rs.StatusCode)
// 	}
// 	defer rs.Body.Close()
// 	body, err := io.ReadAll(rs.Body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if string(body) != "OK" {
// 		t.Errorf("want body to equal %q", "OK")
// 	}
// }

func TestPing(t *testing.T) {
	t.Parallel()

	testApp := newTestApplication(t)
	mux := testApp.setupRoutes()
	ts := newTestServer(mux)
	defer ts.Close()

	code, _, body := ts.get(t, "/healthcheck")

	if code != http.StatusOK && string(body) != "OK" {
		t.Fatalf("want %d:{%s}, got %d:{%s}", http.StatusOK, "OK", code, string(body))
	}
}

func TestShowSnippet(t *testing.T) {
	testApp := newTestApplication(t)
	mux := testApp.setupRoutes()
	ts := newTestServer(mux)
	defer ts.Close()

	// Set up some table-driven tests to check the responses sent by our
	// application for different URLs.
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{"Valid ID", "/snippet/1", http.StatusOK, []byte("An old silent pond...")},
		{"Non-existent ID", "/snippet/2", http.StatusNotFound, nil},
		{"Negative ID", "/snippet/-1", http.StatusNotFound, nil},
		{"Decimal ID", "/snippet/1.23", http.StatusNotFound, nil},
		{"String ID", "/snippet/foo", http.StatusNotFound, nil},
		{"Empty ID", "/snippet/", http.StatusNotFound, nil},
		{"Trailing slash", "/snippet/1/", http.StatusNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}

func TestSignUpUser(t *testing.T) {
	testApp := newTestApplication(t)
	mux := testApp.setupRoutes()
	ts := newTestServer(mux)
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)
	t.Log(csrfToken)
}

// TestSignupUser tests that signupUser handler returns appropriate status codes and error messages
// corresponding logic of signupUser handler.
func TestSignupUser(t *testing.T) {
	t.Parallel()
	// Create the application struct containing our mocked dependencies and
	// set up the test server for running an end-to-test.
	app := newTestApplication(t)
	ts := newTestServer(app.setupRoutes())
	defer ts.Close()

	// Make a GET /user/signup request and then extract the CSRF token from the
	// response body.
	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{"Valid submission", "Bob", "bob@example.com", "validPa$$word", csrfToken,
			http.StatusSeeOther, nil},
		{"Empty name", "", "bob@example.com", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("This field cannot be blank")},
		{"Empty email", "Bob", "", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("This field cannot be blank")},
		{"Empty password", "Bob", "bob@example.com", "", csrfToken, http.StatusOK,
			[]byte("This field cannot be blank")},
		{"Invalid email (incomplete domain)", "Bob", "bob@example.", "validPa$$word",
			csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing @)", "Bob", "bobexample.com", "validPa$$word", csrfToken,
			http.StatusOK, []byte("This field is invalid")},
		{"Invalid email (missing local part)", "Bob", "@example.com", "validPa$$word",
			csrfToken, http.StatusOK, []byte("This field is invalid")},
		{"Short password", "Bob", "bob@example.com", "pa$$word", csrfToken, http.StatusOK,
			[]byte("This field is too short (minimum is 10 characters")},
		{"Duplicate email", "Bob", "dupe@example.com", "validPa$$word", csrfToken, http.StatusOK,
			[]byte("Address is already in use")},
		{"Invalid CSRF Token", "", "", "", "wrongToken", http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)
			t.Logf("testing %q for want-code %d and want-body %q", tt.name, tt.wantCode,
				tt.wantBody)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}

			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q, but got %q", tt.wantBody, body)
			}
		})
	}
}
