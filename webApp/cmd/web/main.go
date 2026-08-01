package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
	"webProj/internal/models"
	"webProj/internal/service"

	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

// append to any css/js, changing version will save the trouble of clearing cache
const cssVersion = "1"

type config struct {
	host       string
	port       int
	env        string
	csrfSecret string
	secretkey  string
	db         struct {
		dsn string
	}
	tls struct {
		cert string
		key  string
	}
}

type application struct {
	ctx           context.Context
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	version       string
	db            *models.Models
	services      *service.Services
	templateCache map[string]*template.Template
	templateMu    sync.RWMutex
	Session       *scs.SessionManager
	staticPath    string
}

func (app *application) serve() error {

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", app.config.host, app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {

		<-app.ctx.Done()
		app.infoLog.Println("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		srv.Shutdown(shutdownCtx)
	}()

	if app.config.tls.cert != "" && app.config.tls.key != "" {

		app.infoLog.Printf("Starting HTTPS server in %s mode on port %d\n", app.config.env, app.config.port)
		return srv.ListenAndServeTLS(app.config.tls.cert, app.config.tls.key)
	}

	app.infoLog.Printf("Starting HTTP server in %s mode on port %d\n", app.config.env, app.config.port)
	return srv.ListenAndServe()
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// FIX: move to structured logging (slog) with a request id on every line before production
	// stdlib log has no request correlation, so a prod error is hard to trace back to its request
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	envErr := godotenv.Load()
	if envErr != nil {

		log.Fatal(envErr)
	}

	var cfg config

	cfg.host = os.Getenv("APP_HOST")

	// FIX: in production load secrets from the environment or a secret manager, not a committed .env
	// the .env values are placeholders — if they ship as-is, anyone can forge CSRF/auth tokens
	cfg.csrfSecret = os.Getenv("CSRF_SECRET")
	portStr := os.Getenv("APP_PORT")
	var parseErr error
	cfg.port, parseErr = strconv.Atoi(portStr)
	if parseErr != nil || cfg.port == 0 {

		log.Fatalf("invalid APP_PORT %q: %v", portStr, parseErr)
	}

	cfg.env = os.Getenv("APP_ENV")
	cfg.db.dsn = os.Getenv("DB_DSN")

	// FIX: SECRET_KEY is loaded here but never used anywhere — either wire it up or remove it
	// dead config is misleading, someone will assume it protects something when it does not
	cfg.secretkey = os.Getenv("SECRET_KEY")

	flag.StringVar(&cfg.host, "host", cfg.host, "Server host")
	flag.IntVar(&cfg.port, "port", cfg.port, "Server port")
	flag.StringVar(&cfg.env, "env", cfg.env, "Application environment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", cfg.db.dsn, "DSN")

	flag.Parse()

	cfg.tls.cert = os.Getenv("TLS_CERT")
	cfg.tls.key = os.Getenv("TLS_KEY")

	conn, err := models.OpenDB(cfg.db.dsn)
	if err != nil {

		errorLog.Fatal(err)
	}
	defer conn.Close()

	db := models.NewModels(conn)

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	staticPath, err := filepath.Abs("./static")
	if err != nil {

		errorLog.Fatal("could not resolve static path:", err)
	}

	app := &application{
		ctx:           ctx,
		config:        cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		version:       version,
		db:            db,
		services:      service.NewServices(db),
		templateCache: make(map[string]*template.Template),
		Session:       sessionManager,
		staticPath:    staticPath,
	}

	err = app.serve()
	if err != nil && err != http.ErrServerClosed {

		log.Fatal(err)
	}
}
