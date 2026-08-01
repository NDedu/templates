package main

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"webProj/internal/models"
	"webProj/internal/testutil"
	"webProj/internal/validator"
)

func Test_application_GetUser(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	created := createTestUser(t, app, "getuser@test.com", "password123", models.RoleUser, true)

	handler := http.HandlerFunc(app.GetUser)

	t.Run("empty uuid", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/admin/get-user/", nil)
		req = withChiURLParams(req, map[string]string{"uuid": ""})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("not found among 101 users", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/admin/get-user/nonexistent", nil)
		req = withChiURLParams(req, map[string]string{"uuid": "nonexistent"})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d; got %d", http.StatusNotFound, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		msg, _ := resp["message"].(string)
		if msg != validator.UserNotFoundErr {
			t.Errorf("want message %q; got %q", validator.UserNotFoundErr, msg)
		}
	})

	t.Run("valid user among 101", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/admin/get-user/"+created.UUID, nil)
		req = withChiURLParams(req, map[string]string{"uuid": created.UUID})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["error"] != false {
			t.Error("want error=false")
		}

		userData, ok := resp["user"].(map[string]any)
		if !ok {
			t.Fatal("want user object in response")
		}
		if userData["email"] != "getuser@test.com" {
			t.Errorf("want email getuser@test.com; got %v", userData["email"])
		}
		if userData["uuid"] != created.UUID {
			t.Errorf("want uuid %s; got %v", created.UUID, userData["uuid"])
		}
	})
}

func Test_application_EditUser(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	created := createTestUser(t, app, "edit@test.com", "password123", models.RoleUser, true)

	handler := http.HandlerFunc(app.EditUser)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/admin/edit-user", "{bad")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	var tests = []struct {
		name           string
		body           map[string]any
		expectedStatus int
		expectedMsg    string
	}{
		{
			"missing first name",
			map[string]any{
				"uuid": created.UUID, "last_name": "User",
				"email": "e@test.com", "role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing last name",
			map[string]any{
				"uuid": created.UUID, "first_name": "Test",
				"email": "e@test.com", "role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing email",
			map[string]any{
				"uuid": created.UUID, "first_name": "Test", "last_name": "User",
				"role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"invalid role",
			map[string]any{
				"uuid": created.UUID, "first_name": "Test", "last_name": "User",
				"email": "e@test.com", "role": "superadmin", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"user not found among 101",
			map[string]any{
				"uuid": "nonexistent", "first_name": "Test", "last_name": "User",
				"email": "e@test.com", "role": "user", "is_active": true,
			},
			http.StatusNotFound, validator.UserNotFoundErr,
		},
		{
			"successful update among 101",
			map[string]any{
				"uuid": created.UUID, "first_name": "Updated", "last_name": "Name",
				"email": "updated@test.com", "role": "editor", "is_active": true,
			},
			http.StatusOK, "User updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testPostJSON(t, handler, "/api/admin/edit-user", tt.body)

			if rr.Code != tt.expectedStatus {
				t.Errorf("want status %d; got %d", tt.expectedStatus, rr.Code)
			}

			var resp map[string]any
			readResponseJSON(t, rr, &resp)
			msg, _ := resp["message"].(string)
			if msg != tt.expectedMsg {
				t.Errorf("want message %q; got %q", tt.expectedMsg, msg)
			}
		})
	}
}

func Test_application_DeleteUser(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	admin := createTestUser(t, app, "admin@test.com", "password123", models.RoleAdmin, true)

	handler := http.HandlerFunc(app.DeleteUser)

	t.Run("empty uuid", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/admin/delete-user/", nil)
		req = withChiURLParams(req, map[string]string{"uuid": ""})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("cannot delete self among 101", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/admin/delete-user/"+admin.UUID, nil)
		req = withChiURLParams(req, map[string]string{"uuid": admin.UUID})
		req = app.contextSetUser(req, admin)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		msg, _ := resp["message"].(string)
		if msg != validator.CannotDeleteSelfErr {
			t.Errorf("want message %q; got %q", validator.CannotDeleteSelfErr, msg)
		}
	})

	t.Run("successful delete among 101", func(t *testing.T) {
		target := createTestUser(t, app, "delete-me@test.com", "password123", models.RoleUser, true)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/admin/delete-user/"+target.UUID, nil)
		req = withChiURLParams(req, map[string]string{"uuid": target.UUID})
		req = app.contextSetUser(req, admin)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["error"] != false {
			t.Error("want error=false")
		}

		// Verify user is actually deleted
		_, err := app.db.User.GetByUUID(t.Context(), target.UUID)
		if err == nil {
			t.Error("want user to be deleted from database")
		}
	})
}

func Test_application_AddUser(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	createTestUser(t, app, "existing@test.com", "password123", models.RoleUser, true)

	handler := http.HandlerFunc(app.AddUser)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/admin/add-user", "{bad")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	var tests = []struct {
		name           string
		body           map[string]any
		expectedStatus int
		expectedMsg    string
	}{
		{
			"missing first name",
			map[string]any{
				"last_name": "User", "email": "new@test.com",
				"password": "password123", "role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing email",
			map[string]any{
				"first_name": "New", "last_name": "User",
				"password": "password123", "role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing password",
			map[string]any{
				"first_name": "New", "last_name": "User", "email": "new@test.com",
				"role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"password too short",
			map[string]any{
				"first_name": "New", "last_name": "User", "email": "new@test.com",
				"password": "short", "role": "user", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"invalid role",
			map[string]any{
				"first_name": "New", "last_name": "User", "email": "new@test.com",
				"password": "password123", "role": "superadmin", "is_active": true,
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"email already registered among 101",
			map[string]any{
				"first_name": "New", "last_name": "User", "email": "existing@test.com",
				"password": "password123", "role": "user", "is_active": true,
			},
			http.StatusBadRequest, validator.EmailAlreadyRegisteredErr,
		},
		{
			"successful creation among 101",
			map[string]any{
				"first_name": "New", "last_name": "User", "email": "brand-new@test.com",
				"password": "password123", "role": "user", "is_active": true,
			},
			http.StatusCreated, "User created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testPostJSON(t, handler, "/api/admin/add-user", tt.body)

			if rr.Code != tt.expectedStatus {
				t.Errorf("want status %d; got %d", tt.expectedStatus, rr.Code)
			}

			var resp map[string]any
			readResponseJSON(t, rr, &resp)
			msg, _ := resp["message"].(string)
			if msg != tt.expectedMsg {
				t.Errorf("want message %q; got %q", tt.expectedMsg, msg)
			}
		})
	}
}

func Test_application_ResetUserPassword(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	created := createTestUser(t, app, "reset@test.com", "password123", models.RoleUser, true)

	handler := http.HandlerFunc(app.ResetUserPassword)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/admin/reset-user-password", "bad")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	var tests = []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedMsg    string
	}{
		{
			"missing password",
			map[string]string{"uuid": created.UUID, "confirm_password": "newpass123"},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"password too short",
			map[string]string{"uuid": created.UUID, "password": "short", "confirm_password": "short"},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"passwords do not match",
			map[string]string{"uuid": created.UUID, "password": "newpassword123", "confirm_password": "different123"},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"user not found among 101",
			map[string]string{"uuid": "nonexistent", "password": "newpassword123", "confirm_password": "newpassword123"},
			http.StatusNotFound, validator.UserNotFoundErr,
		},
		{
			"successful reset among 101",
			map[string]string{"uuid": created.UUID, "password": "newpassword123", "confirm_password": "newpassword123"},
			http.StatusOK, "Password updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testPostJSON(t, handler, "/api/admin/reset-user-password", tt.body)

			if rr.Code != tt.expectedStatus {
				t.Errorf("want status %d; got %d", tt.expectedStatus, rr.Code)
			}

			var resp map[string]any
			readResponseJSON(t, rr, &resp)
			msg, _ := resp["message"].(string)
			if msg != tt.expectedMsg {
				t.Errorf("want message %q; got %q", tt.expectedMsg, msg)
			}
		})
	}
}

func Test_application_GetAllUsers(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	handler := http.HandlerFunc(app.GetAllUsers)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/admin/all-users", "bad")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns page of 10 from 100 users", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/admin/all-users", map[string]any{
			"page_size": 10, "current_page": 1,
		})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["error"] != false {
			t.Error("want error=false")
		}

		users, ok := resp["users"].([]any)
		if !ok {
			t.Fatal("want users array in response")
		}
		if len(users) != 10 {
			t.Errorf("want 10 users; got %d", len(users))
		}

		totalRecords, _ := resp["total_records"].(float64)
		if totalRecords != 100 {
			t.Errorf("want total_records=100; got %v", totalRecords)
		}

		lastPage, _ := resp["last_page"].(float64)
		if lastPage != 10 {
			t.Errorf("want last_page=10; got %v", lastPage)
		}
	})

	t.Run("page size 25 gives 4 pages", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/admin/all-users", map[string]any{
			"page_size": 25, "current_page": 1,
		})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)

		users, ok := resp["users"].([]any)
		if !ok {
			t.Fatal("want users array in response")
		}
		if len(users) != 25 {
			t.Errorf("want 25 users; got %d", len(users))
		}

		lastPage, _ := resp["last_page"].(float64)
		if lastPage != 4 {
			t.Errorf("want last_page=4; got %v", lastPage)
		}
	})

	t.Run("middle page returns correct count", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/admin/all-users", map[string]any{
			"page_size": 10, "current_page": 5,
		})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)

		users, ok := resp["users"].([]any)
		if !ok {
			t.Fatal("want users array in response")
		}
		if len(users) != 10 {
			t.Errorf("want 10 users on page 5; got %d", len(users))
		}

		currentPage, _ := resp["current_page"].(float64)
		if currentPage != 5 {
			t.Errorf("want current_page=5; got %v", currentPage)
		}
	})

	t.Run("last page has correct remainder", func(t *testing.T) {
		// 100 / 7 = 14 full pages + 2 remainder → lastPage = 15
		rr := testPostJSON(t, handler, "/api/admin/all-users", map[string]any{
			"page_size": 7, "current_page": 15,
		})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)

		users, ok := resp["users"].([]any)
		if !ok {
			t.Fatal("want users array in response")
		}
		if len(users) != 2 {
			t.Errorf("want 2 users on last page; got %d", len(users))
		}
	})
}

func Test_application_UploadProfileImage(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create static/img/user directory in temp dir
	uploadDir := filepath.Join(app.staticPath, "img", "user")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {

		t.Fatal(err)
	}

	created := createTestUser(t, app, "upload@test.com", "password123", models.RoleAdmin, true)

	handler := http.HandlerFunc(app.UploadProfileImage)

	t.Run("empty uuid", func(t *testing.T) {
		pngData := testutil.CreateTestPNG(t)
		req := createMultipartRequest(t, "/api/admin/upload-profile-image/", "profile_image", "test.png", pngData, "image/png")
		req = withChiURLParams(req, map[string]string{"uuid": ""})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		pngData := testutil.CreateTestPNG(t)
		req := createMultipartRequest(t, "/api/admin/upload-profile-image/nonexistent", "profile_image", "test.png", pngData, "image/png")
		req = withChiURLParams(req, map[string]string{"uuid": "nonexistent"})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d; got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("invalid content type", func(t *testing.T) {
		req := createMultipartRequest(t, "/api/admin/upload-profile-image/"+created.UUID, "profile_image", "test.txt", []byte("not an image"), "text/plain")
		req = withChiURLParams(req, map[string]string{"uuid": created.UUID})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		msg, _ := resp["message"].(string)
		if msg != validator.InvalidImageTypeErr {
			t.Errorf("want message %q; got %q", validator.InvalidImageTypeErr, msg)
		}
	})

	t.Run("no file provided", func(t *testing.T) {
		// Send a request with no file field
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/admin/upload-profile-image/"+created.UUID, &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = withChiURLParams(req, map[string]string{"uuid": created.UUID})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("successful png upload", func(t *testing.T) {
		pngData := testutil.CreateTestPNG(t)
		req := createMultipartRequest(t, "/api/admin/upload-profile-image/"+created.UUID, "profile_image", "test.png", pngData, "image/png")
		req = withChiURLParams(req, map[string]string{"uuid": created.UUID})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["error"] != false {
			t.Error("want error=false")
		}

		profileImage, _ := resp["profile_image"].(string)
		if !strings.HasPrefix(profileImage, "/static/img/user/") {
			t.Errorf("want profile_image to start with /static/img/user/; got %q", profileImage)
		}
		if !strings.HasSuffix(profileImage, ".png") {
			t.Errorf("want profile_image to end with .png; got %q", profileImage)
		}

		// Verify file exists on disk
		diskPath := filepath.Join(app.staticPath, strings.TrimPrefix(profileImage, "/static/"))
		if _, err := os.Stat(diskPath); os.IsNotExist(err) {
			t.Error("want uploaded file to exist on disk")
		}

		// Verify DB was updated
		user, err := app.db.User.GetByUUID(t.Context(), created.UUID)
		if err != nil {
			t.Fatal(err)
		}
		if user.ProfileImage != profileImage {
			t.Errorf("want DB profile_image %q; got %q", profileImage, user.ProfileImage)
		}
	})

	t.Run("upload replaces old image", func(t *testing.T) {
		// Get current image path
		user, err := app.db.User.GetByUUID(t.Context(), created.UUID)
		if err != nil {
			t.Fatal(err)
		}
		oldImage := user.ProfileImage
		oldDiskPath := filepath.Join(app.staticPath, strings.TrimPrefix(oldImage, "/static/"))

		// Upload a new image
		pngData := testutil.CreateTestPNG(t)
		req := createMultipartRequest(t, "/api/admin/upload-profile-image/"+created.UUID, "profile_image", "new.png", pngData, "image/png")
		req = withChiURLParams(req, map[string]string{"uuid": created.UUID})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		newImage, _ := resp["profile_image"].(string)

		if newImage == oldImage {
			t.Error("want new image path to differ from old")
		}

		// Verify old file was deleted
		if _, err := os.Stat(oldDiskPath); !os.IsNotExist(err) {
			t.Error("want old image file to be deleted from disk")
		}

		// Verify new file exists
		newDiskPath := filepath.Join(app.staticPath, strings.TrimPrefix(newImage, "/static/"))
		if _, err := os.Stat(newDiskPath); os.IsNotExist(err) {
			t.Error("want new uploaded file to exist on disk")
		}
	})

	t.Run("successful jpeg upload", func(t *testing.T) {
		// Use raw JPEG header bytes (minimal valid JPEG)
		jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}

		req := createMultipartRequest(t, "/api/admin/upload-profile-image/"+created.UUID, "profile_image", "test.jpg", jpegData, "image/jpeg")
		req = withChiURLParams(req, map[string]string{"uuid": created.UUID})

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		profileImage, _ := resp["profile_image"].(string)
		if !strings.HasSuffix(profileImage, ".jpg") {
			t.Errorf("want profile_image to end with .jpg; got %q", profileImage)
		}
	})
}
