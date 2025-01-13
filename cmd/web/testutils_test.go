package main

import (
	"html"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/vandit1604/snipshot/pkg/models/mock"
)

// Create a newTestApplication helper which returns an instance of our
// application struct containing mocked dependencies.
func newTestApplication(t *testing.T) *app {
	// Create an instance of the template cache.
	templateCache, err := NewTemplateCache("./../../ui/html/")
	if err != nil {
		t.Fatal(err)
	}
	// Create a session manager instance, with the same settings as production.
	session := sessions.New([]byte("3dSm5MnygFHh7XidAtbskXrjbwfoJcbJ"))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	// Initialize the dependencies, using the mocks for the loggers and
	// database models.
	return &app{
		errorLog:      log.New(io.Discard, "", 0),
		infoLog:       log.New(io.Discard, "", 0),
		session:       session,
		templateCache: templateCache,
		snippets:      &mock.SnippetModel{},
		users:         &mock.UserModel{},
	}
}

// Mock server for E2E testing
type testServer struct {
	*httptest.Server
}

// returns a new instance of test server
func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	// initialize a cookie jar, which stores the cookies which goes with subsequent requests
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	// assigning the jar to the clients Jar field
	ts.Client().Jar = jar

	// to avoid redirecting multiple times, if the client is redirected once it returns back with whatever response it gets. This happens on 3XX responses
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{
		Server: ts,
	}
}

// make a get request to the given url
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	resp, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, body

}

var csrfTokenRx = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

// we're using `html.UnescapeString` because html/template package in go will escape the CSRF token's `+` symbol to `&#43;`.
func extractCSRFToken(t *testing.T, body []byte) string {
	matches := csrfTokenRx.FindSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}

// postForm sends POST requests to the test server. The final parameter to this method is a
// url.Values object which can contain any data that you want to send in the request body.
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header,
	[]byte) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body.
	defer func() {
		if err := rs.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Return the response status, headers, and body.
	return rs.StatusCode, rs.Header, body
}
