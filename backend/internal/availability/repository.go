package availability

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// slotCols decomposes TSTZRANGE into two TIMESTAMPTZ and casts the nullable UUID to text.
const slotCols = `id, user_id,
	lower(slot_range) AS from_t, upper(slot_range) AS to_t,
	status, is_override, recurrence_id::text, created_at`

func scanSlot(scan func(...any) error) (*Slot, error) {
	s := &Slot{}
	var recID *string
	err := scan(&s.ID, &s.UserID, &s.From, &s.To, &s.Status, &s.IsOverride, &recID, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.RecurrenceID = recID
	return s, nil
}

func collectSlots(rows pgx.Rows) ([]*Slot, error) {
	var slots []*Slot
	for rows.Next() {
		s, err := scanSlot(rows.Scan)
		if err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if slots == nil {
		slots = []*Slot{}
	}
	return slots, nil
}

// tmplCols reads the recurrence_templates columns as portable scalar types.
const tmplCols = `id, user_id, pattern,
	COALESCE(array_to_string(days_of_week, ','), '') AS days_csv,
	to_char(start_time::time, 'HH24:MI') AS start_t,
	to_char(end_time::time,   'HH24:MI') AS end_t,
	status, valid_from::text, valid_until::text`

func scanTemplate(scan func(...any) error) (*RecurringTemplate, error) {
	t := &RecurringTemplate{}
	var daysCsv string
	var validUntil *string
	err := scan(&t.ID, &t.UserID, &t.Pattern, &daysCsv,
		&t.StartTime, &t.EndTime, &t.Status, &t.ValidFrom, &validUntil)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.DaysOfWeek = parseIntCSV(daysCsv)
	t.ValidUntil = validUntil
	return t, nil
}

// parseIntCSV converts "1,2,3" → []int{1,2,3}; "" → []int{}.
func parseIntCSV(s string) []int {
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
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
		parts[i] = strconv.Itoa(v)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// --- Availability slots ---

func (r *Repository) GetSlots(ctx context.Context, userID string, from, to time.Time) ([]*Slot, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+slotCols+`
		 FROM availability_slots
		 WHERE user_id = $1
		   AND slot_range && tstzrange($2::timestamptz, $3::timestamptz, '[)')
		 ORDER BY lower(slot_range)`,
		userID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.GetSlots: %w", err)
	}
	defer rows.Close()
	slots, err := collectSlots(rows)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.GetSlots: %w", err)
	}
	return slots, nil
}

// ReplaceSlots deletes all slots in [from, to) for userID and inserts the new list.
// Returns ErrConflict on GIST exclusion violation (overlapping busy/focus slots).
func (r *Repository) ReplaceSlots(ctx context.Context, userID string, from, to time.Time, slots []SlotInput) ([]*Slot, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.ReplaceSlots: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM availability_slots
		 WHERE user_id = $1
		   AND slot_range && tstzrange($2::timestamptz, $3::timestamptz, '[)')`,
		userID, from, to,
	); err != nil {
		return nil, fmt.Errorf("availability.Repository.ReplaceSlots: delete: %w", err)
	}

	var result []*Slot
	for _, s := range slots {
		inserted, err := insertSlotTx(ctx, tx, userID, s)
		if err != nil {
			return nil, wrapSlotErr("ReplaceSlots", err)
		}
		result = append(result, inserted)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("availability.Repository.ReplaceSlots: commit: %w", err)
	}
	if result == nil {
		result = []*Slot{}
	}
	return result, nil
}

// ImportSlots replaces slots day-by-day for each date present in byDate.
func (r *Repository) ImportSlots(ctx context.Context, userID string, byDate map[string][]SlotInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("availability.Repository.ImportSlots: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for dateStr, slots := range byDate {
		if _, err := tx.Exec(ctx,
			`DELETE FROM availability_slots
			 WHERE user_id = $1
			   AND slot_range && tstzrange(
			         $2::date::timestamptz,
			         ($2::date + INTERVAL '1 day')::timestamptz, '[)')`,
			userID, dateStr,
		); err != nil {
			return fmt.Errorf("availability.Repository.ImportSlots: delete %s: %w", dateStr, err)
		}
		for _, s := range slots {
			if _, err := insertSlotTx(ctx, tx, userID, s); err != nil {
				return wrapSlotErr("ImportSlots", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("availability.Repository.ImportSlots: commit: %w", err)
	}
	return nil
}

func insertSlotTx(ctx context.Context, tx pgx.Tx, userID string, s SlotInput) (*Slot, error) {
	return scanSlot(tx.QueryRow(ctx,
		`INSERT INTO availability_slots (user_id, slot_range, status)
		 VALUES ($1, tstzrange($2::timestamptz, $3::timestamptz, '[)'), $4)
		 RETURNING `+slotCols,
		userID, s.From, s.To, s.Status,
	).Scan)
}

func wrapSlotErr(op string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
		return ErrConflict
	}
	return fmt.Errorf("availability.Repository.%s: %w", op, err)
}

// --- Recurring templates ---

func (r *Repository) GetTemplates(ctx context.Context, userID string) ([]*RecurringTemplate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+tmplCols+` FROM recurrence_templates WHERE user_id = $1 ORDER BY valid_from, id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.GetTemplates: %w", err)
	}
	defer rows.Close()

	var templates []*RecurringTemplate
	for rows.Next() {
		t, err := scanTemplate(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("availability.Repository.GetTemplates: scan: %w", err)
		}
		templates = append(templates, t)
	}
	if templates == nil {
		templates = []*RecurringTemplate{}
	}
	return templates, rows.Err()
}

// ReplaceTemplates deletes all templates for the user and inserts the new set.
func (r *Repository) ReplaceTemplates(ctx context.Context, userID string, inputs []RecurringTemplateInput) ([]*RecurringTemplate, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.ReplaceTemplates: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM recurrence_templates WHERE user_id = $1`, userID); err != nil {
		return nil, fmt.Errorf("availability.Repository.ReplaceTemplates: delete: %w", err)
	}

	var templates []*RecurringTemplate
	for _, inp := range inputs {
		var validUntilArg interface{}
		if inp.ValidUntil != nil {
			validUntilArg = *inp.ValidUntil
		}

		t, err := scanTemplate(tx.QueryRow(ctx,
			`INSERT INTO recurrence_templates
			   (user_id, pattern, days_of_week, start_time, end_time, status, valid_from, valid_until)
			 VALUES ($1, $2, $3::int[], $4::timetz, $5::timetz, $6, $7::date, $8::date)
			 RETURNING `+tmplCols,
			userID, inp.Pattern, intSliceToLiteral(inp.DaysOfWeek),
			inp.StartTime+"+00", inp.EndTime+"+00",
			inp.Status, inp.ValidFrom, validUntilArg,
		).Scan)
		if err != nil {
			return nil, fmt.Errorf("availability.Repository.ReplaceTemplates: insert: %w", err)
		}
		templates = append(templates, t)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("availability.Repository.ReplaceTemplates: commit: %w", err)
	}
	if templates == nil {
		templates = []*RecurringTemplate{}
	}
	return templates, nil
}

// --- Team queries ---

type OrgOverlap struct {
	Start string // "HH:MM"
	End   string // "HH:MM"
}

func (r *Repository) GetOrgOverlap(ctx context.Context, orgID string) (OrgOverlap, error) {
	var o OrgOverlap
	err := r.pool.QueryRow(ctx,
		`SELECT to_char(overlap_start::time, 'HH24:MI'),
		        to_char(overlap_end::time,   'HH24:MI')
		 FROM organizations WHERE id = $1`,
		orgID,
	).Scan(&o.Start, &o.End)
	if err != nil {
		return OrgOverlap{}, fmt.Errorf("availability.Repository.GetOrgOverlap: %w", err)
	}
	return o, nil
}

type ActiveUser struct {
	ID       string
	Name     string
	Timezone string
	Role     string
}

func (r *Repository) GetActiveUsers(ctx context.Context, orgID string) ([]ActiveUser, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, timezone, role FROM users
		 WHERE org_id = $1 AND is_active = TRUE ORDER BY name`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.GetActiveUsers: %w", err)
	}
	defer rows.Close()

	var users []ActiveUser
	for rows.Next() {
		var u ActiveUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Timezone, &u.Role); err != nil {
			return nil, fmt.Errorf("availability.Repository.GetActiveUsers: scan: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// GetSlotsForUsers returns slots for multiple users in [from, to), keyed by userID.
func (r *Repository) GetSlotsForUsers(ctx context.Context, userIDs []string, from, to time.Time) (map[string][]*Slot, error) {
	if len(userIDs) == 0 {
		return map[string][]*Slot{}, nil
	}

	// Build: WHERE user_id IN ($1, $2, ...) AND slot_range && tstzrange($N, $N+1, '[)')
	placeholders := make([]string, len(userIDs))
	args := make([]any, len(userIDs)+2)
	for i, id := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(userIDs)] = from
	args[len(userIDs)+1] = to

	q := fmt.Sprintf(
		`SELECT `+slotCols+`
		 FROM availability_slots
		 WHERE user_id IN (%s)
		   AND slot_range && tstzrange($%d::timestamptz, $%d::timestamptz, '[)')
		 ORDER BY user_id, lower(slot_range)`,
		strings.Join(placeholders, ","),
		len(userIDs)+1, len(userIDs)+2,
	)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("availability.Repository.GetSlotsForUsers: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]*Slot)
	for rows.Next() {
		s, err := scanSlot(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("availability.Repository.GetSlotsForUsers: scan: %w", err)
		}
		result[s.UserID] = append(result[s.UserID], s)
	}
	return result, rows.Err()
}
