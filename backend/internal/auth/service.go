package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSetupDone        = errors.New("setup already completed")
	ErrInvalidCreds     = errors.New("invalid credentials")
	ErrInactiveUser     = errors.New("account is inactive")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrResetTokenInvalid = errors.New("invalid, expired, or already-used reset token")
)

type Service struct {
	repo             *Repository
	jwtSecret        string
	jwtRefreshSecret string
}

func NewService(repo *Repository, jwtSecret, jwtRefreshSecret string) *Service {
	return &Service{
		repo:             repo,
		jwtSecret:        jwtSecret,
		jwtRefreshSecret: jwtRefreshSecret,
	}
}

// IsSetupDone returns true when at least one organisation exists in the DB.
// Used by the frontend setup guard without requiring authentication.
func (s *Service) IsSetupDone(ctx context.Context) (bool, error) {
	count, err := s.repo.OrgCount(ctx)
	if err != nil {
		return false, fmt.Errorf("auth.Service.IsSetupDone: %w", err)
	}
	return count > 0, nil
}

func (s *Service) Setup(ctx context.Context, req SetupRequest) (*TokenResponse, error) {
	count, err := s.repo.OrgCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.Setup: %w", err)
	}
	if count > 0 {
		return nil, ErrSetupDone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.Setup: hash: %w", err)
	}

	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}

	user, err := s.repo.CreateOrgAndAdmin(ctx, req.OrgName, req.OrgSlug, tz, req.AdminName, req.Email, string(hash))
	if err != nil {
		return nil, fmt.Errorf("auth.Service.Setup: %w", err)
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.Login: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCreds
	}
	if !user.IsActive {
		return nil, ErrInactiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, tokenStr string) (*TokenResponse, error) {
	claims, err := validateRefreshToken(s.jwtRefreshSecret, tokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	rt, err := s.repo.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.Refresh: %w", err)
	}
	if rt == nil || rt.UserID != claims.UserID {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.Refresh: %w", err)
	}
	if user == nil || !user.IsActive {
		return nil, ErrInvalidToken
	}

	// Rotate: delete the old token, issue a fresh pair
	if err := s.repo.DeleteRefreshToken(ctx, claims.ID); err != nil {
		return nil, fmt.Errorf("auth.Service.Refresh: delete old token: %w", err)
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Logout(ctx context.Context, tokenStr string) error {
	claims, err := validateRefreshToken(s.jwtRefreshSecret, tokenStr)
	if err != nil {
		// Already invalid — nothing to invalidate server-side
		return nil
	}
	return s.repo.DeleteRefreshToken(ctx, claims.ID)
}

// RequestPasswordReset generates a 1-hour reset token for the given email and
// stores it. Returns the token string (empty when email is not found so the
// handler can always respond 200 without leaking whether the address exists).
func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("auth.Service.RequestPasswordReset: %w", err)
	}
	if user == nil || !user.IsActive {
		return "", nil // unknown email — return empty token, caller always 200s
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth.Service.RequestPasswordReset: generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	if err := s.repo.CreatePasswordResetToken(ctx, user.ID, token, time.Now().Add(time.Hour)); err != nil {
		return "", fmt.Errorf("auth.Service.RequestPasswordReset: %w", err)
	}
	return token, nil
}

// ConfirmPasswordReset validates the token, hashes the new password, and
// atomically updates the user record while invalidating all existing sessions.
func (s *Service) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	rt, err := s.repo.GetPasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("auth.Service.ConfirmPasswordReset: %w", err)
	}
	if rt == nil || rt.UsedAt != nil || time.Now().After(rt.ExpiresAt) {
		return ErrResetTokenInvalid
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("auth.Service.ConfirmPasswordReset: hash: %w", err)
	}

	if err := s.repo.ConfirmPasswordReset(ctx, token, rt.UserID, string(hash)); err != nil {
		return fmt.Errorf("auth.Service.ConfirmPasswordReset: %w", err)
	}
	return nil
}

func (s *Service) issueTokens(ctx context.Context, user *User) (*TokenResponse, error) {
	accessToken, err := generateAccessToken(s.jwtSecret, user.ID, user.OrgID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.issueTokens: %w", err)
	}

	refreshToken, jti, expiresAt, err := generateRefreshToken(s.jwtRefreshSecret, user.ID)
	if err != nil {
		return nil, fmt.Errorf("auth.Service.issueTokens: %w", err)
	}

	if err := s.repo.StoreRefreshToken(ctx, user.ID, jti, expiresAt); err != nil {
		return nil, fmt.Errorf("auth.Service.issueTokens: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserInfo: UserInfo{
			ID:    user.ID,
			OrgID: user.OrgID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}
