package broadcast

import (
	"encoding/json"
	"fmt"
	"nugs-dl/pkg/api" // Updated path
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
	fmt.Println("[Hub] Starting broadcaster...")
	for {
		select {
		case client := <-h.register:
			fmt.Println("[Hub] Client registered")
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				fmt.Println("[Hub] Client unregistered")
				delete(h.clients, client)
				close(client) // Close the client channel
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			// fmt.Println("[Hub] Broadcasting message...")
			h.mutex.RLock()
			slowClients := make([]chan []byte, 0) // Collect slow clients for cleanup
			for client := range h.clients {
				// Use non-blocking send
				select {
				case client <- message:
					// Successfully sent
				default:
					// Client channel is full - mark for potential cleanup
					fmt.Println("[Hub Warning] Client channel full, marking slow client.")
					slowClients = append(slowClients, client)
				}
			}
			h.mutex.RUnlock()
			
			// Clean up slow clients (only if there are many to avoid closing responsive clients temporarily)
			if len(slowClients) > 0 && len(slowClients) < len(h.clients) {
				h.mutex.Lock()
				for _, client := range slowClients {
					// Double-check client still exists and is still full before removing
					select {
					case client <- message:
						// Client recovered, don't remove
					default:
						// Client still full, remove it
						if _, exists := h.clients[client]; exists {
							fmt.Println("[Hub] Removing persistently slow client.")
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
	fmt.Printf("[Hub] Broadcasting progress update: Job %s, Percentage: %.1f%%, Speed: %d B/s\n", 
		update.JobID, update.Percentage, update.SpeedBPS)
		
	event := api.SSEEvent{
		Type: api.SSEProgressUpdate,
		Data: update,
	}
	messageBytes, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("[Hub Error] Failed to marshal progress update event: %v\n", err)
		return
	}
	select {
	case h.broadcast <- messageBytes:
		fmt.Printf("[Hub] Progress update queued for broadcast: Job %s\n", update.JobID)
	default:
		fmt.Println("[Hub Warning] Broadcast channel full, discarding progress update.")
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
		fmt.Printf("[Hub Error] Failed to marshal job added event: %v\n", err)
		return
	}
	select {
	case h.broadcast <- messageBytes:
	default:
		fmt.Println("[Hub Warning] Broadcast channel full, discarding job added event.")
	}
}

// BroadcastJobStatusUpdate wraps the job in an SSEEvent and broadcasts it.
func (h *Hub) BroadcastJobStatusUpdate(job *api.DownloadJob) {
	fmt.Printf("[Hub] Broadcasting job status update: Job %s, Status: %s\n", 
		job.ID, job.Status)
		
	event := api.SSEEvent{
		Type: api.SSEJobStatusUpdate,
		Data: job,
	}
	messageBytes, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("[Hub Error] Failed to marshal job status update event: %v\n", err)
		return
	}
	select {
	case h.broadcast <- messageBytes:
		fmt.Printf("[Hub] Job status update queued for broadcast: Job %s\n", job.ID)
	default:
		fmt.Println("[Hub Warning] Broadcast channel full, discarding job status update.")
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
