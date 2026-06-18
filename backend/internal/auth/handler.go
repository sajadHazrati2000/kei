package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Handler struct {
	svc *Service
	env string
}

func NewHandler(svc *Service, env string) *Handler {
	return &Handler{svc: svc, env: env}
}

func (h *Handler) HandleSetup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if req.OrgName == "" || req.OrgSlug == "" || req.AdminName == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "org_name, org_slug, admin_name, email, and password are required", "MISSING_FIELDS")
		return
	}

	resp, err := h.svc.Setup(r.Context(), req)
	if errors.Is(err, ErrSetupDone) {
		writeError(w, http.StatusConflict, "setup already completed", "SETUP_DONE")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "setup failed", "INTERNAL_ERROR")
		return
	}

	h.setTokenCookies(w, resp)
	writeJSON(w, http.StatusCreated, map[string]any{"user": resp.UserInfo})
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	resp, err := h.svc.Login(r.Context(), req)
	if errors.Is(err, ErrInvalidCreds) || errors.Is(err, ErrInactiveUser) {
		writeError(w, http.StatusUnauthorized, "invalid credentials", "INVALID_CREDENTIALS")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "login failed", "INTERNAL_ERROR")
		return
	}

	h.setTokenCookies(w, resp)
	writeJSON(w, http.StatusOK, map[string]any{"user": resp.UserInfo})
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "missing refresh token", "MISSING_TOKEN")
		return
	}

	resp, err := h.svc.Refresh(r.Context(), cookie.Value)
	if errors.Is(err, ErrInvalidToken) {
		h.clearTokenCookies(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired token", "INVALID_TOKEN")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "refresh failed", "INTERNAL_ERROR")
		return
	}

	h.setTokenCookies(w, resp)
	writeJSON(w, http.StatusOK, map[string]any{"user": resp.UserInfo})
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		h.svc.Logout(r.Context(), cookie.Value)
	}
	h.clearTokenCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setTokenCookies(w http.ResponseWriter, resp *TokenResponse) {
	secure := h.env != "development"

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    resp.AccessToken,
		Path:     "/",
		MaxAge:   int((15 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
	// Path scoped to /api/v1/auth so the cookie is sent on both /refresh and /logout
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handler) clearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "access_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", Path: "/api/v1/auth", MaxAge: -1, HttpOnly: true})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}
