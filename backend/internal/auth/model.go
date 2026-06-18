package auth

import "time"

type SetupRequest struct {
	OrgName   string `json:"org_name"`
	OrgSlug   string `json:"org_slug"`
	AdminName string `json:"admin_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Timezone  string `json:"timezone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID           string
	OrgID        string
	Name         string
	Email        string
	PasswordHash string
	Role         string
	Timezone     string
	Language     string
	CalendarPref string
	IsActive     bool
}

type UserInfo struct {
	ID    string `json:"id"`
	OrgID string `json:"org_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	UserInfo     UserInfo
}

type RefreshToken struct {
	JTI       string
	UserID    string
	ExpiresAt time.Time
}
