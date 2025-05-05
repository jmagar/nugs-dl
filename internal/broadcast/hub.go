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
		broadcast:  make(chan []byte),
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
			for client := range h.clients {
				// Use non-blocking send
				select {
				case client <- message:
				default:
					fmt.Println("[Hub Warning] Client channel full, closing client.")
					// Assume client is slow or disconnected, remove it
					close(client)             // Signal closure
					delete(h.clients, client) // Need write lock for this
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastInterface is deprecated, use specific methods below
// func (h *Hub) BroadcastInterface(data interface{}) { ... }

// BroadcastProgressUpdate wraps the update in an SSEEvent and broadcasts it.
func (h *Hub) BroadcastProgressUpdate(update api.ProgressUpdate) {
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

// RegisterClient adds a new client channel to the hub.
func (h *Hub) RegisterClient(clientChan chan []byte) {
	h.register <- clientChan
}

// UnregisterClient removes a client channel from the hub.
func (h *Hub) UnregisterClient(clientChan chan []byte) {
	h.unregister <- clientChan
}
