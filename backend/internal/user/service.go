package user

import (
	"context"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var validRoles = map[string]bool{"admin": true, "member": true, "viewer": true}
var validLanguages = map[string]bool{"en": true, "fa": true}
var validCalendarPrefs = map[string]bool{"gregorian": true, "jalali": true}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, orgID string) ([]*User, error) {
	users, err := s.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("user.Service.List: %w", err)
	}
	if users == nil {
		users = []*User{}
	}
	return users, nil
}

func (s *Service) Invite(ctx context.Context, orgID string, req InviteRequest) (*User, string, error) {
	if !validRoles[req.Role] {
		return nil, "", ErrForbidden
	}

	tempPassword, err := generateTempPassword()
	if err != nil {
		return nil, "", fmt.Errorf("user.Service.Invite: generate password: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("user.Service.Invite: hash: %w", err)
	}

	u, err := s.repo.Create(ctx, orgID, req.Name, req.Email, string(hash), req.Role)
	if err != nil {
		return nil, "", fmt.Errorf("user.Service.Invite: %w", err)
	}
	return u, tempPassword, nil
}

func (s *Service) GetByID(ctx context.Context, orgID, targetID string) (*User, error) {
	u, err := s.repo.GetByID(ctx, targetID, orgID)
	if err != nil {
		return nil, fmt.Errorf("user.Service.GetByID: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, callerID, targetID, orgID string, req UpdateProfileRequest) (*User, error) {
	if callerID != targetID {
		return nil, ErrForbidden
	}
	if !validLanguages[req.Language] {
		return nil, ErrForbidden
	}
	if !validCalendarPrefs[req.CalendarPref] {
		return nil, ErrForbidden
	}

	u, err := s.repo.UpdateProfile(ctx, targetID, orgID, req.Name, req.Timezone, req.Language, req.CalendarPref)
	if err != nil {
		return nil, fmt.Errorf("user.Service.UpdateProfile: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Service) UpdateRole(ctx context.Context, orgID, targetID, role string) (*User, error) {
	if !validRoles[role] {
		return nil, ErrForbidden
	}

	u, err := s.repo.UpdateRole(ctx, targetID, orgID, role)
	if err != nil {
		return nil, fmt.Errorf("user.Service.UpdateRole: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

func (s *Service) Deactivate(ctx context.Context, orgID, callerID, targetID string) error {
	// Block self-deactivation only if caller is the sole admin (business rule from CLAUDE.md)
	if callerID == targetID {
		count, err := s.repo.CountActiveAdmins(ctx, orgID)
		if err != nil {
			return fmt.Errorf("user.Service.Deactivate: %w", err)
		}
		if count <= 1 {
			return ErrLastAdmin
		}
	}

	if err := s.repo.Deactivate(ctx, targetID, orgID); err != nil {
		return fmt.Errorf("user.Service.Deactivate: %w", err)
	}
	return nil
}

// generateTempPassword produces a 12-character alphanumeric password using crypto/rand.
// Characters that are visually ambiguous (0/O, 1/l/I) are excluded.
func generateTempPassword() (string, error) {
	const charset = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateTempPassword: %w", err)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
