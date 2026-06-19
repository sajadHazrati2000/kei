package settings

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("blocked day already exists")
)

// OrgSettings is the full settings payload for GET and PUT responses.
type OrgSettings struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Timezone        string `json:"timezone"`
	OverlapStart    string `json:"overlap_start"`    // "HH:MM"
	OverlapEnd      string `json:"overlap_end"`      // "HH:MM"
	DefaultLanguage string `json:"default_language"` // "en" | "fa"
	WorkingDays     []int  `json:"working_days"`     // 0=Sun … 6=Sat
	WorkingStart    string `json:"working_start"`    // "HH:MM"
	WorkingEnd      string `json:"working_end"`      // "HH:MM"
}

// UpdateSettingsRequest is the body for PUT /api/v1/settings.
// Slug is not updatable — it is part of the org identity.
type UpdateSettingsRequest struct {
	Name            string `json:"name"`
	Timezone        string `json:"timezone"`
	OverlapStart    string `json:"overlap_start"`
	OverlapEnd      string `json:"overlap_end"`
	DefaultLanguage string `json:"default_language"`
	WorkingDays     []int  `json:"working_days"`
	WorkingStart    string `json:"working_start"`
	WorkingEnd      string `json:"working_end"`
}

// BlockedDay is a row from the blocked_days table.
type BlockedDay struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	BlockedDate string    `json:"blocked_date"` // "YYYY-MM-DD"
	Reason      *string   `json:"reason,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// AddBlockedDayRequest is the body for POST /api/v1/settings/blocked-days.
type AddBlockedDayRequest struct {
	BlockedDate string  `json:"blocked_date"` // "YYYY-MM-DD"
	Reason      *string `json:"reason,omitempty"`
}
