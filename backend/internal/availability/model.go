package availability

import (
	"errors"
	"time"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("slot conflicts with an existing busy/focus slot")
	ErrInvalidInput = errors.New("invalid input")
)

var validStatuses = map[string]bool{"free": true, "busy": true, "focus": true}
var validPatterns = map[string]bool{"daily": true, "weekly": true}

// Slot is a concrete availability block persisted in availability_slots.
type Slot struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	Status       string    `json:"status"`
	IsOverride   bool      `json:"is_override"`
	RecurrenceID *string   `json:"recurrence_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// SlotInput is one slot in a PUT or import request.
type SlotInput struct {
	From   time.Time `json:"from"`
	To     time.Time `json:"to"`
	Status string    `json:"status"`
}

// SetSlotsRequest replaces all slots in [From,To) with the given list.
type SetSlotsRequest struct {
	From  time.Time   `json:"from"`
	To    time.Time   `json:"to"`
	Slots []SlotInput `json:"slots"`
}

// RecurringTemplate is a row from recurrence_templates.
type RecurringTemplate struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	Pattern    string  `json:"pattern"`
	DaysOfWeek []int   `json:"days_of_week"`
	StartTime  string  `json:"start_time"` // "HH:MM" UTC
	EndTime    string  `json:"end_time"`   // "HH:MM" UTC
	Status     string  `json:"status"`
	ValidFrom  string  `json:"valid_from"`            // "YYYY-MM-DD"
	ValidUntil *string `json:"valid_until,omitempty"` // "YYYY-MM-DD"
}

// RecurringTemplateInput is one template in a PUT request.
type RecurringTemplateInput struct {
	Pattern    string  `json:"pattern"`
	DaysOfWeek []int   `json:"days_of_week"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Status     string  `json:"status"`
	ValidFrom  string  `json:"valid_from"`
	ValidUntil *string `json:"valid_until,omitempty"`
}

type SetRecurringRequest struct {
	Templates []RecurringTemplateInput `json:"templates"`
}

// UserSummary is the subset of user fields embedded in team responses.
type UserSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
	Role     string `json:"role"`
}

// UserSlots pairs a user summary with their availability slots.
type UserSlots struct {
	User  UserSummary `json:"user"`
	Slots []*Slot     `json:"slots"`
}

// OverlapResult is the response from GET /team/overlap.
type OverlapResult struct {
	Date         string      `json:"date"`
	OverlapStart string      `json:"overlap_start"`
	OverlapEnd   string      `json:"overlap_end"`
	Members      []UserSlots `json:"members"`
}
