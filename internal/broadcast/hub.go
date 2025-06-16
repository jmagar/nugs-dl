package broadcast

import (
	"encoding/json"
	"fmt"                     // Needed for Sprintf
	"nugs-dl/internal/logger" // Import the logger package
	"nugs-dl/pkg/api"         // Updated path
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients. Using map with bool value for easy addition/deletion.
	clients map[chan []byte]bool

	// Inbound messages from the progress consumer.
	broadcast chan []byte // Channel for byte slices (marshalled JSON)

	// Register requests from the clients.
	register chan chan []byte

	// Unregister requests from clients.
	unregister chan chan []byte

	mutex sync.RWMutex // To protect the clients map
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 1000), // Large buffer for high-frequency progress updates
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		clients:    make(map[chan []byte]bool),
	}
}

// Run starts the hub's processing loop.
func (h *Hub) Run() {
	logger.Info("[Hub] Starting broadcaster...")
	for {
		select {
		case client := <-h.register:
			logger.Info("[Hub] Client registered", "client", fmt.Sprintf("%p", client)) // Log client address for tracking
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				logger.Info("[Hub] Client unregistered", "client", fmt.Sprintf("%p", client))
				delete(h.clients, client)
				close(client) // Close the client channel
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			// logger.Debug("[Hub] Broadcasting message...") // Potentially too verbose
			h.mutex.RLock()
			slowClients := make([]chan []byte, 0) // Collect slow clients for cleanup
			for client := range h.clients {
				// Use non-blocking send
				select {
				case client <- message:
					// Successfully sent
				default:
					// Client channel is full - mark for potential cleanup
					logger.Warn("[Hub] Client channel full, marking slow client.", "client", fmt.Sprintf("%p", client))
					slowClients = append(slowClients, client)
				}
			}
			h.mutex.RUnlock()

			// Clean up slow clients (only if there are many to avoid closing responsive clients temporarily)
			if len(slowClients) > 0 && len(slowClients) < len(h.clients) { // Avoid removing all clients if broadcast channel was overwhelmed
				h.mutex.Lock()
				for _, client := range slowClients {
					// Double-check client still exists and is still full before removing
					select {
					case client <- message:
						// Client recovered, don't remove
						logger.Debug("[Hub] Slow client recovered, not removing.", "client", fmt.Sprintf("%p", client))
					default:
						// Client still full, remove it
						if _, exists := h.clients[client]; exists {
							logger.Warn("[Hub] Removing persistently slow client.", "client", fmt.Sprintf("%p", client))
							delete(h.clients, client)
							close(client)
						}
					}
				}
				h.mutex.Unlock()
			}
		}
	}
}

// BroadcastInterface is deprecated, use specific methods below
// func (h *Hub) BroadcastInterface(data interface{}) { ... }

// BroadcastProgressUpdate wraps the update in an SSEEvent and broadcasts it.
func (h *Hub) BroadcastProgressUpdate(update api.ProgressUpdate) {
	logger.Debug("[Hub] Broadcasting progress update", 
		"jobID", update.JobID, 
		"percentage", update.Percentage, 
		"speedBPS", update.SpeedBPS,
		"currentFile", update.CurrentFile,
		"currentTrack", update.CurrentTrack,
		"totalTracks", update.TotalTracks,
	)
		
	event := api.SSEEvent{
		Type: api.SSEProgressUpdate,
		Data: update,
	}
	messageBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("[Hub] Failed to marshal progress update event", "error", err, "jobID", update.JobID)
		return
	}
	select {
	case h.broadcast <- messageBytes:
		logger.Debug("[Hub] Progress update queued for broadcast", "jobID", update.JobID)
	default:
		logger.Warn("[Hub] Broadcast channel full, discarding progress update.", "jobID", update.JobID)
	}
}

// BroadcastJobAdded wraps the job in an SSEEvent and broadcasts it.
func (h *Hub) BroadcastJobAdded(job *api.DownloadJob) {
	event := api.SSEEvent{
		Type: api.SSEJobAdded,
		Data: job,
	}
	messageBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("[Hub] Failed to marshal job added event", "error", err, "jobID", job.ID)
		return
	}
	select {
	case h.broadcast <- messageBytes:
		logger.Debug("[Hub] Job added event queued for broadcast", "jobID", job.ID)
	default:
		logger.Warn("[Hub] Broadcast channel full, discarding job added event.", "jobID", job.ID)
	}
}

// BroadcastJobStatusUpdate wraps the job in an SSEEvent and broadcasts it.
func (h *Hub) BroadcastJobStatusUpdate(job *api.DownloadJob) {
	logger.Info("[Hub] Broadcasting job status update", "jobID", job.ID, "status", job.Status)
		
	event := api.SSEEvent{
		Type: api.SSEJobStatusUpdate,
		Data: job,
	}
	messageBytes, err := json.Marshal(event)
	if err != nil {
		logger.Error("[Hub] Failed to marshal job status update event", "error", err, "jobID", job.ID)
		return
	}
	select {
	case h.broadcast <- messageBytes:
		logger.Debug("[Hub] Job status update queued for broadcast", "jobID", job.ID)
	default:
		logger.Warn("[Hub] Broadcast channel full, discarding job status update.", "jobID", job.ID)
	}
}

// RegisterClient adds a new client channel to the hub.
func (h *Hub) RegisterClient(clientChan chan []byte) {
	h.register <- clientChan
}

// UnregisterClient removes a client channel from the hub.
func (h *Hub) UnregisterClient(clientChan chan []byte) {
	h.unregister <- clientChan
}
