package service

import "webProj/internal/models"

type Services struct {
	User *UserService
	Book *BookService
}

func NewServices(db *models.Models) *Services {

	return &Services{
		User: &UserService{db: db},
		Book: &BookService{db: db},
	}
}
