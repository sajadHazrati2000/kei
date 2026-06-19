package availability

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// --- Slots ---

func (s *Service) GetSlots(ctx context.Context, callerID, callerRole, targetUserID, orgID string, from, to time.Time) ([]*Slot, error) {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, ErrInvalidInput
	}
	// Any authenticated user can view any member's slots in the same org.
	// Org scoping is enforced by the user_id check at the repository level —
	// the handler already verified targetUserID belongs to orgID via the JWT orgID.
	slots, err := s.repo.GetSlots(ctx, targetUserID, from, to)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetSlots: %w", err)
	}
	return slots, nil
}

func (s *Service) SetSlots(ctx context.Context, callerID, callerRole, targetUserID string, req SetSlotsRequest) ([]*Slot, error) {
	if callerRole != "admin" && callerID != targetUserID {
		return nil, ErrForbidden
	}
	if req.From.IsZero() || req.To.IsZero() || !req.To.After(req.From) {
		return nil, ErrInvalidInput
	}
	for _, sl := range req.Slots {
		if !validStatuses[sl.Status] {
			return nil, ErrInvalidInput
		}
		if !sl.To.After(sl.From) {
			return nil, ErrInvalidInput
		}
	}

	merged := mergeAdjacent(req.Slots)

	slots, err := s.repo.ReplaceSlots(ctx, targetUserID, req.From, req.To, merged)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.SetSlots: %w", err)
	}
	return slots, nil
}

// --- Recurring templates ---

func (s *Service) GetTemplates(ctx context.Context, targetUserID string) ([]*RecurringTemplate, error) {
	t, err := s.repo.GetTemplates(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTemplates: %w", err)
	}
	return t, nil
}

func (s *Service) SetTemplates(ctx context.Context, callerID, callerRole, targetUserID string, req SetRecurringRequest) ([]*RecurringTemplate, error) {
	if callerRole != "admin" && callerID != targetUserID {
		return nil, ErrForbidden
	}
	for _, t := range req.Templates {
		if !validPatterns[t.Pattern] {
			return nil, ErrInvalidInput
		}
		if !validStatuses[t.Status] {
			return nil, ErrInvalidInput
		}
		if t.ValidFrom == "" || t.StartTime == "" || t.EndTime == "" {
			return nil, ErrInvalidInput
		}
		for _, d := range t.DaysOfWeek {
			if d < 0 || d > 6 {
				return nil, ErrInvalidInput
			}
		}
	}
	templates, err := s.repo.ReplaceTemplates(ctx, targetUserID, req.Templates)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.SetTemplates: %w", err)
	}
	return templates, nil
}

// --- CSV import ---

// ImportCSV parses a CSV with header date,start_time,end_time,status (times in UTC HH:MM).
// Merges adjacent same-type slots per day and replaces existing slots for those dates only.
func (s *Service) ImportCSV(ctx context.Context, callerID, callerRole, targetUserID string, r io.Reader) (int, error) {
	if callerRole != "admin" && callerID != targetUserID {
		return 0, ErrForbidden
	}

	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		return 0, fmt.Errorf("%w: missing CSV header", ErrInvalidInput)
	}
	if len(header) < 4 {
		return 0, fmt.Errorf("%w: CSV must have columns: date,start_time,end_time,status", ErrInvalidInput)
	}

	byDate := map[string][]SlotInput{}
	rowNum := 1
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("%w: row %d: %s", ErrInvalidInput, rowNum, err)
		}
		rowNum++
		if len(row) < 4 {
			return 0, fmt.Errorf("%w: row %d has fewer than 4 columns", ErrInvalidInput, rowNum)
		}

		dateStr := strings.TrimSpace(row[0])
		startStr := strings.TrimSpace(row[1])
		endStr := strings.TrimSpace(row[2])
		status := strings.TrimSpace(row[3])

		if !validStatuses[status] {
			return 0, fmt.Errorf("%w: row %d: invalid status %q", ErrInvalidInput, rowNum, status)
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return 0, fmt.Errorf("%w: row %d: invalid date %q", ErrInvalidInput, rowNum, dateStr)
		}
		startT, err := time.Parse("15:04", startStr)
		if err != nil {
			return 0, fmt.Errorf("%w: row %d: invalid start_time %q", ErrInvalidInput, rowNum, startStr)
		}
		endT, err := time.Parse("15:04", endStr)
		if err != nil {
			return 0, fmt.Errorf("%w: row %d: invalid end_time %q", ErrInvalidInput, rowNum, endStr)
		}

		from := time.Date(date.Year(), date.Month(), date.Day(), startT.Hour(), startT.Minute(), 0, 0, time.UTC)
		to := time.Date(date.Year(), date.Month(), date.Day(), endT.Hour(), endT.Minute(), 0, 0, time.UTC)
		if !to.After(from) {
			return 0, fmt.Errorf("%w: row %d: end_time must be after start_time", ErrInvalidInput, rowNum)
		}

		byDate[dateStr] = append(byDate[dateStr], SlotInput{From: from, To: to, Status: status})
	}

	if len(byDate) == 0 {
		return 0, nil
	}

	// Merge adjacent per date before importing
	total := 0
	merged := map[string][]SlotInput{}
	for dateStr, slots := range byDate {
		m := mergeAdjacent(slots)
		merged[dateStr] = m
		total += len(m)
	}

	if err := s.repo.ImportSlots(ctx, targetUserID, merged); err != nil {
		return 0, fmt.Errorf("availability.Service.ImportCSV: %w", err)
	}
	return total, nil
}

// --- Team ---

func (s *Service) GetTeamAvailability(ctx context.Context, orgID string, from, to time.Time) ([]UserSlots, error) {
	if from.IsZero() || to.IsZero() || !to.After(from) {
		return nil, ErrInvalidInput
	}
	users, err := s.repo.GetActiveUsers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamAvailability: %w", err)
	}

	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}

	slotsByUser, err := s.repo.GetSlotsForUsers(ctx, ids, from, to)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamAvailability: %w", err)
	}

	result := make([]UserSlots, len(users))
	for i, u := range users {
		slots := slotsByUser[u.ID]
		if slots == nil {
			slots = []*Slot{}
		}
		result[i] = UserSlots{
			User:  UserSummary{ID: u.ID, Name: u.Name, Timezone: u.Timezone, Role: u.Role},
			Slots: slots,
		}
	}
	return result, nil
}

func (s *Service) GetTeamOverlap(ctx context.Context, orgID, dateStr string) (*OverlapResult, error) {
	if dateStr == "" {
		dateStr = time.Now().UTC().Format("2006-01-02")
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid date %q", ErrInvalidInput, dateStr)
	}

	overlap, err := s.repo.GetOrgOverlap(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamOverlap: %w", err)
	}

	windowFrom, err := combineDateHHMM(date, overlap.Start)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamOverlap: overlap_start: %w", err)
	}
	windowTo, err := combineDateHHMM(date, overlap.End)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamOverlap: overlap_end: %w", err)
	}

	users, err := s.repo.GetActiveUsers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamOverlap: %w", err)
	}

	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}

	slotsByUser, err := s.repo.GetSlotsForUsers(ctx, ids, windowFrom, windowTo)
	if err != nil {
		return nil, fmt.Errorf("availability.Service.GetTeamOverlap: %w", err)
	}

	members := make([]UserSlots, len(users))
	for i, u := range users {
		slots := slotsByUser[u.ID]
		if slots == nil {
			slots = []*Slot{}
		}
		members[i] = UserSlots{
			User:  UserSummary{ID: u.ID, Name: u.Name, Timezone: u.Timezone, Role: u.Role},
			Slots: slots,
		}
	}

	return &OverlapResult{
		Date:         dateStr,
		OverlapStart: overlap.Start,
		OverlapEnd:   overlap.End,
		Members:      members,
	}, nil
}

// --- helpers ---

// mergeAdjacent sorts slots by From and merges contiguous or overlapping slots
// that share the same status into a single slot.
func mergeAdjacent(slots []SlotInput) []SlotInput {
	if len(slots) == 0 {
		return slots
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].From.Before(slots[j].From)
	})
	merged := []SlotInput{slots[0]}
	for _, s := range slots[1:] {
		last := &merged[len(merged)-1]
		// Adjacent means last.To == s.From; overlapping means last.To > s.From.
		// Only merge when status matches.
		if last.Status == s.Status && !last.To.Before(s.From) {
			if s.To.After(last.To) {
				last.To = s.To
			}
		} else {
			merged = append(merged, s)
		}
	}
	return merged
}

// combineDateHHMM combines a date (year/month/day) with a "HH:MM" string into a UTC time.
func combineDateHHMM(date time.Time, hhmm string) (time.Time, error) {
	parts := strings.SplitN(hhmm, ":", 2)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid HH:MM %q", hhmm)
	}
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC), nil
}
