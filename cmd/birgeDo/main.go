package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/golangcollege/sessions"
	"github.com/jumagaliev1/birgeDo/internal/data"
	"github.com/jumagaliev1/birgeDo/internal/jsonlog"
	_ "github.com/lib/pq"
	"html/template"
	"net/http"
	"os"
	"time"
)

type contextKey string

var contextKeyUser = contextKey("user")

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}
type application struct {
	config        config
	logger        *jsonlog.Logger
	models        data.Models
	session       *sessions.Session
	templateCache map[string]*template.Template
	//users         interface {
	//	Insert(string, string, string) error
	//	Authenticate(string, string) (int, error)
	//	Get(int) (*models.User, error)
	//}
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "Server port")

	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://birgedo:password@localhost/birgedo", "PostgreSQL DSN")

	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		logger.PrintError(err, nil)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	//session.Secure = true

	app := &application{
		logger:        logger,
		config:        cfg,
		models:        data.NewModels(db),
		templateCache: templateCache,
		session:       session,
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		<-ticker.C
		err = app.models.Task.ResetAllTasks()
		if err != nil {
			app.logger.PrintError(err, nil)
			ticker.Stop()
			return
		}
		app.logger.PrintInfo("Success Reset All Tasks", nil)
	}()
	logger.PrintInfo(fmt.Sprintf("Starting server on %d", cfg.port), nil)
	err = srv.ListenAndServe()
	logger.PrintError(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {
	// sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// Context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an
	// error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	// Return the sql.DB connection pool.
	return db, nil
}
