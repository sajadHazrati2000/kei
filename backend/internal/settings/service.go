package settings

import (
	"context"
	"fmt"
	"time"
)

var validLanguages = map[string]bool{"en": true, "fa": true}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetSettings(ctx context.Context, orgID string) (*OrgSettings, error) {
	org, err := s.repo.GetSettings(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("settings.Service.GetSettings: %w", err)
	}
	if org == nil {
		return nil, ErrNotFound
	}
	return org, nil
}

func (s *Service) UpdateSettings(ctx context.Context, orgID string, req UpdateSettingsRequest) (*OrgSettings, error) {
	if err := validateSettingsRequest(req); err != nil {
		return nil, err
	}
	org, err := s.repo.UpdateSettings(ctx, orgID, req)
	if err != nil {
		return nil, fmt.Errorf("settings.Service.UpdateSettings: %w", err)
	}
	return org, nil
}

func (s *Service) ListBlockedDays(ctx context.Context, orgID string) ([]*BlockedDay, error) {
	days, err := s.repo.ListBlockedDays(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("settings.Service.ListBlockedDays: %w", err)
	}
	return days, nil
}

func (s *Service) AddBlockedDay(ctx context.Context, orgID, callerID string, req AddBlockedDayRequest) (*BlockedDay, error) {
	if req.BlockedDate == "" {
		return nil, fmt.Errorf("%w: blocked_date is required", ErrInvalidInput)
	}
	if _, err := time.Parse("2006-01-02", req.BlockedDate); err != nil {
		return nil, fmt.Errorf("%w: blocked_date must be YYYY-MM-DD", ErrInvalidInput)
	}
	day, err := s.repo.AddBlockedDay(ctx, orgID, callerID, req.BlockedDate, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("settings.Service.AddBlockedDay: %w", err)
	}
	return day, nil
}

func (s *Service) DeleteBlockedDay(ctx context.Context, orgID, dateStr string) error {
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return fmt.Errorf("%w: date must be YYYY-MM-DD", ErrInvalidInput)
	}
	if err := s.repo.DeleteBlockedDay(ctx, orgID, dateStr); err != nil {
		return fmt.Errorf("settings.Service.DeleteBlockedDay: %w", err)
	}
	return nil
}

// validateSettingsRequest checks all updatable fields before hitting the DB.
func validateSettingsRequest(req UpdateSettingsRequest) error {
	if req.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if req.Timezone == "" {
		return fmt.Errorf("%w: timezone is required", ErrInvalidInput)
	}
	if !validLanguages[req.DefaultLanguage] {
		return fmt.Errorf("%w: default_language must be 'en' or 'fa'", ErrInvalidInput)
	}
	if err := validateHHMM("overlap_start", req.OverlapStart); err != nil {
		return err
	}
	if err := validateHHMM("overlap_end", req.OverlapEnd); err != nil {
		return err
	}
	if err := validateHHMM("working_start", req.WorkingStart); err != nil {
		return err
	}
	if err := validateHHMM("working_end", req.WorkingEnd); err != nil {
		return err
	}
	if !timeBefore(req.OverlapStart, req.OverlapEnd) {
		return fmt.Errorf("%w: overlap_start must be before overlap_end", ErrInvalidInput)
	}
	if !timeBefore(req.WorkingStart, req.WorkingEnd) {
		return fmt.Errorf("%w: working_start must be before working_end", ErrInvalidInput)
	}
	for _, d := range req.WorkingDays {
		if d < 0 || d > 6 {
			return fmt.Errorf("%w: working_days values must be 0–6 (0=Sun, 6=Sat)", ErrInvalidInput)
		}
	}
	return nil
}

func validateHHMM(field, v string) error {
	if v == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidInput, field)
	}
	if _, err := time.Parse("15:04", v); err != nil {
		return fmt.Errorf("%w: %s must be HH:MM (e.g. 09:00)", ErrInvalidInput, field)
	}
	return nil
}

// timeBefore returns true when hhmm a comes strictly before hhmm b.
func timeBefore(a, b string) bool {
	ta, err1 := time.Parse("15:04", a)
	tb, err2 := time.Parse("15:04", b)
	if err1 != nil || err2 != nil {
		return false
	}
	return ta.Before(tb)
}
