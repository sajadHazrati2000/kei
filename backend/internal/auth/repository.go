package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) OrgCount(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM organizations`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("auth.Repository.OrgCount: %w", err)
	}
	return count, nil
}

func (r *Repository) CreateOrgAndAdmin(ctx context.Context, orgName, orgSlug, orgTZ, adminName, email, passwordHash string) (*User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.CreateOrgAndAdmin: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var orgID string
	err = tx.QueryRow(ctx,
		`INSERT INTO organizations (name, slug, timezone) VALUES ($1, $2, $3) RETURNING id`,
		orgName, orgSlug, orgTZ,
	).Scan(&orgID)
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.CreateOrgAndAdmin: insert org: %w", err)
	}

	u := &User{}
	err = tx.QueryRow(ctx,
		`INSERT INTO users (org_id, name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4, 'admin')
		 RETURNING id, org_id, name, email, password_hash, role, timezone, language, calendar_pref, is_active`,
		orgID, adminName, email, passwordHash,
	).Scan(&u.ID, &u.OrgID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Timezone, &u.Language, &u.CalendarPref, &u.IsActive)
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.CreateOrgAndAdmin: insert user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("auth.Repository.CreateOrgAndAdmin: commit: %w", err)
	}
	return u, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, org_id, name, email, password_hash, role, timezone, language, calendar_pref, is_active
		 FROM users WHERE email = $1 LIMIT 1`,
		email,
	).Scan(&u.ID, &u.OrgID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Timezone, &u.Language, &u.CalendarPref, &u.IsActive)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.GetUserByEmail: %w", err)
	}
	return u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, org_id, name, email, password_hash, role, timezone, language, calendar_pref, is_active
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.OrgID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Timezone, &u.Language, &u.CalendarPref, &u.IsActive)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.GetUserByID: %w", err)
	}
	return u, nil
}

func (r *Repository) StoreRefreshToken(ctx context.Context, userID, jti string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, jti, expires_at) VALUES ($1, $2, $3)`,
		userID, jti, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("auth.Repository.StoreRefreshToken: %w", err)
	}
	return nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, jti string) (*RefreshToken, error) {
	rt := &RefreshToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT jti, user_id, expires_at FROM refresh_tokens WHERE jti = $1`,
		jti,
	).Scan(&rt.JTI, &rt.UserID, &rt.ExpiresAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.GetRefreshToken: %w", err)
	}
	return rt, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, jti string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE jti = $1`, jti)
	if err != nil {
		return fmt.Errorf("auth.Repository.DeleteRefreshToken: %w", err)
	}
	return nil
}

// CreatePasswordResetToken stores a new reset token for the given user.
func (r *Repository) CreatePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("auth.Repository.CreatePasswordResetToken: %w", err)
	}
	return nil
}

// GetPasswordResetToken fetches a token record by token string.
// Returns nil, nil when not found.
func (r *Repository) GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	rt := &PasswordResetToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, token, expires_at, used_at
		 FROM password_reset_tokens WHERE token = $1`,
		token,
	).Scan(&rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.UsedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("auth.Repository.GetPasswordResetToken: %w", err)
	}
	return rt, nil
}

// ConfirmPasswordReset runs three steps atomically:
//  1. Marks the token as used (fails if already used — 0 rows affected).
//  2. Updates the user's password hash.
//  3. Deletes all refresh tokens for that user, invalidating all sessions.
func (r *Repository) ConfirmPasswordReset(ctx context.Context, token, userID, passwordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("auth.Repository.ConfirmPasswordReset: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW()
		 WHERE token = $1 AND used_at IS NULL`,
		token,
	)
	if err != nil {
		return fmt.Errorf("auth.Repository.ConfirmPasswordReset: mark used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResetTokenInvalid
	}

	if _, err := tx.Exec(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`,
		passwordHash, userID,
	); err != nil {
		return fmt.Errorf("auth.Repository.ConfirmPasswordReset: update password: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	); err != nil {
		return fmt.Errorf("auth.Repository.ConfirmPasswordReset: delete sessions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("auth.Repository.ConfirmPasswordReset: commit: %w", err)
	}
	return nil
}
