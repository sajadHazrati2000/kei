package realtime

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 54 * time.Second // must be < pongWait
	maxMsgSize = 512
	sendBuf    = 256
)

// Client is a single WebSocket connection registered to a Hub.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	UserID string
	OrgID  string
}

// readPump pumps messages from the WebSocket to /dev/null.
// Clients don't send data in Phase 1 — this goroutine exists only to detect
// disconnection and handle pong frames so the connection stays alive.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: read error for user %s: %v", c.UserID, err)
			}
			break
		}
	}
}

// writePump pumps messages from the send channel to the WebSocket.
// One writePump per client — the only goroutine that writes to the connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub maintains the set of active clients for one org and broadcasts messages
// to all of them. The clients map is owned exclusively by the run() goroutine
// — no mutex needed on the map itself.
type Hub struct {
	orgID      string
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub(orgID string) *Hub {
	return &Hub{
		orgID:      orgID,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 64),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("ws: client connected org=%s user=%s total=%d", h.orgID, client.UserID, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("ws: client disconnected org=%s user=%s total=%d", h.orgID, client.UserID, len(h.clients))
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Send buffer full — drop slow client
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast sends a pre-serialised JSON message to all clients in this hub.
// Non-blocking: drops the message if the broadcast channel is full.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
		log.Printf("ws: broadcast channel full for org=%s, message dropped", h.orgID)
	}
}

// HubRegistry holds one Hub per org, creating them lazily.
type HubRegistry struct {
	mu   sync.RWMutex
	hubs map[string]*Hub
}

func NewHubRegistry() *HubRegistry {
	return &HubRegistry{hubs: make(map[string]*Hub)}
}

// GetOrCreate returns the Hub for orgID, creating and starting it if needed.
func (r *HubRegistry) GetOrCreate(orgID string) *Hub {
	r.mu.RLock()
	h, ok := r.hubs[orgID]
	r.mu.RUnlock()
	if ok {
		return h
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok = r.hubs[orgID]; ok { // re-check under write lock
		return h
	}
	h = newHub(orgID)
	r.hubs[orgID] = h
	go h.run()
	return h
}

// Broadcast serialises delivery to the org's hub without the caller needing
// to look up the hub themselves. Safe to call even if no clients are connected.
func (r *HubRegistry) Broadcast(orgID string, msg []byte) {
	r.mu.RLock()
	h, ok := r.hubs[orgID]
	r.mu.RUnlock()
	if ok {
		h.Broadcast(msg)
	}
}
