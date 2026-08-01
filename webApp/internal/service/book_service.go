package service

import (
	"context"
	"errors"
	"webProj/internal/models"
)

var ErrBookNotFound = errors.New("book not found")

type BookService struct {
	db *models.Models
}

func (s *BookService) Create(ctx context.Context, book *models.Book) (int, error) {

	return s.db.Book.Insert(ctx, book)
}

func (s *BookService) Update(ctx context.Context, uuid, name, description string) error {

	book, err := s.db.Book.GetByUUID(ctx, uuid)
	if err != nil {

		return ErrBookNotFound
	}

	book.Name = name
	book.Description = description

	return s.db.Book.Update(ctx, book)
}

func (s *BookService) Delete(ctx context.Context, uuid string) error {

	book, err := s.db.Book.GetByUUID(ctx, uuid)
	if err != nil {

		return ErrBookNotFound
	}

	return s.db.Book.Delete(ctx, book.ID)
}
