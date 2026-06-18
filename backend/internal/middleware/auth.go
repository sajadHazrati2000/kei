package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ClaimsKey contextKey = "auth_claims"

type Claims struct {
	UserID string
	OrgID  string
	Role   string
}

// RequireAuth validates the JWT from the Authorization header or access_token cookie.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := tokenFromRequest(r)
			if tokenStr == "" {
				writeUnauthorized(w)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				writeUnauthorized(w)
				return
			}

			mapClaims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeUnauthorized(w)
				return
			}

			userID, _ := mapClaims["user_id"].(string)
			orgID, _ := mapClaims["org_id"].(string)
			role, _ := mapClaims["role"].(string)
			if userID == "" || orgID == "" {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, Claims{
				UserID: userID,
				OrgID:  orgID,
				Role:   role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole wraps RequireAuth routes and additionally enforces a set of allowed roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok || !allowed[claims.Role] {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"forbidden","code":"FORBIDDEN"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetClaims(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(ClaimsKey).(Claims)
	return c, ok
}

func tokenFromRequest(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := r.Cookie("access_token"); err == nil {
		return cookie.Value
	}
	return ""
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	http.Error(w, `{"error":"unauthorized","code":"UNAUTHORIZED"}`, http.StatusUnauthorized)
}
