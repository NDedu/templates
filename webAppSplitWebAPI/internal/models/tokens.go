package models

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"
	"webProj/internal/tokens"
)

type TokenModel struct {
	db DBTX
}

func (model *TokenModel) Insert(ctx context.Context, t *tokens.Token, u *User) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	// Delete only expired tokens to prevent database bloat, allowing multiple active sessions
	_, err := model.db.ExecContext(ctx, `DELETE FROM tokens WHERE user_id = $1 AND scope = $2 AND expiry < $3`, u.ID, t.Scope, time.Now())
	if err != nil {

		return err
	}

	query := `
		INSERT INTO tokens
	        (user_id, token_hash, expiry, scope, created_at, updated_at)
	    VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = model.db.ExecContext(ctx, query,
		u.ID,
		t.Hash,
		t.Expiry,
		t.Scope,
		time.Now(),
		time.Now(),
	)
	if err != nil {

		return err
	}

	return nil
}

func (model *TokenModel) GetUserByToken(ctx context.Context, plainToken string) (*User, error) {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	hash := sha256.Sum256([]byte(plainToken))

	query := `
		SELECT u.id, u.uuid, u.first_name, u.last_name, u.email, u.role, u.is_active, u.created_at, u.updated_at
		FROM users u
		INNER JOIN tokens t ON u.id = t.user_id
		WHERE t.token_hash = $1
		AND t.expiry > $2
		AND t.scope = $3
	`

	var user User
	err := model.db.QueryRowContext(ctx, query, hash[:], time.Now(), tokens.ScopeAuthentication).Scan(
		&user.ID,
		&user.UUID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return nil, errors.New("No record")
		}
		return nil, err
	}

	return &user, nil
}

// FIX: this deletes users session on all devices, allow users to hold multiple tokens and add a worker that deletes expired tokes
// without this, the tokens stay and grow in the db
func (model *TokenModel) DeleteByToken(ctx context.Context, plainToken string) error {

	ctx, cancel := dbCtx(ctx)
	defer cancel()

	hash := sha256.Sum256([]byte(plainToken))
	_, err := model.db.ExecContext(ctx, `DELETE FROM tokens WHERE token_hash = $1`, hash[:])
	return err
}
