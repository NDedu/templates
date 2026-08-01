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

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

// append to any css/js, changing version will save the trouble of clearing cache
const cssVersion = "1"


type config struct {
	host       	string
	port       	int
	env        	string
	api        	string
	csrfSecret 	string
	secretkey	string
	frontend	string
	db struct {
		dsn 	string
	}
	tls struct {
		cert 	string
		key  	string
	}
}

type application struct {
	ctx           context.Context
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	templateMu    sync.RWMutex
	version       string
	DB            models.Models
	Session       *scs.SessionManager
	wsHub         wsHub
	wsChan        chan WsPayload
	wsUpgrader    websocket.Upgrader
	staticPath    string
}

func (app *application) serve() error {

	srv := &http.Server{
		Addr:              	fmt.Sprintf("%s:%d", app.config.host, app.config.port),
		Handler:           	app.routes(),
		IdleTimeout:       	30 * time.Second,
		ReadTimeout:       	10 * time.Second,
		ReadHeaderTimeout: 	10 * time.Second,
		// WriteTimeout disabled: WebSocket connections handle write deadlines per-write via SetWriteDeadline
		WriteTimeout: 0,
	}

	go func() {

		<-app.ctx.Done()
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

	//gob.Register(TransactionData{})

	// remember to have .env file with variables set
	envErr := godotenv.Load()
	if envErr != nil {

		log.Fatal(envErr)
	}

	var cfg config

	cfg.host = os.Getenv("APP_HOST")
	cfg.csrfSecret = os.Getenv("CSRF_SECRET")
	portStr := os.Getenv("WEB_PORT")
	var parseErr error
	cfg.port, parseErr = strconv.Atoi(portStr)
	if parseErr != nil || cfg.port == 0 {
		log.Fatalf("invalid WEB_PORT %q: %v", portStr, parseErr)
	}
	cfg.frontend = os.Getenv("FRONTEND_URL")
	cfg.secretkey = os.Getenv("SECRET_KEY")

	apiPort := os.Getenv("API_PORT")
	cfg.api = fmt.Sprintf("http://%s:%s", cfg.host, apiPort)

	cfg.env = os.Getenv("APP_ENV")

	cfg.db.dsn = os.Getenv("DB_DSN")

	flag.StringVar(&cfg.host, "host", cfg.host, "Server host")
	flag.IntVar(&cfg.port, "port", cfg.port, "Server port")
	flag.StringVar(&cfg.env, "env", cfg.env, "Application environment {development|production}")
	flag.StringVar(&cfg.api, "api", cfg.api, "URL to api")
	flag.StringVar(&cfg.db.dsn, "dsn", cfg.db.dsn, "DNS")

	flag.Parse()

	cfg.tls.cert = os.Getenv("TLS_CERT")
	cfg.tls.key = os.Getenv("TLS_KEY")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := models.OpenDB(cfg.db.dsn)
	if err != nil {

		errorLog.Fatal(err)
	}
	defer conn.Close()

	// set session
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	templateCache := make(map[string]*template.Template)

	staticPath, err := filepath.Abs("./static")
	if err != nil {

		errorLog.Fatal("could not resolve static path:", err)
	}

	app := application{
		ctx:           ctx,
		config:        cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: templateCache,
		version:       version,
		DB:            *models.NewModels(conn),
		Session:       sessionManager,
		wsHub: wsHub{
			clients: make(map[*WebSocketConnection]bool),
		},
		wsChan: make(chan WsPayload, 100),
		staticPath: staticPath,
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return r.Header.Get("Origin") == cfg.frontend
			},
		},
	}

	go app.ListenToWsChannel()

	err = app.serve()
	if err != nil {

		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
