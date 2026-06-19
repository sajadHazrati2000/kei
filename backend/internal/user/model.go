package user

import (
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrForbidden  = errors.New("forbidden")
	ErrEmailTaken = errors.New("email already in use")
	ErrLastAdmin  = errors.New("cannot deactivate the only admin")
)

type User struct {
	ID           string    `json:"id"`
	OrgID        string    `json:"org_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Timezone     string    `json:"timezone"`
	Language     string    `json:"language"`
	CalendarPref string    `json:"calendar_pref"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type InviteRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateProfileRequest struct {
	Name         string `json:"name"`
	Timezone     string `json:"timezone"`
	Language     string `json:"language"`
	CalendarPref string `json:"calendar_pref"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}
