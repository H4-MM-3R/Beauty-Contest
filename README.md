# Beauty Contest 

implementaion of keynesian beauty contest based on the game from alice in borderland (S2-EP6)

# Folder structure

```
project-root/
├── go.mod                // Go module file
├── go.sum                // Go module checksum file
├── main.go               // Entry point
├── server/               // All the server-side code
│   ├── main.go               // Server side entry point: setup routes, global hub manager, server start
│   ├── hub.go                // Hub struct & methods (hub management, averaging logic)
│   ├── client.go             // Client struct & methods (WebSocket read/write pumps)
│   ├── handlers.go           // HTTP and WebSocket handlers (create-hub, serveHome, serveWs)
│   └── utils.go              // (Optional) Utility functions, e.g. generating unique hashes
├── templates/            // HTML templates for the front end
│   ├── home.html         // Home page: hub creation/joining UI
│   └── game.html  
```

# How to run

``` bash
go run main.go
```
