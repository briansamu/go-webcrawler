// internal/api/server.go
package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"webcrawler/internal/models"
	"webcrawler/internal/queue"
	"webcrawler/internal/stats"
	"webcrawler/internal/storage"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type APIServer struct {
	storage       *storage.MongoDB
	stats         *stats.CrawlerStats
	crawledSet    *queue.CrawledSet
	queue         *queue.Queue
	upgrader      websocket.Upgrader
	wsConnections map[*websocket.Conn]bool
}

type StatsResponse struct {
	TotalCrawled    int     `json:"totalCrawled"`
	TotalQueued     int     `json:"totalQueued"`
	QueueSize       int     `json:"queueSize"`
	CrawlRate       float64 `json:"crawlRate"`
	CrawledToQueued float64 `json:"crawledToQueued"`
	UptimeMinutes   float64 `json:"uptimeMinutes"`
	Status          string  `json:"status"`
}

type SearchResponse struct {
	Pages       []models.Page `json:"pages"`
	TotalCount  int           `json:"totalCount"`
	CurrentPage int           `json:"currentPage"`
	TotalPages  int           `json:"totalPages"`
}

func NewAPIServer(storage *storage.MongoDB, stats *stats.CrawlerStats, crawled *queue.CrawledSet, queue *queue.Queue) *APIServer {
	return &APIServer{
		storage:       storage,
		stats:         stats,
		crawledSet:    crawled,
		queue:         queue,
		upgrader:      websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		wsConnections: make(map[*websocket.Conn]bool),
	}
}

func (s *APIServer) Start(port string) {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/stats", s.handleStats).Methods("GET")
	r.HandleFunc("/api/search", s.handleSearch).Methods("GET")
	r.HandleFunc("/api/pages", s.handlePages).Methods("GET")

	// WebSocket for live updates
	r.HandleFunc("/ws/stats", s.handleWebSocket)

	// Static files - try multiple locations
	staticPaths := []string{"./web/static/", "../web/static/", "../../web/static/"}
	var staticDir string

	for _, path := range staticPaths {
		if _, err := http.Dir(path).Open("index.html"); err == nil {
			staticDir = path
			break
		}
	}

	if staticDir == "" {
		staticDir = "./web/static/" // fallback
	}

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(staticDir)))

	fmt.Printf("API Server starting on port %s\n", port)
	fmt.Printf("Static files serving from: %s\n", staticDir)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func (s *APIServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.getCurrentStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *APIServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	results := s.searchPages(query, page, 10)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *APIServer) handlePages(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	results := s.getPages(page, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *APIServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	s.wsConnections[conn] = true
	defer delete(s.wsConnections, conn)

	// Send initial stats
	stats := s.getCurrentStats()
	conn.WriteJSON(stats)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *APIServer) BroadcastStats() {
	stats := s.getCurrentStats()

	for conn := range s.wsConnections {
		err := conn.WriteJSON(stats)
		if err != nil {
			conn.Close()
			delete(s.wsConnections, conn)
		}
	}
}

func (s *APIServer) getCurrentStats() StatsResponse {
	return StatsResponse{
		TotalCrawled:    s.crawledSet.Size(),
		TotalQueued:     s.queue.TotalQueued(),
		QueueSize:       s.queue.Size(),
		CrawlRate:       float64(s.crawledSet.Size()) / time.Since(s.stats.GetStartTime()).Minutes(),
		CrawledToQueued: float64(s.crawledSet.Size()) / float64(s.queue.TotalQueued()),
		UptimeMinutes:   time.Since(s.stats.GetStartTime()).Minutes(),
		Status:          "running",
	}
}

func (s *APIServer) searchPages(query string, page, limit int) SearchResponse {
	pages, total, err := s.storage.SearchPages(query, page, limit)
	if err != nil {
		log.Printf("Error searching pages: %v", err)
		return SearchResponse{Pages: []models.Page{}, TotalCount: 0, CurrentPage: page, TotalPages: 0}
	}

	totalPages := (total + limit - 1) / limit
	return SearchResponse{
		Pages:       pages,
		TotalCount:  total,
		CurrentPage: page,
		TotalPages:  totalPages,
	}
}

func (s *APIServer) getPages(page, limit int) SearchResponse {
	pages, total, err := s.storage.GetPages(page, limit)
	if err != nil {
		log.Printf("Error getting pages: %v", err)
		return SearchResponse{Pages: []models.Page{}, TotalCount: 0, CurrentPage: page, TotalPages: 0}
	}

	totalPages := (total + limit - 1) / limit
	return SearchResponse{
		Pages:       pages,
		TotalCount:  total,
		CurrentPage: page,
		TotalPages:  totalPages,
	}
}
