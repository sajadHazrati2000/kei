package settings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// orgCols selects all org settings columns using the portable casting rules
// established for TIMETZ (::time → to_char) and INT[] (array_to_string).
const orgCols = `id, name, slug, timezone,
	to_char(overlap_start::time, 'HH24:MI') AS overlap_start,
	to_char(overlap_end::time,   'HH24:MI') AS overlap_end,
	default_language,
	COALESCE(array_to_string(working_days, ','), '') AS working_days_csv,
	to_char(working_start::time, 'HH24:MI') AS working_start,
	to_char(working_end::time,   'HH24:MI') AS working_end`

func scanOrg(scan func(...any) error) (*OrgSettings, error) {
	s := &OrgSettings{}
	var daysCsv string
	err := scan(&s.ID, &s.Name, &s.Slug, &s.Timezone,
		&s.OverlapStart, &s.OverlapEnd,
		&s.DefaultLanguage, &daysCsv,
		&s.WorkingStart, &s.WorkingEnd)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.WorkingDays = parseIntCSV(daysCsv)
	return s, nil
}

// parseIntCSV converts "1,2,3" → []int{1,2,3}; "" → []int{}.
func parseIntCSV(s string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		var v int
		if _, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &v); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// intSliceToLiteral converts []int{1,2,3} → "{1,2,3}" for PostgreSQL int[] literal.
func intSliceToLiteral(ints []int) string {
	if len(ints) == 0 {
		return "{}"
	}
	parts := make([]string, len(ints))
	for i, v := range ints {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetSettings returns the org settings row. Returns nil, nil when not found.
func (r *Repository) GetSettings(ctx context.Context, orgID string) (*OrgSettings, error) {
	s, err := scanOrg(r.pool.QueryRow(ctx,
		`SELECT `+orgCols+` FROM organizations WHERE id = $1`, orgID,
	).Scan)
	if err != nil {
		return nil, fmt.Errorf("settings.Repository.GetSettings: %w", err)
	}
	return s, nil
}

// UpdateSettings replaces updatable org fields and returns the new state.
func (r *Repository) UpdateSettings(ctx context.Context, orgID string, req UpdateSettingsRequest) (*OrgSettings, error) {
	s, err := scanOrg(r.pool.QueryRow(ctx,
		`UPDATE organizations SET
		    name             = $1,
		    timezone         = $2,
		    overlap_start    = $3::timetz,
		    overlap_end      = $4::timetz,
		    default_language = $5,
		    working_days     = $6::int[],
		    working_start    = $7::timetz,
		    working_end      = $8::timetz
		 WHERE id = $9
		 RETURNING `+orgCols,
		req.Name,
		req.Timezone,
		req.OverlapStart+"+00",
		req.OverlapEnd+"+00",
		req.DefaultLanguage,
		intSliceToLiteral(req.WorkingDays),
		req.WorkingStart+"+00",
		req.WorkingEnd+"+00",
		orgID,
	).Scan)
	if err != nil {
		return nil, fmt.Errorf("settings.Repository.UpdateSettings: %w", err)
	}
	if s == nil {
		return nil, ErrNotFound
	}
	return s, nil
}

// --- Blocked days ---

const blockedCols = `id, org_id, blocked_date::text, reason, created_by::text, created_at`

func scanBlockedDay(scan func(...any) error) (*BlockedDay, error) {
	d := &BlockedDay{}
	err := scan(&d.ID, &d.OrgID, &d.BlockedDate, &d.Reason, &d.CreatedBy, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

// ListBlockedDays returns all blocked days for an org, ordered by date.
func (r *Repository) ListBlockedDays(ctx context.Context, orgID string) ([]*BlockedDay, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+blockedCols+`
		 FROM blocked_days WHERE org_id = $1 ORDER BY blocked_date`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("settings.Repository.ListBlockedDays: %w", err)
	}
	defer rows.Close()

	var days []*BlockedDay
	for rows.Next() {
		d, err := scanBlockedDay(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("settings.Repository.ListBlockedDays: scan: %w", err)
		}
		days = append(days, d)
	}
	if days == nil {
		days = []*BlockedDay{}
	}
	return days, rows.Err()
}

// AddBlockedDay inserts a new blocked day. Returns ErrAlreadyExists on duplicate.
func (r *Repository) AddBlockedDay(ctx context.Context, orgID, callerID, dateStr string, reason *string) (*BlockedDay, error) {
	d, err := scanBlockedDay(r.pool.QueryRow(ctx,
		`INSERT INTO blocked_days (org_id, blocked_date, reason, created_by)
		 VALUES ($1, $2::date, $3, $4)
		 RETURNING `+blockedCols,
		orgID, dateStr, reason, callerID,
	).Scan)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("settings.Repository.AddBlockedDay: %w", err)
	}
	return d, nil
}

// DeleteBlockedDay removes a blocked day by date. Returns ErrNotFound when the
// date has no entry for this org.
func (r *Repository) DeleteBlockedDay(ctx context.Context, orgID, dateStr string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM blocked_days WHERE org_id = $1 AND blocked_date = $2::date`,
		orgID, dateStr,
	)
	if err != nil {
		return fmt.Errorf("settings.Repository.DeleteBlockedDay: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
