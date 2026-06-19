package settings

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

// HandleGetSettings — GET /api/v1/settings
func (h *Handler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	org, err := h.svc.GetSettings(r.Context(), claims.OrgID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "settings not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get settings", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": org})
}

// HandleUpdateSettings — PUT /api/v1/settings (admin only)
func (h *Handler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Timezone = strings.TrimSpace(req.Timezone)

	org, err := h.svc.UpdateSettings(r.Context(), claims.OrgID, req)
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_INPUT")
		return
	}
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "org not found", "NOT_FOUND")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update settings", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": org})
}

// HandleListBlockedDays — GET /api/v1/settings/blocked-days
func (h *Handler) HandleListBlockedDays(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	days, err := h.svc.ListBlockedDays(r.Context(), claims.OrgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list blocked days", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": days, "total": len(days)})
}

// HandleAddBlockedDay — POST /api/v1/settings/blocked-days (admin only)
func (h *Handler) HandleAddBlockedDay(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	var req AddBlockedDayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	req.BlockedDate = strings.TrimSpace(req.BlockedDate)

	day, err := h.svc.AddBlockedDay(r.Context(), claims.OrgID, claims.UserID, req)
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_INPUT")
		return
	}
	if errors.Is(err, ErrAlreadyExists) {
		writeError(w, http.StatusConflict, "a blocked day already exists for that date", "ALREADY_EXISTS")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add blocked day", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": day})
}

// HandleDeleteBlockedDay — DELETE /api/v1/settings/blocked-days/{date} (admin only)
func (h *Handler) HandleDeleteBlockedDay(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	dateStr := r.PathValue("date")

	if err := h.svc.DeleteBlockedDay(r.Context(), claims.OrgID, dateStr); err != nil {
		if errors.Is(err, ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error(), "INVALID_DATE")
			return
		}
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "no blocked day found for that date", "NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete blocked day", "INTERNAL_ERROR")
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
