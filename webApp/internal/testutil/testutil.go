package testutil

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/tokens"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// NewTestDB opens an in-memory SQLite database with the full production schema.
func NewTestDB(t *testing.T) *models.Models {

	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {

		t.Fatal(err)
	}
	db.SetMaxOpenConns(1) // in-memory SQLite requires single connection
	t.Cleanup(func() { db.Close() })

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			first_name TEXT NOT NULL DEFAULT '',
			last_name TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			password TEXT NOT NULL DEFAULT '',
			role TEXT NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT 0,
			profile_image TEXT NOT NULL DEFAULT '',
			uuid TEXT NOT NULL DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX idx_users_name ON users (last_name, first_name)`,
		`CREATE UNIQUE INDEX idx_users_email ON users (email)`,
		`CREATE INDEX idx_users_role ON users (role)`,
		`CREATE INDEX idx_users_status ON users (is_active)`,
		`CREATE UNIQUE INDEX idx_users_uuid ON users (uuid)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash BLOB NOT NULL,
			expiry DATETIME NOT NULL,
			scope TEXT NOT NULL DEFAULT 'authentication',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			image TEXT DEFAULT '',
			uuid TEXT NOT NULL DEFAULT (lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6)))),
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX idx_books_uuid ON books (uuid)`,
	}

	for _, stmt := range stmts {

		if _, err := db.Exec(stmt); err != nil {

			t.Fatalf("schema exec failed: %v", err)
		}
	}

	return models.NewModels(db)
}

// CreateUser inserts a user with the given credentials and returns the full record.
// Uses bcrypt cost 4 for speed.
func CreateUser(t *testing.T, m *models.Models, email, password, role string, isActive bool) *models.User {

	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {

		t.Fatal(err)
	}

	user := &models.User{
		FirstName: "Test",
		LastName:  strings.Split(email, "@")[0],
		Email:     email,
		Password:  string(hash),
		Role:      role,
		IsActive:  isActive,
	}

	id, err := m.User.Insert(t.Context(), user)
	if err != nil {

		t.Fatalf("create user insert: %v", err)
	}

	created, err := m.User.Get(t.Context(), id)
	if err != nil {

		t.Fatalf("create user get: %v", err)
	}

	return created
}

// CreateUsers inserts n active regular users with emails user1@test.com, user2@test.com, etc.
func CreateUsers(t *testing.T, m *models.Models, n int) []*models.User {

	t.Helper()

	users := make([]*models.User, 0, n)
	for i := 1; i <= n; i++ {

		u := CreateUser(t, m, fmt.Sprintf("user%d@test.com", i), "password123", models.RoleUser, true)
		users = append(users, u)
	}

	return users
}

// CreateMixedUsers inserts n users with varied roles and active states.
// Every 10th user is admin, every 5th user is inactive.
func CreateMixedUsers(t *testing.T, m *models.Models, n int) []*models.User {

	t.Helper()

	users := make([]*models.User, 0, n)
	for i := 1; i <= n; i++ {

		role := models.RoleUser
		if i%10 == 0 {
			role = models.RoleAdmin
		}

		isActive := i%5 != 0

		u := CreateUser(t, m, fmt.Sprintf("mixed%d@test.com", i), "password123", role, isActive)
		users = append(users, u)
	}

	return users
}

// NewTestServices creates a Services instance backed by an in-memory test DB.
func NewTestServices(t *testing.T) (*service.Services, *models.Models) {

	t.Helper()

	m := NewTestDB(t)
	return service.NewServices(m), m
}

// CreateTestPNG generates a minimal 10x10 red PNG image for upload tests.
func CreateTestPNG(t *testing.T) []byte {

	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := range 10 {
		for y := range 10 {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {

		t.Fatal(err)
	}
	return buf.Bytes()
}

// CreateUserAndToken creates a user and inserts a valid auth token for them.
func CreateUserAndToken(t *testing.T, m *models.Models) (*models.User, *tokens.Token) {

	t.Helper()

	user := CreateUser(t, m, "token@test.com", "password123", models.RoleUser, true)

	token, err := tokens.GenerateAuthToken(user.ID, 24*time.Hour, tokens.ScopeAuthentication)
	if err != nil {
		t.Fatalf("generate auth token: %v", err)
	}

	err = m.Token.Insert(t.Context(), token, user)
	if err != nil {
		t.Fatalf("insert token: %v", err)
	}

	return user, token
}

// CreateBook inserts a book and returns the full record.
func CreateBook(t *testing.T, m *models.Models, name, description, image string) *models.Book {

	t.Helper()

	book := &models.Book{
		Name:        name,
		Description: description,
		Image:       image,
	}

	id, err := m.Book.Insert(t.Context(), book)
	if err != nil {

		t.Fatalf("create book insert: %v", err)
	}

	created, err := m.Book.Get(t.Context(), id)
	if err != nil {

		t.Fatalf("create book get: %v", err)
	}

	return created
}

// CreateBooks inserts n books with names "Book 1", "Book 2", etc.
func CreateBooks(t *testing.T, m *models.Models, n int) []*models.Book {

	t.Helper()

	books := make([]*models.Book, 0, n)
	for i := 1; i <= n; i++ {

		b := CreateBook(t, m, fmt.Sprintf("Book %d", i), fmt.Sprintf("Description %d", i), "")
		books = append(books, b)
	}

	return books
}
