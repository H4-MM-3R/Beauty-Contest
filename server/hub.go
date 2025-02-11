package server

import (
	"encoding/json"
	"fmt"
	"time"
)

// PlayerState represents one player’s status.
type PlayerState struct {
	Name     string      `json:"name"`
	Response interface{} `json:"response"` // either an int or the string "still needs to respond"
}

// StateMessage is broadcast to all clients.
type StateMessage struct {
	Type    string        `json:"type"` // "state" (ongoing round) or "result" (round complete)
	Players []PlayerState `json:"players"`
	Average *float64      `json:"average,omitempty"`
}

// Hub maintains the set of active clients and round responses.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Channels for registration/unregistration.
	register   chan *Client
	unregister chan *Client

	// Channel for inbound responses from clients.
	response chan *Response

	// Map from client name to their response (for the current round).
	responses map[string]int

	// When a round is complete, roundLocked is set to true.
	roundLocked bool
}

// Response wraps a client’s input.
type Response struct {
	client *Client
	value  int
}

// newHub creates a new Hub instance.
func newHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		response:    make(chan *Response),
		responses:   make(map[string]int),
		roundLocked: false,
	}
}

// buildState constructs a StateMessage.
// If avg is nil then the round is ongoing; if not, it is the round result.
func (h *Hub) buildState(avg *float64, msgType string) StateMessage {
	players := []PlayerState{}
	for client := range h.clients {
		var resp interface{}
		if val, ok := h.responses[client.name]; ok {
			resp = val
		} else {
			resp = "still needs to respond"
		}
		players = append(players, PlayerState{
			Name:     client.name,
			Response: resp,
		})
	}
	return StateMessage{
		Type:    msgType,
		Players: players,
		Average: avg,
	}
}

// broadcastMessage sends the given JSON message to all clients.
func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// If the client is unresponsive, remove it.
			delete(h.clients, client)
			close(client.send)
		}
	}
}

// broadcastState marshals and broadcasts the current round state.
func (h *Hub) broadcastState() {
	state := h.buildState(nil, "state")
	data, err := json.Marshal(state)
	if err != nil {
		fmt.Println("Error marshaling state:", err)
		return
	}
	h.broadcastMessage(data)
}

// run listens for client registration, unregistration, and responses.
// When all connected clients have responded, it computes the average,
// broadcasts the result, waits 10 seconds, then resets for the next round.
func (h *Hub) run() {
	var resetTimer <-chan time.Time = nil
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.broadcastState()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if _, exists := h.responses[client.name]; exists {
					delete(h.responses, client.name)
				}
				h.broadcastState()
			}
		case res := <-h.response:
			// If the round is locked, ignore new responses.
			if h.roundLocked {
				break
			}
			// Record the response if it hasn't been recorded yet.
			if _, exists := h.responses[res.client.name]; !exists {
				h.responses[res.client.name] = res.value
			}
			// Check if all clients have responded.
			if len(h.responses) == len(h.clients) && len(h.clients) > 0 {
				sum := 0
				for _, v := range h.responses {
					sum += v
				}
				avg := float64(sum) / float64(len(h.responses))
				resultState := h.buildState(&avg, "result")
				data, err := json.Marshal(resultState)
				if err == nil {
					h.broadcastMessage(data)
				}
				h.roundLocked = true
				// Set a timer to reset the round after 10 seconds.
				resetTimer = time.After(10 * time.Second)
			} else {
				// Otherwise, just update the state.
				h.broadcastState()
			}

		case <-resetTimer:
			// Clear responses and unlock the round.
			h.responses = make(map[string]int)
			h.roundLocked = false
			resetTimer = nil
			h.broadcastState()
		}
	}
}
