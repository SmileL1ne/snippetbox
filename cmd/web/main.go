package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
)

type config struct {
	addr      string
	staticDir string
}

type application struct {
	logger *slog.Logger
}

func main() {
	cfg := &config{}
	flag.StringVar(&cfg.addr, "addr", ":7000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static", "./ui/static", "Directory path for static files")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		logger: logger,
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(cfg.staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet/view", app.snippetView)
	mux.HandleFunc("/snippet/create", app.snippetCreate)

	logger.Info("starting the server", slog.String("addr", cfg.addr))
	err := http.ListenAndServe("127.0.0.1"+cfg.addr, mux)
	logger.Error(err.Error())
	os.Exit(1)
}
