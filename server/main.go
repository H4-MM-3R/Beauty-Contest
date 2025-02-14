package server

import (
	"flag"
	"log"
	"net/http"
)

var hubs = make(map[string]*Hub)
var addr = flag.String("addr", ":8080", "http service address")

func StartServer() {
	flag.Parse()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/create-hub", createHub)
	http.HandleFunc("/ws", serveWs)
	log.Printf("Listening on http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		serveHome(w, r)
	} else {
		serveGame(w, r)
	}
}
