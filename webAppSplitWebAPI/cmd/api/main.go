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
	"strconv"
	"sync"
	"syscall"
	"time"
	"webProj/internal/encryption"
	"webProj/internal/models"
	"webProj/internal/service"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type config struct {
	host       		string
	port       		int
	env        		string
	csrfSecret 		string
	secretkey		string
	frontend		string
	db         struct {
		dsn 		string
	}
	tls struct {
		cert 		string
		key  		string
	}
	smtp struct {
		host 		string
		port		int
		username	string
		password	string
		sender		string
	}
}

type application struct {
	ctx      	        context.Context
	config   	        config
	infoLog  	        *log.Logger
	errorLog 	        *log.Logger
	version  	        string
	db       	        *models.Models
	services 	        *service.Services
	enc    		        *encryption.Encryption
	mailer 		        *WorkQueue[emailMessage]
	wg     		        *sync.WaitGroup
	emailTemplateCache  map[string]*template.Template
	emailTemplateMu     sync.RWMutex
}

func (app *application) serve() error {

	// FIX: slowloris protection? it's fine, can get better?
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", app.config.host, app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	// graceful shutdown in background
	go func() {

		<-app.ctx.Done()
		app.infoLog.Println("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Close the mailer queue first so in-flight handlers get an error
		// instead of blocking on a full channel during shutdown
		app.mailer.Close()

		srv.Shutdown(shutdownCtx)

		// Wait for the background worker to finish processing remaining emails
		app.wg.Wait()
		app.infoLog.Println("background tasks finished")
	}()

	if app.config.tls.cert != "" && app.config.tls.key != "" {

		app.infoLog.Printf("Starting backend HTTPS server in %s mode on port %d\n", app.config.env, app.config.port)
		return srv.ListenAndServeTLS(app.config.tls.cert, app.config.tls.key)
	}

	app.infoLog.Printf("Starting backend HTTP server in %s mode on port %d\n", app.config.env, app.config.port)
	return srv.ListenAndServe()
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	envErr := godotenv.Load()
	if envErr != nil {

		log.Fatal(envErr)
	}

	var cfg config

	cfg.host = os.Getenv("APP_HOST")
	cfg.csrfSecret = os.Getenv("CSRF_SECRET")
	apiPort := os.Getenv("API_PORT")
	var parseErr error
	cfg.port, parseErr = strconv.Atoi(apiPort)
	if parseErr != nil || cfg.port == 0 {

		log.Fatalf("invalid API_PORT %q: %v", apiPort, parseErr)
	}
	cfg.frontend = os.Getenv("FRONTEND_URL")

	cfg.env = os.Getenv("APP_ENV")
	cfg.db.dsn = os.Getenv("DB_DSN")

	cfg.smtp.host =  os.Getenv("MAILTRAP_HOST")
	cfg.smtp.port, parseErr = strconv.Atoi(os.Getenv("MAILTRAP_PORT"))
	if parseErr != nil {

		log.Fatalf("invalid MAILTRAP_PORT: %v", parseErr)
	}
	cfg.smtp.username = os.Getenv("MAILTRAP_USERNAME")
	cfg.smtp.password = os.Getenv("MAILTRAP_PASSWORD")
	cfg.smtp.sender = os.Getenv("SENDER_EMAIL")
	if cfg.smtp.sender == "" {
		cfg.smtp.sender = "info@dedu.com"
	}

	// FIX: consider key separation
	cfg.secretkey = os.Getenv("SECRET_KEY")

	flag.StringVar(&cfg.host, "host", cfg.host, "Server host")
	flag.IntVar(&cfg.port, "port", cfg.port, "Server port")
	flag.StringVar(&cfg.env, "env", cfg.env, "Application environment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", cfg.db.dsn, "DSN")
	flag.StringVar(&cfg.frontend, "frontend", cfg.frontend, "URL to frontend")

	flag.Parse()

	cfg.tls.cert = os.Getenv("TLS_CERT")
	cfg.tls.key = os.Getenv("TLS_KEY")

	conn, err := models.OpenDB(cfg.db.dsn)
	if err != nil {

		errorLog.Fatal(err)
	}
	defer conn.Close()

	db := models.NewModels(conn)

	enc, err := encryption.New([]byte(cfg.secretkey))
	if err != nil {

		errorLog.Fatal(err)
	}

	app := &application{
		ctx:      ctx,
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		db:       db,
		services: service.NewServices(db),
		enc:                enc,
		mailer:             NewWorkQueue[emailMessage](100),
		wg:                 &sync.WaitGroup{},
		emailTemplateCache: make(map[string]*template.Template),
	}

	// listen to services for shutdown, if we have 5 email services increase to 5 for example
	app.wg.Add(1)
	go app.listenForMail()

	err = app.serve()
	if err != nil && err != http.ErrServerClosed {

		log.Fatal(err)
	}
}
