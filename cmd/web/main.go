package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vandit1604/snipshot/pkg/models/mysql"
)

type app struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	// cli flags
	addr := flag.String("addr", ":4000", "Host Port")
	dsn := flag.String("dsn", "web:qwe@/snippetbox?parseTime=true", "Connection string for MySQL Database")
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

	// SnippetModel
	app := &app{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: cache,
	}

	mux := app.setupRoutes()

	srv := http.Server{
		Addr:     *addr,
		Handler:  mux,
		ErrorLog: errorLog,
	}

	infoLog.Println("Starting server on", *addr)
	err = srv.ListenAndServe()

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
