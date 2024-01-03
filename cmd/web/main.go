package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"snippetbox.msarvaro.com/internal/models"
)

type config struct {
	addr      string
	staticDir string
	dsn       string
}

type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	cfg            config
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// Config by user's options or by default settings
	cfg := &config{}
	flag.StringVar(&cfg.addr, "addr", ":7000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static", "./ui/static", "Directory path for static files")
	flag.StringVar(&cfg.dsn, "dsn", "web:nah@tcp(localhost:4400)/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	// Creating logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Database connection
	db, err := openDB(cfg.dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Pre-parse templates and save them as cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.IdleTimeout = 10 * time.Minute
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		cfg:            *cfg,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		logger.Info("signal received", slog.String("signal", sig.String()))
		db.Close()

		os.Exit(0)
	}()

	// Creating server
	server := &http.Server{
		Addr:     "127.0.0.1" + cfg.addr,
		Handler:  app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Starting server
	logger.Info("starting the server", slog.String("addr", cfg.addr))
	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
