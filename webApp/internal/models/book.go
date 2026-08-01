package models

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type BookModel struct {
	db DBTX
}

type Book struct {
	ID          int       `json:"id"`
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (model *BookModel) Search(ctx context.Context, query string) ([]*SearchResult, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	escaped := escapeLike(query)

	sql := `
		SELECT uuid, name,
			CASE WHEN LOWER(name) LIKE LOWER(?) || '%' ESCAPE '\' THEN 0
			     WHEN LOWER(name) LIKE '%' || LOWER(?) || '%' ESCAPE '\' THEN 1
			     ELSE 2
			END AS rank
		FROM books
		WHERE LOWER(name) LIKE '%' || LOWER(?) || '%' ESCAPE '\'
		   OR LOWER(description) LIKE '%' || LOWER(?) || '%' ESCAPE '\'
		ORDER BY rank, name
		LIMIT 10
	`

	rows, err := model.db.QueryContext(ctx, sql, escaped, escaped, escaped, escaped)
	if err != nil {

		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {

		var r SearchResult
		var uuid string
		err := rows.Scan(&uuid, &r.Name, &r.Rank)
		if err != nil {

			return nil, err
		}

		r.Kind = "book"
		r.URL = fmt.Sprintf("/books/%s", uuid)
		results = append(results, &r)
	}

	if err := rows.Err(); err != nil {

		return nil, err
	}

	return results, nil
}

func (model *BookModel) Get(ctx context.Context, id int) (*Book, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	var book Book
	query := `
		SELECT
			id, uuid, name, description, COALESCE(image, ''), created_at, updated_at
		FROM books WHERE id = ?
	`

	row := model.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&book.ID,
		&book.UUID,
		&book.Name,
		&book.Description,
		&book.Image,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {

		return nil, err
	}

	return &book, nil
}

func (model *BookModel) GetByUUID(ctx context.Context, uuid string) (*Book, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	var book Book
	query := `
		SELECT
			id, uuid, name, description, COALESCE(image, ''), created_at, updated_at
		FROM books WHERE uuid = ?
	`

	row := model.db.QueryRowContext(ctx, query, uuid)
	err := row.Scan(
		&book.ID,
		&book.UUID,
		&book.Name,
		&book.Description,
		&book.Image,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {

		return nil, err
	}

	return &book, nil
}

func (model *BookModel) GetAllPaginated(ctx context.Context, pageSize, currentPage int, sortCol, sortDir string) ([]*Book, int, int, int, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	if pageSize < 1 {

		pageSize = 1
	}
	if currentPage < 1 {

		currentPage = 1
	}

	var totalRecords int
	countQuery := "select count(*) from books"
	err := model.db.QueryRowContext(ctx, countQuery).Scan(&totalRecords)
	if err != nil {

		return nil, 0, 0, 0, err
	}

	lastPage := (totalRecords + pageSize - 1) / pageSize
	if lastPage > 0 && currentPage > lastPage {

		currentPage = lastPage
	}

	offset := (currentPage - 1) * pageSize

	dir := ""
	if sortDir == "desc" {

		dir = " desc"
	}

	var orderClause string
	switch sortCol {
	case "description":
		orderClause = "description" + dir
	default:
		orderClause = "name" + dir
	}

	query := `
		select
			id, uuid, name, description, COALESCE(image, ''), created_at, updated_at
		from
			books
		order by
			` + orderClause + `
		limit ? offset ?
	`

	rows, err := model.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {

		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	var books []*Book
	for rows.Next() {

		var b Book
		err := rows.Scan(
			&b.ID, &b.UUID, &b.Name, &b.Description,
			&b.Image, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {

			return nil, 0, 0, 0, err
		}
		books = append(books, &b)
	}

	if err = rows.Err(); err != nil {

		return nil, 0, 0, 0, err
	}

	return books, currentPage, lastPage, totalRecords, nil
}

func (model *BookModel) Insert(ctx context.Context, book *Book) (int, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := `insert into books (name, description, image, created_at, updated_at)
		values (?, ?, ?, ?, ?) returning id`

	var id int
	err := model.db.QueryRowContext(ctx, query,
		book.Name, book.Description, book.Image, time.Now(), time.Now()).Scan(&id)
	if err != nil {

		return 0, err
	}

	return id, nil
}

func (model *BookModel) Update(ctx context.Context, book *Book) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := `update books set name = ?, description = ?, updated_at = ? where id = ?`

	_, err := model.db.ExecContext(ctx, query, book.Name, book.Description, time.Now(), book.ID)
	if err != nil {

		return err
	}

	return nil
}

func (model *BookModel) Delete(ctx context.Context, id int) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := "delete from books where id = ?"

	_, err := model.db.ExecContext(ctx, query, id)
	if err != nil {

		return err
	}

	return nil
}
