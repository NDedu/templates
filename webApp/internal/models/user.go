package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const (
	RoleUser   = "user"
	RoleAdmin  = "admin"
	RoleEditor = "editor"
)

type UserModel struct {
	db DBTX
}

type User struct {
	ID           int       `json:"id"`
	UUID         string    `json:"uuid"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	ProfileImage string    `json:"profile_image"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (model *UserModel) Get(ctx context.Context, id int) (*User, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	var user User
	query := "select id, uuid, first_name, last_name, email, role, is_active, profile_image, created_at, updated_at from users where id = ?"

	row := model.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID, &user.UUID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.IsActive, &user.ProfileImage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {

		return nil, err
	}

	return &user, nil
}

func (model *UserModel) GetByUUID(ctx context.Context, uuid string) (*User, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	var user User
	query := "select id, uuid, first_name, last_name, email, role, is_active, profile_image, created_at, updated_at from users where uuid = ?"

	row := model.db.QueryRowContext(ctx, query, uuid)
	err := row.Scan(&user.ID, &user.UUID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.IsActive, &user.ProfileImage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {

		return nil, err
	}

	return &user, nil
}

func (model *UserModel) GetUserByEmail(ctx context.Context, email string) (*User, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	email = strings.ToLower(email)

	var user User
	query := "select id, uuid, first_name, last_name, email, password, role, is_active, profile_image, created_at, updated_at from users where email = ?"

	row := model.db.QueryRowContext(ctx, query, email)
	err := row.Scan(&user.ID, &user.UUID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Role, &user.IsActive, &user.ProfileImage, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {

			return nil, errors.New("no matching record found")
		}
		return nil, err
	}

	return &user, nil
}

func (model *UserModel) Insert(ctx context.Context, user *User) (int, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	query := `insert into users (first_name, last_name, email, password, role, is_active, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?) returning id`

	var id int
	err := model.db.QueryRowContext(ctx, query,
		user.FirstName, user.LastName, user.Email, user.Password, user.Role, user.IsActive, time.Now(), time.Now()).Scan(&id)
	if err != nil {

		return 0, err
	}

	return id, nil
}

func (model *UserModel) Activate(ctx context.Context, id int) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := "update users set is_active = true, updated_at = ? where id = ?"

	_, err := model.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {

		return err
	}

	return nil
}

func (model *UserModel) GetAll(ctx context.Context) ([]*User, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := `
		select
			id, uuid, first_name, last_name, email, role, is_active, profile_image, created_at, updated_at
		from
			users
		order by
			last_name, first_name
	`

	rows, err := model.db.QueryContext(ctx, query)
	if err != nil {

		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {

		var u User
		err := rows.Scan(
			&u.ID, &u.UUID, &u.FirstName, &u.LastName, &u.Email,
			&u.Role, &u.IsActive, &u.ProfileImage, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {

			return nil, err
		}
		users = append(users, &u)
	}

	if err = rows.Err(); err != nil {

		return nil, err
	}

	return users, nil
}

func (model *UserModel) GetAllPaginated(ctx context.Context, pageSize, currentPage int, sortCol, sortDir string) ([]*User, int, int, int, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	if pageSize < 1 {

		pageSize = 1
	}
	if currentPage < 1 {

		currentPage = 1
	}

	var totalRecords int
	countQuery := "select count(*) from users"
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
	case "email":
		orderClause = "email" + dir
	case "role":
		orderClause = "role" + dir
	case "status":
		orderClause = "is_active" + dir
	default:
		orderClause = "last_name" + dir + ", first_name asc"
	}

	query := `
		select
			id, uuid, first_name, last_name, email, role, is_active, profile_image, created_at, updated_at
		from
			users
		order by
			` + orderClause + `
		limit ? offset ?
	`

	rows, err := model.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {

		return nil, 0, 0, 0, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {

		var u User
		err := rows.Scan(
			&u.ID, &u.UUID, &u.FirstName, &u.LastName, &u.Email,
			&u.Role, &u.IsActive, &u.ProfileImage, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {

			return nil, 0, 0, 0, err
		}
		users = append(users, &u)
	}

	if err = rows.Err(); err != nil {

		return nil, 0, 0, 0, err
	}

	return users, currentPage, lastPage, totalRecords, nil
}

func (model *UserModel) Update(ctx context.Context, user *User) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	query := `update users set first_name = ?, last_name = ?, email = ?, role = ?, is_active = ?, updated_at = ? where id = ?`

	_, err := model.db.ExecContext(ctx, query, user.FirstName, user.LastName, user.Email, user.Role, user.IsActive, time.Now(), user.ID)
	if err != nil {

		return err
	}

	return nil
}

func (model *UserModel) Delete(ctx context.Context, id int) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := "delete from users where id = ?"

	_, err := model.db.ExecContext(ctx, query, id)
	if err != nil {

		return err
	}

	return nil
}

func (model *UserModel) UpdatePassword(ctx context.Context, user *User, hash string) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := "update users set password = ?, updated_at = ? where id = ?"

	_, err := model.db.ExecContext(ctx, query, hash, time.Now(), user.ID)
	if err != nil {

		return err
	}

	return nil
}

func (model *UserModel) UpdateProfileImage(ctx context.Context, id int, imagePath string) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	query := "update users set profile_image = ?, updated_at = ? where id = ?"

	_, err := model.db.ExecContext(ctx, query, imagePath, time.Now(), id)
	if err != nil {

		return err
	}

	return nil
}
