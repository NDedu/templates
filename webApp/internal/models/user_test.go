package models_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"webProj/internal/models"
	"webProj/internal/testutil"
)

func TestUserModel_Insert(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)

	t.Run("bulk insert 100 users with unique IDs", func(t *testing.T) {

		ids := make(map[int]bool, 100)
		for i := range 100 {

			id, err := m.User.Insert(t.Context(), &models.User{
				FirstName: "Bulk", LastName: fmt.Sprintf("User%03d", i),
				Email: fmt.Sprintf("bulk%d@test.com", i),
				Password: "hash", Role: models.RoleUser, IsActive: true,
			})
			if err != nil {
				t.Fatalf("insert %d: %v", i, err)
			}
			if ids[id] {
				t.Fatalf("duplicate id %d on insert %d", id, i)
			}
			ids[id] = true
		}

		if len(ids) != 100 {
			t.Errorf("want 100 unique IDs; got %d", len(ids))
		}
	})

	t.Run("email is lowercased and trimmed", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		testutil.CreateUsers(t, m, 100)

		_, err := m.User.Insert(t.Context(), &models.User{
			FirstName: "Jane", LastName: "Doe", Email: " FOO@Bar.COM ",
			Password: "hash", Role: models.RoleUser,
		})
		if err != nil {
			t.Fatalf("insert: %v", err)
		}

		user, err := m.User.GetUserByEmail(t.Context(), "foo@bar.com")
		if err != nil {
			t.Fatalf("get by email: %v", err)
		}
		if user.Email != "foo@bar.com" {
			t.Errorf("want email foo@bar.com; got %s", user.Email)
		}
	})
}

func TestUserModel_Get(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	users := testutil.CreateUsers(t, m, 100)

	t.Run("returns correct user among 100", func(t *testing.T) {

		target := users[49]
		user, err := m.User.Get(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if user.Email != target.Email {
			t.Errorf("want email %s; got %s", target.Email, user.Email)
		}
		if user.UUID != target.UUID {
			t.Errorf("want UUID %s; got %s", target.UUID, user.UUID)
		}
	})

	t.Run("returns first and last among 100", func(t *testing.T) {

		for _, target := range []*models.User{users[0], users[99]} {

			user, err := m.User.Get(t.Context(), target.ID)
			if err != nil {
				t.Fatalf("get id %d: %v", target.ID, err)
			}
			if user.Email != target.Email {
				t.Errorf("want email %s; got %s", target.Email, user.Email)
			}
		}
	})

	t.Run("returns error for nonexistent id", func(t *testing.T) {

		_, err := m.User.Get(t.Context(), 99999)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("want sql.ErrNoRows; got %v", err)
		}
	})
}

func TestUserModel_GetByUUID(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	users := testutil.CreateUsers(t, m, 100)

	t.Run("returns correct user by uuid among 100", func(t *testing.T) {

		target := users[73]
		user, err := m.User.GetByUUID(t.Context(), target.UUID)
		if err != nil {
			t.Fatalf("get by uuid: %v", err)
		}
		if user.Email != target.Email {
			t.Errorf("want email %s; got %s", target.Email, user.Email)
		}
	})

	t.Run("returns error for unknown uuid", func(t *testing.T) {

		_, err := m.User.GetByUUID(t.Context(), "nonexistent")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("want sql.ErrNoRows; got %v", err)
		}
	})
}

func TestUserModel_GetUserByEmail(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	users := testutil.CreateUsers(t, m, 100)

	t.Run("returns user with password populated among 100", func(t *testing.T) {

		user, err := m.User.GetUserByEmail(t.Context(), users[42].Email)
		if err != nil {
			t.Fatalf("get by email: %v", err)
		}
		if user.Password == "" {
			t.Error("want password to be populated")
		}
	})

	t.Run("email lookup is case-insensitive among 100", func(t *testing.T) {

		target := users[88]
		user, err := m.User.GetUserByEmail(t.Context(), fmt.Sprintf("USER%d@TEST.COM", 89))
		if err != nil {
			t.Fatalf("get by email: %v", err)
		}
		if user.Email != target.Email {
			t.Errorf("want %s; got %s", target.Email, user.Email)
		}
	})

	t.Run("returns error for missing email", func(t *testing.T) {

		_, err := m.User.GetUserByEmail(t.Context(), "nobody@test.com")
		if err == nil {
			t.Fatal("want error; got nil")
		}
		if err.Error() != "no matching record found" {
			t.Errorf("want 'no matching record found'; got %q", err.Error())
		}
	})
}

func TestUserModel_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("returns all 100 users in order", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		testutil.CreateUsers(t, m, 100)

		users, err := m.User.GetAll(t.Context())
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if len(users) != 100 {
			t.Fatalf("want 100 users; got %d", len(users))
		}
		for i := 1; i < len(users); i++ {

			if users[i-1].LastName > users[i].LastName {
				t.Errorf("want ascending last_name order; got %q before %q at index %d", users[i-1].LastName, users[i].LastName, i)
				break
			}
		}
	})

	t.Run("returns empty slice when no users", func(t *testing.T) {

		m := testutil.NewTestDB(t)

		users, err := m.User.GetAll(t.Context())
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if len(users) != 0 {
			t.Errorf("want 0 users; got %d", len(users))
		}
	})
}

func TestUserModel_GetAllPaginated(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	testutil.CreateUsers(t, m, 100)

	t.Run("page size 10 gives 10 pages", func(t *testing.T) {

		users, currentPage, lastPage, totalRecords, err := m.User.GetAllPaginated(t.Context(), 10, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(users) != 10 {
			t.Errorf("want 10 users; got %d", len(users))
		}
		if currentPage != 1 {
			t.Errorf("want currentPage 1; got %d", currentPage)
		}
		if lastPage != 10 {
			t.Errorf("want lastPage 10; got %d", lastPage)
		}
		if totalRecords != 100 {
			t.Errorf("want totalRecords 100; got %d", totalRecords)
		}
	})

	t.Run("page size 25 gives 4 pages", func(t *testing.T) {

		users, _, lastPage, _, err := m.User.GetAllPaginated(t.Context(), 25, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(users) != 25 {
			t.Errorf("want 25 users; got %d", len(users))
		}
		if lastPage != 4 {
			t.Errorf("want lastPage 4; got %d", lastPage)
		}
	})

	t.Run("middle page returns correct slice", func(t *testing.T) {

		users, currentPage, _, _, err := m.User.GetAllPaginated(t.Context(), 10, 5, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(users) != 10 {
			t.Errorf("want 10 users on page 5; got %d", len(users))
		}
		if currentPage != 5 {
			t.Errorf("want currentPage 5; got %d", currentPage)
		}
	})

	t.Run("last page has correct remainder", func(t *testing.T) {

		// 100 / 7 = 14 full pages + 2 remainder → lastPage = 15
		users, currentPage, lastPage, _, err := m.User.GetAllPaginated(t.Context(), 7, 15, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if lastPage != 15 {
			t.Errorf("want lastPage 15; got %d", lastPage)
		}
		if currentPage != 15 {
			t.Errorf("want currentPage 15; got %d", currentPage)
		}
		if len(users) != 2 {
			t.Errorf("want 2 users on last page; got %d", len(users))
		}
	})

	t.Run("clamps currentPage to lastPage", func(t *testing.T) {

		_, currentPage, lastPage, _, err := m.User.GetAllPaginated(t.Context(), 10, 99, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if currentPage != lastPage {
			t.Errorf("want currentPage clamped to %d; got %d", lastPage, currentPage)
		}
	})

	t.Run("clamps pageSize minimum to 1", func(t *testing.T) {

		users, _, _, _, err := m.User.GetAllPaginated(t.Context(), 0, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(users) != 1 {
			t.Errorf("want 1 user; got %d", len(users))
		}
	})

	t.Run("sort by email desc across 100", func(t *testing.T) {

		users, _, _, _, err := m.User.GetAllPaginated(t.Context(), 100, 1, "email", "desc")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(users) != 100 {
			t.Fatalf("want 100 users; got %d", len(users))
		}
		for i := 1; i < len(users); i++ {

			if users[i-1].Email < users[i].Email {
				t.Errorf("want descending email order; got %s before %s", users[i-1].Email, users[i].Email)
				break
			}
		}
	})

	t.Run("default sort is last_name asc across 100", func(t *testing.T) {

		users, _, _, _, err := m.User.GetAllPaginated(t.Context(), 100, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		for i := 1; i < len(users); i++ {

			if users[i-1].LastName > users[i].LastName {
				t.Errorf("want ascending last_name order; got %s before %s", users[i-1].LastName, users[i].LastName)
				break
			}
		}
	})
}

func TestUserModel_Update(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	users := testutil.CreateUsers(t, m, 100)
	target := users[50]

	t.Run("updates one user among 100", func(t *testing.T) {

		target.FirstName = "Updated"
		target.LastName = "Name"
		target.Email = "updated@test.com"
		target.Role = models.RoleEditor
		target.IsActive = false

		err := m.User.Update(t.Context(), target)
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		got, err := m.User.Get(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.FirstName != "Updated" {
			t.Errorf("want FirstName Updated; got %s", got.FirstName)
		}
		if got.LastName != "Name" {
			t.Errorf("want LastName Name; got %s", got.LastName)
		}
		if got.Email != "updated@test.com" {
			t.Errorf("want email updated@test.com; got %s", got.Email)
		}
		if got.Role != models.RoleEditor {
			t.Errorf("want role editor; got %s", got.Role)
		}
		if got.IsActive {
			t.Error("want IsActive false")
		}
	})

	t.Run("other users unchanged after update", func(t *testing.T) {

		neighbor := users[51]
		got, err := m.User.Get(t.Context(), neighbor.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.FirstName != neighbor.FirstName {
			t.Errorf("want neighbor FirstName %s unchanged; got %s", neighbor.FirstName, got.FirstName)
		}
	})

	t.Run("email is lowercased on update", func(t *testing.T) {

		target.Email = "UPPER@TEST.COM"
		err := m.User.Update(t.Context(), target)
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		got, err := m.User.Get(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Email != "upper@test.com" {
			t.Errorf("want upper@test.com; got %s", got.Email)
		}
	})
}

func TestUserModel_Activate(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	testutil.CreateUsers(t, m, 100)
	inactive := testutil.CreateUser(t, m, "inactive@test.com", "pass", models.RoleUser, false)

	err := m.User.Activate(t.Context(), inactive.ID)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}

	got, err := m.User.Get(t.Context(), inactive.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.IsActive {
		t.Error("want IsActive true after activation")
	}
}

func TestUserModel_Delete(t *testing.T) {
	t.Parallel()

	t.Run("deletes one among 100 and verifies count", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		users := testutil.CreateUsers(t, m, 100)
		target := users[50]

		err := m.User.Delete(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = m.User.Get(t.Context(), target.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("want sql.ErrNoRows after delete; got %v", err)
		}

		remaining, err := m.User.GetAll(t.Context())
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if len(remaining) != 99 {
			t.Errorf("want 99 remaining users; got %d", len(remaining))
		}
	})

	t.Run("neighbors still accessible after delete", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		users := testutil.CreateUsers(t, m, 100)

		m.User.Delete(t.Context(), users[50].ID)

		for _, idx := range []int{49, 51} {

			_, err := m.User.Get(t.Context(), users[idx].ID)
			if err != nil {
				t.Errorf("want neighbor %d still accessible; got %v", idx, err)
			}
		}
	})

	t.Run("deleting nonexistent id does not error", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		testutil.CreateUsers(t, m, 100)

		err := m.User.Delete(t.Context(), 99999)
		if err != nil {
			t.Errorf("want nil error; got %v", err)
		}
	})
}

func TestUserModel_UpdatePassword(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	users := testutil.CreateUsers(t, m, 100)
	target := users[75]

	err := m.User.UpdatePassword(t.Context(), target, "newhash")
	if err != nil {
		t.Fatalf("update password: %v", err)
	}

	got, err := m.User.GetUserByEmail(t.Context(), target.Email)
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if got.Password != "newhash" {
		t.Errorf("want password newhash; got %s", got.Password)
	}

	// Verify other users' passwords unchanged
	other, err := m.User.GetUserByEmail(t.Context(), users[76].Email)
	if err != nil {
		t.Fatalf("get other: %v", err)
	}
	if other.Password == "newhash" {
		t.Error("want other user's password unchanged")
	}
}
