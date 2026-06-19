package realtime

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// SnapshotFunc fetches the current team availability for an org and returns it
// as a pre-serialised JSON message ready to send to a new client.
type SnapshotFunc func(ctx context.Context, orgID string) ([]byte, error)

// WSHandler handles the WebSocket upgrade at GET /ws/availability.
// Authentication uses ?token= (a valid access JWT) because browsers cannot set
// Authorization headers or custom cookies during the WS handshake.
type WSHandler struct {
	registry      *HubRegistry
	jwtSecret     string
	allowedOrigin string
	snapshot      SnapshotFunc
}

func NewWSHandler(registry *HubRegistry, jwtSecret, allowedOrigin string, snapshot SnapshotFunc) *WSHandler {
	return &WSHandler{
		registry:      registry,
		jwtSecret:     jwtSecret,
		allowedOrigin: allowedOrigin,
		snapshot:      snapshot,
	}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, `{"error":"missing token","code":"UNAUTHORIZED"}`, http.StatusUnauthorized)
		return
	}

	userID, orgID, err := h.validateToken(tokenStr)
	if err != nil {
		http.Error(w, `{"error":"invalid token","code":"UNAUTHORIZED"}`, http.StatusUnauthorized)
		return
	}

	up := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade writes its own error response; just log.
		log.Printf("ws: upgrade failed for user %s: %v", userID, err)
		return
	}

	hub := h.registry.GetOrCreate(orgID)
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, sendBuf),
		UserID: userID,
		OrgID:  orgID,
	}

	hub.register <- client

	// Push the current team snapshot to the newly connected client.
	go h.pushSnapshot(client)

	// readPump and writePump run concurrently; writePump owns all writes.
	go client.writePump()
	go client.readPump()
}

func (h *WSHandler) pushSnapshot(c *Client) {
	data, err := h.snapshot(context.Background(), c.OrgID)
	if err != nil {
		log.Printf("ws: snapshot error for org %s: %v", c.OrgID, err)
		return
	}
	select {
	case c.send <- data:
	default:
		log.Printf("ws: send buffer full during snapshot for user %s", c.UserID)
	}
}

func (h *WSHandler) validateToken(tokenStr string) (userID, orgID string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	userID, _ = claims["user_id"].(string)
	orgID, _ = claims["org_id"].(string)
	if userID == "" || orgID == "" {
		return "", "", fmt.Errorf("missing claims")
	}
	return userID, orgID, nil
}

func (h *WSHandler) checkOrigin(r *http.Request) bool {
	if h.allowedOrigin == "" || h.allowedOrigin == "*" {
		return true
	}
	origin := r.Header.Get("Origin")
	return origin == "" || origin == h.allowedOrigin
}
