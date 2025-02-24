package server

import (
	"encoding/json"
	"fmt"
	"time"
)

// -------------------- Types --------------------

type PlayerState struct {
	Name       string      `json:"name"`
	Response   interface{} `json:"response"`
	Score      int         `json:"score"`
	Eliminated bool        `json:"eliminated"`
}

type StateMessage struct {
	Type    string        `json:"type"`
	Players []PlayerState `json:"players"`
	Target  *float64      `json:"target,omitempty"`
	Average *float64      `json:"average,omitempty"`
	Winners []string      `json:"winners,omitempty"`
}

type Response struct {
	client *Client
	value  int
}

type Hub struct {
	clients     map[*Client]bool
	register    chan *Client
	unregister  chan *Client
	response    chan *Response
	responses   map[*Client]int
	roundLocked bool
	gameOver    bool
	players     map[string]*Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		response:   make(chan *Response),
		responses:  make(map[*Client]int),
		players:    make(map[string]*Client),
	}
}

// -------------------- Helper Functions --------------------

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// broadcastMessage sends the given JSON message to all clients.
func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			delete(h.clients, client)
			close(client.send)
		}
	}
}

// broadcastState builds a StateMessage and broadcasts it.
func (h *Hub) broadcastState(msgType string, avg, target *float64, winners []string) {
	players := []PlayerState{}
	for client := range h.clients {
		var resp interface{}
		if client.eliminated {
			// For eliminated players, we may simply show no response.
			resp = nil
		} else {
			if r, ok := h.responses[client]; ok {
				resp = r
			} else {
				resp = "Waiting to respond..."
			}
		}
		players = append(players, PlayerState{
			Name:       client.name,
			Response:   resp,
			Score:      client.score,
			Eliminated: client.eliminated,
		})
	}
	stateMsg := StateMessage{
		Type:    msgType,
		Players: players,
	}
	if target != nil {
		stateMsg.Target = target
	}
	if avg != nil {
		stateMsg.Average = avg
	}
	if winners != nil {
		stateMsg.Winners = winners
	}
	data, err := json.Marshal(stateMsg)
	if err != nil {
		fmt.Println("Error marshaling state:", err)
		return
	}
	h.broadcastMessage(data)
}

// -------------------- Main Run Loop --------------------

// run listens for registrations, unregistrations, responses, and timers.
// It implements the round logic and game progression.
func (h *Hub) run() {
	var resetTimer <-chan time.Time = nil

	for {
		select {
		case client := <-h.register:
			if _, ok := h.players[client.name]; !ok {
				h.players[client.name] = client
			}
			h.clients[client] = true
			h.broadcastState("state", nil, nil, nil)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.responses, client)
				h.broadcastState("state", nil, nil, nil)
			}

		case res := <-h.response:
			// Ignore responses if the round is locked or the client is eliminated.
			if h.roundLocked {
				break
			}
			if res.client.eliminated {
				h.broadcastState("eliminated", nil, nil, nil)
			}

			// Record the response only if not already recorded.
			if _, exists := h.responses[res.client]; !exists {
				h.responses[res.client] = res.value
			}

			// Count active (non‑eliminated) players.
			activeCount := 0
			for client := range h.clients {
				if !client.eliminated {
					activeCount++
				}
			}

			if len(h.responses) == activeCount && activeCount > 0 {
                var winners []string
                var avg float64
                var target float64
                switch activeCount {
                case 2:
                    winners, avg, target = twoPlayerLogic(h)
                default:
                    winners, avg, target = defaultLogic(h, activeCount)
                }

                updateScores(h, winners)

				// Check for game over.
				activePlayers := 0
				for client := range h.clients {
					if !client.eliminated {
						activePlayers++
					}
				}

				if activePlayers <= 1 {
					h.broadcastState("gameover", &avg, &target, winners)
					h.roundLocked = true
				} else {
					// Broadcast round result (including responses, target and winner).
					h.broadcastState("result", &avg, &target, winners)
					h.roundLocked = true
					resetTimer = time.After(5 * time.Second)
				}

			} else {
				// Not all active players have responded yet: broadcast ongoing state.
				h.broadcastState("state", nil, nil, nil)
			}

		case <-resetTimer:
			// Reset round state.
			h.responses = make(map[*Client]int)
			h.roundLocked = false
			h.broadcastState("state", nil, nil, nil)
			resetTimer = nil
		}
	}
}
