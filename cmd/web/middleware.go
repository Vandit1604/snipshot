package main

import (
	"fmt"
	"net/http"
)

type Middleware struct{}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		// In any middleware handler, code which comes before next.ServeHTTP() will be executed on the way down the chain, and any code after next.ServeHTTP() — or in a deferred function — will be executed on the way back up.

		// func myMiddleware(next http.Handler) http.Handler {
		// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 		// Any code here will execute on the way down the chain.
		// 		next.ServeHTTP(w, r)
		// 		// Any code here will execute on the way back up the chain.
		// 	})
		// }

		// secureHeaders → servemux → application handler → servemux → secureHeaders
		next.ServeHTTP(w, r)
	})
}

func (app *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.UserAgent())
		next.ServeHTTP(w, r)
	})
}

func (app *app) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if in app, there's any panic in our handlers we check that via inbuilt recover() function when the middleware request returns after the request is served.
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
