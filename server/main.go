package server

import (
	"flag"
	"log"
	"net/http"
)

// hubs maps hub hash strings to active Hub instances.
var hubs = make(map[string]*Hub)

var addr = flag.String("addr", ":8080", "http service address")

// Run starts the HTTP server.
func StartServer() {
	flag.Parse()

	// "/" serves home.html or game.html depending on the URL.
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/create-hub", createHub)
	http.HandleFunc("/ws", serveWs)

	log.Printf("Listening on http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// defaultHandler dispatches "/" to the home page or the game page.
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// If the URL is exactly "/", serve home page.
	if r.URL.Path == "/" {
		serveHome(w, r)
	} else {
		serveGame(w, r)
	}
}

