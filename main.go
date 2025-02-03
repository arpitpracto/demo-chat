package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr:     "https://evolved-midge-10211.upstash.io", // Replace with your Upstash Redis URL
		Password: "ASfjAAIjcDFlNzcyMThkOWE5MmM0ZDdhOGQ2OTE5ZTIyNmZlZmY4NXAxMA", // If authentication is required
		DB:       0,                              // Default DB
	})
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins (for development). Secure this for production.
		},
	}
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws/{channel}", handleWebSocket)

	fs := http.FileServer(http.Dir("./web")) // Serve frontend files
	r.PathPrefix("/").Handler(fs)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channel := vars["channel"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	sub := rdb.Subscribe(ctx, channel)
	defer sub.Close()
	ch := sub.Channel()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}

			if err := rdb.Publish(ctx, channel, string(msg)).Err(); err != nil {
				log.Println("Publish error:", err)
				return
			}
		}
	}()

	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Println("Write error:", err)
			return
		}
	}
}
