package availability

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sajadHazrati2000/kei/backend/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// HandleGetSlots — GET /api/v1/availability/{user_id}?from=&to=
func (h *Handler) HandleGetSlots(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("user_id")

	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	slots, err := h.svc.GetSlots(r.Context(), claims.UserID, claims.Role, targetID, claims.OrgID, from, to)
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "from must be before to", "INVALID_RANGE")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get slots", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": slots, "total": len(slots)})
}

// HandleSetSlots — PUT /api/v1/availability/{user_id}
func (h *Handler) HandleSetSlots(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("user_id")

	var req SetSlotsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	slots, err := h.svc.SetSlots(r.Context(), claims.UserID, claims.Role, targetID, req)
	if errors.Is(err, ErrForbidden) {
		writeError(w, http.StatusForbidden, "you can only edit your own availability", "FORBIDDEN")
		return
	}
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "invalid slot data — check status values and time ordering", "INVALID_INPUT")
		return
	}
	if errors.Is(err, ErrConflict) {
		writeError(w, http.StatusConflict, "slot conflicts with an existing busy/focus slot", "SLOT_CONFLICT")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set slots", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": slots, "total": len(slots)})
}

// HandleGetTemplates — GET /api/v1/availability/{user_id}/recurring
func (h *Handler) HandleGetTemplates(w http.ResponseWriter, r *http.Request) {
	targetID := r.PathValue("user_id")

	templates, err := h.svc.GetTemplates(r.Context(), targetID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get templates", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": templates, "total": len(templates)})
}

// HandleSetTemplates — PUT /api/v1/availability/{user_id}/recurring
func (h *Handler) HandleSetTemplates(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("user_id")

	var req SetRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	templates, err := h.svc.SetTemplates(r.Context(), claims.UserID, claims.Role, targetID, req)
	if errors.Is(err, ErrForbidden) {
		writeError(w, http.StatusForbidden, "you can only edit your own templates", "FORBIDDEN")
		return
	}
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "invalid template data", "INVALID_INPUT")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to set templates", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": templates, "total": len(templates)})
}

// HandleImport — POST /api/v1/availability/{user_id}/import
// Expects multipart or raw CSV body with Content-Type: text/csv.
func (h *Handler) HandleImport(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	targetID := r.PathValue("user_id")

	count, err := h.svc.ImportCSV(r.Context(), claims.UserID, claims.Role, targetID, r.Body)
	if errors.Is(err, ErrForbidden) {
		writeError(w, http.StatusForbidden, "you can only import your own availability", "FORBIDDEN")
		return
	}
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_CSV")
		return
	}
	if errors.Is(err, ErrConflict) {
		writeError(w, http.StatusConflict, "CSV contains conflicting busy/focus slots", "SLOT_CONFLICT")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "import failed", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"imported": count})
}

// HandleTeamAvailability — GET /api/v1/team/availability?from=&to=
func (h *Handler) HandleTeamAvailability(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())

	from, to, ok := parseDateRange(w, r)
	if !ok {
		return
	}

	result, err := h.svc.GetTeamAvailability(r.Context(), claims.OrgID, from, to)
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "from must be before to", "INVALID_RANGE")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get team availability", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// HandleTeamOverlap — GET /api/v1/team/overlap?date=YYYY-MM-DD
func (h *Handler) HandleTeamOverlap(w http.ResponseWriter, r *http.Request) {
	claims, _ := middleware.GetClaims(r.Context())
	dateStr := r.URL.Query().Get("date")

	result, err := h.svc.GetTeamOverlap(r.Context(), claims.OrgID, dateStr)
	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "invalid date — use YYYY-MM-DD", "INVALID_DATE")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get overlap", "INTERNAL_ERROR")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// parseDateRange reads ?from= and ?to= as ISO 8601 UTC timestamps.
func parseDateRange(w http.ResponseWriter, r *http.Request) (from, to time.Time, ok bool) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "from and to query params are required (ISO 8601 UTC)", "MISSING_PARAMS")
		return time.Time{}, time.Time{}, false
	}
	var err error
	from, err = time.Parse(time.RFC3339, fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from — use ISO 8601 UTC e.g. 2026-06-19T00:00:00Z", "INVALID_FROM")
		return time.Time{}, time.Time{}, false
	}
	to, err = time.Parse(time.RFC3339, toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to — use ISO 8601 UTC e.g. 2026-06-26T00:00:00Z", "INVALID_TO")
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}
