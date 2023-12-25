package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"snippetbox.msarvaro.com/internal/models"
)

type config struct {
	addr      string
	staticDir string
	dsn       string
}

type application struct {
	logger   *slog.Logger
	cfg      config
	snippets *models.SnippetModel
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

	app := &application{
		logger:   logger,
		cfg:      *cfg,
		snippets: &models.SnippetModel{DB: db},
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

	// Starting the server
	logger.Info("starting the server", slog.String("addr", cfg.addr))
	err = http.ListenAndServe("127.0.0.1"+cfg.addr, app.routes())
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
