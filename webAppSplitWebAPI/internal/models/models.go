package models

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// FIX: for search on many books/records fast implement a virtual FTS5 table (sqlite) or tsvector (postresql)

const dbTimeout = 5 * time.Second

// context, cancels every request if the http request done/5 seconds pass (useful to not waste resources for a query when user cloases the tab for example)
// use context.Background() for "run forever, no way to cancel"
func dbCtx(ctx context.Context) (context.Context, context.CancelFunc) {

	return context.WithTimeout(ctx, dbTimeout)
}

type DBTX interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type SearchResult struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	URL  string `json:"url"`
	Rank int    `json:"-"`
}

type Searchable interface {

	Search(ctx context.Context, query string) ([]*SearchResult, error)
}

type Models struct {
	db                *sql.DB
	Token             TokenModel
	Book              BookModel
	User              UserModel
}

func NewModels(db *sql.DB) *Models {

	return &Models{
		db:                db,
		Token:             TokenModel{db: db},
		Book:              BookModel{db: db},
		User:              UserModel{db: db},
	}
}

func (model *Models) Searchables() []Searchable {

	return []Searchable{&model.Book}
}

func (model *Models) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {

	return model.db.BeginTx(ctx, opts)
}

func OpenDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {

		return nil, err
	}

	if err = db.Ping(); err != nil {

		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
