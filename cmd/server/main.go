package main

import (
	"log"
	"net/http"
	"GomokuRenjuOnline-Backend/internal/hub"
	"GomokuRenjuOnline-Backend/internal/transport"
)

func main() {
	gameHub := hub.NewHub()
	go gameHub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		transport.HandleConnections(gameHub, w, r)
	})
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}