// Package federation provides server federation and matchmaking.
package federation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Region represents a geographical region for server filtering.
type Region string

const (
	RegionUSEast  Region = "us-east"
	RegionUSWest  Region = "us-west"
	RegionEUWest  Region = "eu-west"
	RegionEUEast  Region = "eu-east"
	RegionAsiaPac Region = "asia-pac"
	RegionSouthAm Region = "south-am"
	RegionUnknown Region = "unknown"
)

// ServerAnnouncement is sent from game servers to the federation hub.
type ServerAnnouncement struct {
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Region     Region    `json:"region"`
	Genre      string    `json:"genre"`
	Players    int       `json:"players"`
	MaxPlayers int       `json:"maxPlayers"`
	PlayerList []string  `json:"playerList,omitempty"` // List of player IDs currently on this server
	Timestamp  time.Time `json:"timestamp"`
}

// ServerQuery specifies filtering criteria for server discovery.
type ServerQuery struct {
	Region     *Region `json:"region,omitempty"`
	Genre      *string `json:"genre,omitempty"`
	MinPlayers *int    `json:"minPlayers,omitempty"`
	MaxPlayers *int    `json:"maxPlayers,omitempty"`
}

// FederationHub manages server announcements and client queries.
type FederationHub struct {
	servers         map[string]*ServerAnnouncement
	playerIndex     map[string]string // playerID -> serverName
	mu              sync.RWMutex
	upgrader        websocket.Upgrader
	staleTimeout    time.Duration
	cleanupInterval time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	httpServer      *http.Server
}

// NewFederationHub creates a new federation hub.
func NewFederationHub() *FederationHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &FederationHub{
		servers:         make(map[string]*ServerAnnouncement),
		playerIndex:     make(map[string]string),
		upgrader:        websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		staleTimeout:    30 * time.Second,
		cleanupInterval: 10 * time.Second,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start begins the federation hub HTTP server.
func (h *FederationHub) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/announce", h.handleAnnounce)
	mux.HandleFunc("/query", h.handleQuery)
	mux.HandleFunc("/lookup", h.handleLookup)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	h.httpServer = &http.Server{
		Addr:    listener.Addr().String(),
		Handler: mux,
	}

	go h.cleanupStaleServers()

	go func() {
		if err := h.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("federation hub server error")
		}
	}()

	return nil
}

// GetAddr returns the address the hub is listening on.
func (h *FederationHub) GetAddr() string {
	if h.httpServer == nil {
		return ""
	}
	return h.httpServer.Addr
}

// Stop gracefully shuts down the federation hub.
func (h *FederationHub) Stop() error {
	h.cancel()
	if h.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return h.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleAnnounce processes WebSocket connections from game servers.
func (h *FederationHub) handleAnnounce(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("failed to upgrade websocket")
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Debug("websocket read error")
			return
		}

		var announcement ServerAnnouncement
		if err := json.Unmarshal(msg, &announcement); err != nil {
			logrus.WithError(err).Error("failed to parse announcement")
			continue
		}

		announcement.Timestamp = time.Now()
		h.registerServer(&announcement)

		logrus.WithFields(logrus.Fields{
			"server_name": announcement.Name,
			"region":      announcement.Region,
			"genre":       announcement.Genre,
			"players":     fmt.Sprintf("%d/%d", announcement.Players, announcement.MaxPlayers),
		}).Debug("server announced")
	}
}

// handleQuery processes REST queries from clients.
func (h *FederationHub) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query ServerQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "invalid query", http.StatusBadRequest)
		return
	}

	servers := h.queryServers(&query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// PlayerLookupRequest represents a player lookup query.
type PlayerLookupRequest struct {
	PlayerID string `json:"playerID"`
}

// PlayerLookupResponse contains the server address if player is online.
type PlayerLookupResponse struct {
	Online        bool   `json:"online"`
	ServerAddress string `json:"serverAddress,omitempty"`
	ServerName    string `json:"serverName,omitempty"`
}

// handleLookup processes player presence lookup requests.
func (h *FederationHub) handleLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlayerLookupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.PlayerID == "" {
		http.Error(w, "playerID is required", http.StatusBadRequest)
		return
	}

	response := h.lookupPlayer(req.PlayerID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// lookupPlayer finds a player across all federated servers.
func (h *FederationHub) lookupPlayer(playerID string) PlayerLookupResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	serverName, found := h.playerIndex[playerID]
	if !found {
		return PlayerLookupResponse{Online: false}
	}

	server, exists := h.servers[serverName]
	if !exists {
		return PlayerLookupResponse{Online: false}
	}

	return PlayerLookupResponse{
		Online:        true,
		ServerAddress: server.Address,
		ServerName:    server.Name,
	}
}

// registerServer adds or updates a server announcement.
func (h *FederationHub) registerServer(announcement *ServerAnnouncement) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove old player mappings for this server
	if oldAnnouncement, exists := h.servers[announcement.Name]; exists {
		for _, playerID := range oldAnnouncement.PlayerList {
			delete(h.playerIndex, playerID)
		}
	}

	// Add new player mappings
	for _, playerID := range announcement.PlayerList {
		h.playerIndex[playerID] = announcement.Name
	}

	h.servers[announcement.Name] = announcement
}

// RegisterServer adds or updates a server announcement (public API).
func (h *FederationHub) RegisterServer(announcement *ServerAnnouncement) {
	h.registerServer(announcement)
}

// queryServers filters servers based on query criteria.
func (h *FederationHub) queryServers(query *ServerQuery) []*ServerAnnouncement {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []*ServerAnnouncement
	for _, server := range h.servers {
		if h.matchesQuery(server, query) {
			results = append(results, server)
		}
	}
	return results
}

// matchesQuery checks if a server matches query criteria.
func (h *FederationHub) matchesQuery(server *ServerAnnouncement, query *ServerQuery) bool {
	if query.Region != nil && server.Region != *query.Region {
		return false
	}
	if query.Genre != nil && server.Genre != *query.Genre {
		return false
	}
	if query.MinPlayers != nil && server.Players < *query.MinPlayers {
		return false
	}
	if query.MaxPlayers != nil && server.Players > *query.MaxPlayers {
		return false
	}
	return true
}

// cleanupStaleServers removes servers that haven't announced recently.
func (h *FederationHub) cleanupStaleServers() {
	ticker := time.NewTicker(h.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.mu.Lock()
			now := time.Now()
			for name, server := range h.servers {
				if now.Sub(server.Timestamp) > h.staleTimeout {
					// Remove player mappings
					for _, playerID := range server.PlayerList {
						delete(h.playerIndex, playerID)
					}
					delete(h.servers, name)
					logrus.WithField("server_name", name).Debug("removed stale server")
				}
			}
			h.mu.Unlock()
		}
	}
}

// GetServerCount returns the number of registered servers.
func (h *FederationHub) GetServerCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.servers)
}

// QueryServers filters servers based on query criteria (public API).
func (h *FederationHub) QueryServers(query *ServerQuery) []*ServerAnnouncement {
	return h.queryServers(query)
}

// SetCleanupInterval updates the cleanup interval (primarily for testing).
func (h *FederationHub) SetCleanupInterval(interval time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanupInterval = interval
}

// SetStaleTimeout updates the stale timeout (primarily for testing).
func (h *FederationHub) SetStaleTimeout(timeout time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.staleTimeout = timeout
}

// ServerAnnouncer sends periodic announcements to the federation hub.
type ServerAnnouncer struct {
	hubURL       string
	conn         *websocket.Conn
	announcement ServerAnnouncement
	interval     time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
}

// NewServerAnnouncer creates a new server announcer.
func NewServerAnnouncer(hubURL string, announcement ServerAnnouncement) *ServerAnnouncer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServerAnnouncer{
		hubURL:       hubURL,
		announcement: announcement,
		interval:     10 * time.Second,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// SetInterval sets the announcement interval (primarily for testing).
func (a *ServerAnnouncer) SetInterval(interval time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.interval = interval
}

// Start begins announcing to the federation hub.
func (a *ServerAnnouncer) Start() error {
	conn, _, err := websocket.DefaultDialer.Dial(a.hubURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to federation hub: %w", err)
	}
	a.conn = conn

	go a.announceLoop()
	return nil
}

// Stop halts the announcer.
func (a *ServerAnnouncer) Stop() error {
	a.cancel()
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// UpdatePlayers updates the player count in announcements.
func (a *ServerAnnouncer) UpdatePlayers(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.announcement.Players = count
}

// UpdatePlayerList updates the list of players on this server.
func (a *ServerAnnouncer) UpdatePlayerList(playerIDs []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.announcement.PlayerList = playerIDs
	a.announcement.Players = len(playerIDs)
}

// announceLoop sends periodic announcements.
func (a *ServerAnnouncer) announceLoop() {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	// Send initial announcement
	a.sendAnnouncement()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.sendAnnouncement()
		}
	}
}

// sendAnnouncement sends a single announcement.
func (a *ServerAnnouncer) sendAnnouncement() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn == nil {
		return
	}

	data, err := json.Marshal(a.announcement)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal announcement")
		return
	}

	if err := a.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		logrus.WithError(err).Error("failed to send announcement")
	}
}

// DiscoverServers queries a remote federation hub for available servers.
func DiscoverServers(hubURL string, query *ServerQuery, timeout time.Duration) ([]ServerAnnouncement, error) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	data, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	url := hubURL + "/query"
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to query federation hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("federation hub returned status %d", resp.StatusCode)
	}

	var results []ServerAnnouncement
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// LookupPlayer queries a remote federation hub for player presence.
func LookupPlayer(hubURL, playerID string, timeout time.Duration) (*PlayerLookupResponse, error) {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	req := PlayerLookupRequest{PlayerID: playerID}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := hubURL + "/lookup"
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to lookup player: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("federation hub returned status %d", resp.StatusCode)
	}

	var result PlayerLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
