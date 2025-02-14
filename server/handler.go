package server

import (
	"encoding/json"
	"fmt"
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

    // Get the name from the query parameter.
	name := r.URL.Query().Get("name")
	if name == "" {
		// If no name provided, serve a simple HTML+JS prompt.
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!doctype html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Enter Your Name</title>
			</head>
			<body>
				<script>
					var name = prompt("Enter your name:");
					if (name) {
						// Redirect to the same URL with the entered name.
						window.location.href = window.location.pathname + "?name=" + encodeURIComponent(name);
					} else {
						document.write("Name is required to join the game.");
					}
				</script>
			</body>
			</html>
		`))
		return
	}


	tmplPath := filepath.Join("templates", "game.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// name := r.URL.Query().Get("name")
	// if name == "" {
	// 	http.Error(w, "Name is required", http.StatusBadRequest)
	// 	return
	// }
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

// func serveWs(w http.ResponseWriter, r *http.Request) {
// }

func serveWs(w http.ResponseWriter, r *http.Request) {
	fmt.Print("new connection\n")
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
	client, ok := hub.players[name]
	if ok {
		fmt.Println("client already exists")
		client.conn = conn
		client.send = make(chan []byte, 256)
	} else {

		client = &Client{
			hub:        hub,
			conn:       conn,
			send:       make(chan []byte, 256),
			name:       name,
			score:      3,
			eliminated: false,
		}
	}

    hub.register <- client

	go client.writePump()
	go client.readPump()
}
