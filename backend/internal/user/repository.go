package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const userCols = `id, org_id, name, email, role, timezone, language, calendar_pref, is_active, created_at`

// scanUser accepts any Scan(dest ...any) error function so it works for both
// pgx.Row (QueryRow) and pgx.Rows (iteration).
func scanUser(scan func(...any) error) (*User, error) {
	u := &User{}
	err := scan(&u.ID, &u.OrgID, &u.Name, &u.Email, &u.Role, &u.Timezone, &u.Language, &u.CalendarPref, &u.IsActive, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListByOrg(ctx context.Context, orgID string) ([]*User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+userCols+` FROM users WHERE org_id = $1 AND is_active = TRUE ORDER BY name`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("user.Repository.ListByOrg: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := scanUser(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("user.Repository.ListByOrg: scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user.Repository.ListByOrg: rows: %w", err)
	}
	return users, nil
}

func (r *Repository) GetByID(ctx context.Context, id, orgID string) (*User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE id = $1 AND org_id = $2`,
		id, orgID,
	).Scan)
	if err != nil {
		return nil, fmt.Errorf("user.Repository.GetByID: %w", err)
	}
	return u, nil
}

func (r *Repository) Create(ctx context.Context, orgID, name, email, passwordHash, role string) (*User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx,
		`INSERT INTO users (org_id, name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+userCols,
		orgID, name, email, passwordHash, role,
	).Scan)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("user.Repository.Create: %w", err)
	}
	return u, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, id, orgID, name, timezone, language, calendarPref string) (*User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx,
		`UPDATE users SET name = $1, timezone = $2, language = $3, calendar_pref = $4
		 WHERE id = $5 AND org_id = $6
		 RETURNING `+userCols,
		name, timezone, language, calendarPref, id, orgID,
	).Scan)
	if err != nil {
		return nil, fmt.Errorf("user.Repository.UpdateProfile: %w", err)
	}
	return u, nil
}

func (r *Repository) UpdateRole(ctx context.Context, id, orgID, role string) (*User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx,
		`UPDATE users SET role = $1 WHERE id = $2 AND org_id = $3 RETURNING `+userCols,
		role, id, orgID,
	).Scan)
	if err != nil {
		return nil, fmt.Errorf("user.Repository.UpdateRole: %w", err)
	}
	return u, nil
}

func (r *Repository) Deactivate(ctx context.Context, id, orgID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET is_active = FALSE WHERE id = $1 AND org_id = $2`,
		id, orgID,
	)
	if err != nil {
		return fmt.Errorf("user.Repository.Deactivate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CountActiveAdmins(ctx context.Context, orgID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE org_id = $1 AND role = 'admin' AND is_active = TRUE`,
		orgID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("user.Repository.CountActiveAdmins: %w", err)
	}
	return count, nil
}
