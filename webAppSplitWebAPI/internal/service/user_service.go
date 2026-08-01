package service

import (
	"context"
	"errors"
	"webProj/internal/models"
	"webProj/internal/tokens"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	MaxPasswordBytes = 72
	MinPasswordBytes = 8
	BcryptCost       = 12
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotActive   = errors.New("account not active")
	ErrEmailAlreadyTaken  = errors.New("email already registered")
	ErrCannotDeleteSelf   = errors.New("cannot delete own account")
)

type UserService struct {
	db *models.Models
}

func (s *UserService) Create(ctx context.Context, user *models.User, password string) (int, error) {

	_, err := s.db.User.GetUserByEmail(ctx, user.Email)
	if err == nil {

		return 0, ErrEmailAlreadyTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {

		return 0, err
	}

	user.Password = string(hash)

	return s.db.User.Insert(ctx, user)
}

// Authenticate verifies credentials and active status.
// Returns the user on ErrAccountNotActive so the caller can access user.Email.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {

	user, err := s.db.User.GetUserByEmail(ctx, email)
	if err != nil {

		return nil, ErrInvalidCredentials
	}

	if !s.passwordMatches(user.Password, password) {

		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {

		return user, ErrAccountNotActive
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*models.User, *tokens.Token, error) {

	user, err := s.Authenticate(ctx, email, password)
	if err != nil {

		return user, nil, err
	}

	token, err := tokens.GenerateAuthToken(user.ID, 24*time.Hour, tokens.ScopeAuthentication)
	if err != nil {

		return nil, nil, err
	}

	err = s.db.Token.Insert(ctx, token, user)
	if err != nil {

		return nil, nil, err
	}

	return user, token, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, user *models.User, newPassword string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {

		return err
	}

	return s.db.User.UpdatePassword(ctx, user, string(hash))
}

func (s *UserService) Delete(ctx context.Context, currentUserUUID, targetUUID string) error {

	if currentUserUUID == targetUUID {

		return ErrCannotDeleteSelf
	}

	user, err := s.db.User.GetByUUID(ctx, targetUUID)
	if err != nil {

		return err
	}

	return s.db.User.Delete(ctx, user.ID)
}

func (s *UserService) passwordMatches(hash, password string) bool {

	if len(password) > MaxPasswordBytes {

		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
