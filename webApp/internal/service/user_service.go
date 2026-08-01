package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"webProj/internal/models"
	"webProj/internal/tokens"

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
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidImageType   = errors.New("invalid image type")
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

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

func (s *UserService) Update(ctx context.Context, uuid, firstName, lastName, email, role string, isActive bool) error {

	user, err := s.db.User.GetByUUID(ctx, uuid)
	if err != nil {

		return ErrUserNotFound
	}

	existing, err := s.db.User.GetUserByEmail(ctx, email)
	if err == nil && existing.ID != user.ID {

		return ErrEmailAlreadyTaken
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Role = role
	user.IsActive = isActive

	return s.db.User.Update(ctx, user)
}

// FIX: for picture flashing consider server side loading
func (s *UserService) UploadProfileImage(ctx context.Context, uuid string, file io.ReadSeeker, staticPath string) (string, error) {

	buff := make([]byte, 512)
	n, err := file.Read(buff)
	if err != nil && err != io.EOF {

		return "", err
	}
	buff = buff[:n]

	contentType := http.DetectContentType(buff)
	ext, ok := allowedImageTypes[contentType]
	if !ok {

		return "", ErrInvalidImageType
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {

		return "", err
	}

	user, err := s.db.User.GetByUUID(ctx, uuid)
	if err != nil {

		return "", ErrUserNotFound
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {

		return "", err
	}
	filename := hex.EncodeToString(randBytes) + ext

	uploadDir := filepath.Join(staticPath, "img", "user")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {

		return "", err
	}

	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {

		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {

		return "", err
	}

	// Delete old profile image if it exists
	if user.ProfileImage != "" {

		oldPath := filepath.Join(staticPath, strings.TrimPrefix(user.ProfileImage, "/static/"))
		os.Remove(oldPath)
	}

	imagePath := "/static/img/user/" + filename
	if err := s.db.User.UpdateProfileImage(ctx, user.ID, imagePath); err != nil {

		// Clean up the newly uploaded file on DB error
		os.Remove(filepath.Join(uploadDir, filename))
		return "", err
	}

	return imagePath, nil
}

func (s *UserService) passwordMatches(hash, password string) bool {

	if len(password) > MaxPasswordBytes {

		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
