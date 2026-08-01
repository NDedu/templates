package service

import "webProj/internal/models"

type Services struct {
	User *UserService
}

func NewServices(db *models.Models) *Services {

	return &Services{
		User: &UserService{db: db},
	}
}
