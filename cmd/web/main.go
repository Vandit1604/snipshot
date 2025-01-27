package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
	"github.com/vandit1604/snipshot/pkg/models"
	"github.com/vandit1604/snipshot/pkg/models/mysql"
)

type contextKey string

var contextKeyUser = contextKey("user")

type app struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session
	templateCache map[string]*template.Template
	// during testing this will complain when creating a mock for the mock app instance. That's why we created this as a interface which contains both the functions which are defined in mock package.
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	}
	users interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
	}
}

func main() {
	// cli flags
	addr := flag.String("addr", ":4000", "Host Port")
	dsn := flag.String("dsn", "web:qwe@/snippetbox?parseTime=true", "Connection string for MySQL Database")
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "JWT secret key")
	flag.Parse()

	// loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// DB
	db, err := OpenDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	// templateSet cache
	cache, err := NewTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	// session // here i have passed the pointer deference which gives the value
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	// to mitigate csrf attacks
	session.SameSite = http.SameSiteStrictMode

	// SnippetModel
	app := &app{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &mysql.SnippetModel{DB: db},
		users:         &mysql.UserModel{DB: db},
		templateCache: cache,
		session:       session,
	}

	mux := app.setupRoutes()

	// tls config
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: false,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	srv := http.Server{
		Addr:         *addr,
		Handler:      mux,
		ErrorLog:     errorLog,
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	errorLog.Fatal(err)
}

func OpenDB(dsn *string) (*sql.DB, error) {
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
