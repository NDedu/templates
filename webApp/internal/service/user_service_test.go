package service_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/testutil"

	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Create(t *testing.T) {

	t.Parallel()

	t.Run("creates user with hashed password among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		id, err := svc.User.Create(t.Context(), &models.User{
			FirstName: "John", LastName: "Doe", Email: "john@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "password123")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if id < 1 {
			t.Errorf("want id >= 1; got %d", id)
		}

		user, err := m.User.GetUserByEmail(t.Context(), "john@test.com")
		if err != nil {
			t.Fatalf("get by email: %v", err)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123")); err != nil {
			t.Error("want password to be valid bcrypt hash of 'password123'")
		}
	})

	t.Run("returns ErrEmailAlreadyTaken with 100 existing", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		users := testutil.CreateUsers(t, m, 100)

		// Try to create a user with an email that already exists among the 100
		_, err := svc.User.Create(t.Context(), &models.User{
			FirstName: "Dup", LastName: "User", Email: users[50].Email,
			Role: models.RoleUser, IsActive: true,
		}, "password123")
		if !errors.Is(err, service.ErrEmailAlreadyTaken) {
			t.Errorf("want ErrEmailAlreadyTaken; got %v", err)
		}
	})

	t.Run("assigns correct role and active status among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		id, err := svc.User.Create(t.Context(), &models.User{
			FirstName: "Ed", LastName: "Itor", Email: "editor@test.com",
			Role: models.RoleEditor, IsActive: false,
		}, "password123")
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		user, err := m.User.Get(t.Context(), id)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if user.Role != models.RoleEditor {
			t.Errorf("want role editor; got %s", user.Role)
		}
		if user.IsActive {
			t.Error("want IsActive false")
		}
	})
}

func TestUserService_Authenticate(t *testing.T) {

	t.Parallel()

	t.Run("returns user for valid credentials among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Auth", LastName: "User", Email: "auth@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "password123")

		user, err := svc.User.Authenticate(t.Context(), "auth@test.com", "password123")
		if err != nil {
			t.Fatalf("authenticate: %v", err)
		}
		if user.Email != "auth@test.com" {
			t.Errorf("want email auth@test.com; got %s", user.Email)
		}
	})

	t.Run("returns ErrInvalidCredentials for wrong email among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		_, err := svc.User.Authenticate(t.Context(), "nobody@test.com", "password123")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("want ErrInvalidCredentials; got %v", err)
		}
	})

	t.Run("returns ErrInvalidCredentials for wrong password among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Wrong", LastName: "Pass", Email: "wrongpass@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "password123")

		_, err := svc.User.Authenticate(t.Context(), "wrongpass@test.com", "wrongpassword")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("want ErrInvalidCredentials; got %v", err)
		}
	})

	t.Run("returns ErrAccountNotActive for inactive user among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Inactive", LastName: "User", Email: "inactive@test.com",
			Role: models.RoleUser, IsActive: false,
		}, "password123")

		user, err := svc.User.Authenticate(t.Context(), "inactive@test.com", "password123")
		if !errors.Is(err, service.ErrAccountNotActive) {
			t.Errorf("want ErrAccountNotActive; got %v", err)
		}
		if user == nil {
			t.Fatal("want user to be returned on ErrAccountNotActive")
		}
		if user.Email != "inactive@test.com" {
			t.Errorf("want email inactive@test.com; got %s", user.Email)
		}
	})
}

func TestUserService_Login(t *testing.T) {

	t.Parallel()

	t.Run("returns user and token on success among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Login", LastName: "User", Email: "login@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "password123")

		user, token, err := svc.User.Login(t.Context(), "login@test.com", "password123")
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if user == nil {
			t.Fatal("want user to be returned")
		}
		if token == nil {
			t.Fatal("want token to be returned")
		}
		if token.PlainText == "" {
			t.Error("want token PlainText to be set")
		}
	})

	t.Run("token is persisted in database with 100 users", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Token", LastName: "User", Email: "token@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "password123")

		_, token, err := svc.User.Login(t.Context(), "token@test.com", "password123")
		if err != nil {
			t.Fatalf("login: %v", err)
		}

		tokenUser, err := m.Token.GetUserByToken(t.Context(), token.PlainText)
		if err != nil {
			t.Fatalf("get user by token: %v", err)
		}
		if tokenUser.Email != "token@test.com" {
			t.Errorf("want email token@test.com; got %s", tokenUser.Email)
		}
	})

	t.Run("returns ErrInvalidCredentials for bad password among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Bad", LastName: "Pass", Email: "badpass@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "password123")

		user, token, err := svc.User.Login(t.Context(), "badpass@test.com", "wrongpassword")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("want ErrInvalidCredentials; got %v", err)
		}
		if user != nil {
			t.Error("want user nil on invalid credentials")
		}
		if token != nil {
			t.Error("want token nil on invalid credentials")
		}
	})

	t.Run("returns ErrAccountNotActive for inactive user among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Inactive", LastName: "Login", Email: "inactive-login@test.com",
			Role: models.RoleUser, IsActive: false,
		}, "password123")

		user, token, err := svc.User.Login(t.Context(), "inactive-login@test.com", "password123")
		if !errors.Is(err, service.ErrAccountNotActive) {
			t.Errorf("want ErrAccountNotActive; got %v", err)
		}
		if user == nil {
			t.Fatal("want user returned on inactive account")
		}
		if token != nil {
			t.Error("want token nil on inactive account")
		}
	})
}

func TestUserService_UpdatePassword(t *testing.T) {

	t.Parallel()

	t.Run("new password works after update among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "PW", LastName: "Update", Email: "pwupdate@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "oldpassword")

		user, _ := svc.User.Authenticate(t.Context(), "pwupdate@test.com", "oldpassword")

		err := svc.User.UpdatePassword(t.Context(), user, "newpassword123")
		if err != nil {
			t.Fatalf("update password: %v", err)
		}

		_, err = svc.User.Authenticate(t.Context(), "pwupdate@test.com", "newpassword123")
		if err != nil {
			t.Errorf("want authentication with new password to succeed; got %v", err)
		}
	})

	t.Run("old password no longer works among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "PW", LastName: "Old", Email: "pwold@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "oldpassword")

		user, _ := svc.User.Authenticate(t.Context(), "pwold@test.com", "oldpassword")

		svc.User.UpdatePassword(t.Context(), user, "newpassword123")

		_, err := svc.User.Authenticate(t.Context(), "pwold@test.com", "oldpassword")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("want ErrInvalidCredentials with old password; got %v", err)
		}
	})

	t.Run("other users passwords unaffected", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Target", LastName: "User", Email: "target@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "targetpass")

		svc.User.Create(t.Context(), &models.User{
			FirstName: "Other", LastName: "User", Email: "other@test.com",
			Role: models.RoleUser, IsActive: true,
		}, "otherpass")

		target, _ := svc.User.Authenticate(t.Context(), "target@test.com", "targetpass")

		svc.User.UpdatePassword(t.Context(), target, "newpassword123")

		// Other user's password should still work
		_, err := svc.User.Authenticate(t.Context(), "other@test.com", "otherpass")
		if err != nil {
			t.Errorf("want other user's auth to still work; got %v", err)
		}
	})
}

func TestUserService_Delete(t *testing.T) {

	t.Parallel()

	t.Run("deletes one user among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		users := testutil.CreateUsers(t, m, 100)
		target := users[50]
		admin := users[0]

		err := svc.User.Delete(t.Context(), admin.UUID, target.UUID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = m.User.GetByUUID(t.Context(), target.UUID)
		if err == nil {
			t.Error("want user to be deleted")
		}

		remaining, err := m.User.GetAll(t.Context())
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if len(remaining) != 99 {
			t.Errorf("want 99 remaining users; got %d", len(remaining))
		}
	})

	t.Run("returns ErrCannotDeleteSelf among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		users := testutil.CreateUsers(t, m, 100)

		err := svc.User.Delete(t.Context(), users[0].UUID, users[0].UUID)
		if !errors.Is(err, service.ErrCannotDeleteSelf) {
			t.Errorf("want ErrCannotDeleteSelf; got %v", err)
		}
	})

	t.Run("returns error for nonexistent target uuid among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		users := testutil.CreateUsers(t, m, 100)

		err := svc.User.Delete(t.Context(), users[0].UUID, "nonexistent")
		if err == nil {
			t.Error("want error for nonexistent target")
		}
	})
}

func TestUserService_Update(t *testing.T) {

	t.Parallel()

	t.Run("updates user fields among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		user := testutil.CreateUser(t, m, "update@test.com", "password123", models.RoleUser, true)

		err := svc.User.Update(t.Context(), user.UUID, "NewFirst", "NewLast", "newemail@test.com", models.RoleEditor, false)
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		updated, err := m.User.GetByUUID(t.Context(), user.UUID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if updated.FirstName != "NewFirst" {
			t.Errorf("want first_name 'NewFirst'; got %q", updated.FirstName)
		}
		if updated.LastName != "NewLast" {
			t.Errorf("want last_name 'NewLast'; got %q", updated.LastName)
		}
		if updated.Email != "newemail@test.com" {
			t.Errorf("want email 'newemail@test.com'; got %q", updated.Email)
		}
		if updated.Role != models.RoleEditor {
			t.Errorf("want role editor; got %s", updated.Role)
		}
		if updated.IsActive {
			t.Error("want IsActive false")
		}
	})

	t.Run("returns ErrUserNotFound among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		err := svc.User.Update(t.Context(), "nonexistent", "F", "L", "e@test.com", models.RoleUser, true)
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("want ErrUserNotFound; got %v", err)
		}
	})

	t.Run("other users unaffected", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		target := testutil.CreateUser(t, m, "target@test.com", "password123", models.RoleUser, true)
		other := testutil.CreateUser(t, m, "other@test.com", "password123", models.RoleUser, true)

		svc.User.Update(t.Context(), target.UUID, "Changed", "Changed", "changed@test.com", models.RoleAdmin, false)

		otherUser, _ := m.User.GetByUUID(t.Context(), other.UUID)
		if otherUser.FirstName != "Test" {
			t.Errorf("want other user unaffected; got first_name %q", otherUser.FirstName)
		}
	})

	t.Run("rejects an email already used by another user", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)

		testutil.CreateUser(t, m, "taken@test.com", "password123", models.RoleUser, true)
		user := testutil.CreateUser(t, m, "editme@test.com", "password123", models.RoleUser, true)

		err := svc.User.Update(t.Context(), user.UUID, "F", "L", "taken@test.com", models.RoleUser, true)
		if !errors.Is(err, service.ErrEmailAlreadyTaken) {
			t.Errorf("want ErrEmailAlreadyTaken; got %v", err)
		}
	})

	t.Run("allows keeping the same email", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)

		user := testutil.CreateUser(t, m, "keep@test.com", "password123", models.RoleUser, true)

		err := svc.User.Update(t.Context(), user.UUID, "F", "L", "keep@test.com", models.RoleEditor, true)
		if err != nil {
			t.Fatalf("want no error keeping same email; got %v", err)
		}
	})
}

func TestUserService_UploadProfileImage(t *testing.T) {

	t.Parallel()

	t.Run("valid PNG upload among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateUsers(t, m, 100)

		user := testutil.CreateUser(t, m, "upload@test.com", "password123", models.RoleUser, true)

		staticPath := t.TempDir()
		pngData := testutil.CreateTestPNG(t)

		imagePath, err := svc.User.UploadProfileImage(t.Context(), user.UUID, bytes.NewReader(pngData), staticPath)
		if err != nil {
			t.Fatalf("upload: %v", err)
		}

		if !strings.HasPrefix(imagePath, "/static/img/user/") {
			t.Errorf("want prefix /static/img/user/; got %q", imagePath)
		}
		if !strings.HasSuffix(imagePath, ".png") {
			t.Errorf("want suffix .png; got %q", imagePath)
		}

		// Verify file on disk
		diskPath := filepath.Join(staticPath, strings.TrimPrefix(imagePath, "/static/"))
		if _, err := os.Stat(diskPath); os.IsNotExist(err) {
			t.Error("want uploaded file to exist on disk")
		}

		// Verify DB
		updated, _ := m.User.GetByUUID(t.Context(), user.UUID)
		if updated.ProfileImage != imagePath {
			t.Errorf("want DB profile_image %q; got %q", imagePath, updated.ProfileImage)
		}
	})

	t.Run("valid JPEG upload", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		user := testutil.CreateUser(t, m, "jpeg@test.com", "password123", models.RoleUser, true)

		staticPath := t.TempDir()
		jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}

		imagePath, err := svc.User.UploadProfileImage(t.Context(), user.UUID, bytes.NewReader(jpegData), staticPath)
		if err != nil {
			t.Fatalf("upload: %v", err)
		}

		if !strings.HasSuffix(imagePath, ".jpg") {
			t.Errorf("want suffix .jpg; got %q", imagePath)
		}
	})

	t.Run("invalid image type", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		user := testutil.CreateUser(t, m, "bad@test.com", "password123", models.RoleUser, true)

		staticPath := t.TempDir()

		_, err := svc.User.UploadProfileImage(t.Context(), user.UUID, bytes.NewReader([]byte("not an image")), staticPath)
		if !errors.Is(err, service.ErrInvalidImageType) {
			t.Errorf("want ErrInvalidImageType; got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {

		t.Parallel()

		svc, _ := testutil.NewTestServices(t)

		staticPath := t.TempDir()
		pngData := testutil.CreateTestPNG(t)

		_, err := svc.User.UploadProfileImage(t.Context(), "nonexistent", bytes.NewReader(pngData), staticPath)
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("want ErrUserNotFound; got %v", err)
		}
	})

	t.Run("replaces old image", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		user := testutil.CreateUser(t, m, "replace@test.com", "password123", models.RoleUser, true)

		staticPath := t.TempDir()
		pngData := testutil.CreateTestPNG(t)

		// First upload
		oldPath, err := svc.User.UploadProfileImage(t.Context(), user.UUID, bytes.NewReader(pngData), staticPath)
		if err != nil {
			t.Fatalf("first upload: %v", err)
		}

		// Second upload
		newPath, err := svc.User.UploadProfileImage(t.Context(), user.UUID, bytes.NewReader(pngData), staticPath)
		if err != nil {
			t.Fatalf("second upload: %v", err)
		}

		if newPath == oldPath {
			t.Error("want different path for new upload")
		}

		// Old file should be deleted
		oldDiskPath := filepath.Join(staticPath, strings.TrimPrefix(oldPath, "/static/"))
		if _, err := os.Stat(oldDiskPath); !os.IsNotExist(err) {
			t.Error("want old image to be deleted")
		}

		// New file should exist
		newDiskPath := filepath.Join(staticPath, strings.TrimPrefix(newPath, "/static/"))
		if _, err := os.Stat(newDiskPath); os.IsNotExist(err) {
			t.Error("want new image to exist on disk")
		}
	})
}
