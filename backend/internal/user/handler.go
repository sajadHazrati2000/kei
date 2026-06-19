package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sajadHazrati2000/kei/backend/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleList — GET /api/v1/users
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	users, err := h.svc.List(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users, "total": len(users)})
}

// HandleInvite — POST /api/v1/users/invite (admin only)
func (h *Handler) HandleInvite(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	var req InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	if req.Name == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "name and email are required", "MISSING_FIELDS")
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email address", "INVALID_EMAIL")
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		writeError(w, http.StatusBadRequest, "role must be admin, member, or viewer", "INVALID_ROLE")
		return
	}

	u, tempPassword, err := h.svc.Invite(r.Context(), claims.OrgID, req)
	if errors.Is(err, ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email already in use", "EMAIL_TAKEN")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to invite user", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": u, "temp_password": tempPassword})
}

// HandleGet — GET /api/v1/users/{id}
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("id")

	u, err := h.svc.GetByID(r.Context(), claims.OrgID, targetID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": u})
}

// HandleUpdateProfile — PUT /api/v1/users/{id} (own profile only)
func (h *Handler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("id")

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Timezone = strings.TrimSpace(req.Timezone)

	if req.Name == "" || req.Timezone == "" {
		writeError(w, http.StatusBadRequest, "name and timezone are required", "MISSING_FIELDS")
		return
	}
	if req.Language != "en" && req.Language != "fa" {
		writeError(w, http.StatusBadRequest, "language must be en or fa", "INVALID_LANGUAGE")
		return
	}
	if req.CalendarPref != "gregorian" && req.CalendarPref != "jalali" {
		writeError(w, http.StatusBadRequest, "calendar_pref must be gregorian or jalali", "INVALID_CALENDAR_PREF")
		return
	}

	u, err := h.svc.UpdateProfile(r.Context(), claims.UserID, targetID, claims.OrgID, req)
	if errors.Is(err, ErrForbidden) {
		writeError(w, http.StatusForbidden, "you can only update your own profile", "FORBIDDEN")
		return
	}
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": u})
}

// HandleUpdateRole — PUT /api/v1/users/{id}/role (admin only)
func (h *Handler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("id")

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		writeError(w, http.StatusBadRequest, "role must be admin, member, or viewer", "INVALID_ROLE")
		return
	}

	u, err := h.svc.UpdateRole(r.Context(), claims.OrgID, targetID, req.Role)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update role", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": u})
}

// HandleDeactivate — DELETE /api/v1/users/{id} (admin only)
func (h *Handler) HandleDeactivate(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("id")

	err := h.svc.Deactivate(r.Context(), claims.OrgID, claims.UserID, targetID)
	if errors.Is(err, ErrLastAdmin) {
		writeError(w, http.StatusConflict, "cannot deactivate the only admin", "LAST_ADMIN")
		return
	}
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to deactivate user", "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}
