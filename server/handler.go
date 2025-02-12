package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// serveHome serves the home page.
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tmplPath := filepath.Join("templates", "home.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// serveGame serves the game page (game.html). It expects a hub hash in the URL
// (e.g. /<hubHash>) and a query parameter "name" containing the client's name.
func serveGame(w http.ResponseWriter, r *http.Request) {
	hubHash := r.URL.Path[1:]
	if _, ok := hubs[hubHash]; !ok {
		http.Error(w, "Hub not found", http.StatusNotFound)
		return
	}
	tmplPath := filepath.Join("templates", "game.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	data := map[string]string{
		"HubHash": hubHash,
		"Name":    name,
	}
	tmpl.Execute(w, data)
}

// createHub generates a new hub, starts its run loop, and returns its hash as JSON.
func createHub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	hubHash, err := GenerateHubHash()
	if err != nil {
		http.Error(w, "Failed to create hub", http.StatusInternalServerError)
		return
	}
	hub := newHub()
	hubs[hubHash] = hub
	go hub.run()

	resp := map[string]string{
		"hub": hubHash,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// serveWs upgrades an HTTP connection to a WebSocket.
// It requires query parameters "hub" (the hub hash) and "name" (the client's unique name).
func serveWs(w http.ResponseWriter, r *http.Request) {
	hubHash := r.URL.Query().Get("hub")
	name := r.URL.Query().Get("name")
	if hubHash == "" || name == "" {
		http.Error(w, "Hub and name parameters required", http.StatusBadRequest)
		return
	}
	hub, ok := hubs[hubHash]
	if !ok {
		http.Error(w, "Hub not found", http.StatusNotFound)
		return
	}
	// Ensure the chosen name is unique in this hub.
	for client := range hub.clients {
		if client.name == name {
			http.Error(w, "Name already in use", http.StatusBadRequest)
			return
		}
	}
	if len(hub.clients) >= 7 {
		http.Error(w, "Hub is full", http.StatusForbidden)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	// Create a new client with an initial score of 10.
	client := &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		name:       name,
		score:      10,
		eliminated: false,
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

