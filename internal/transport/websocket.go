package transport

import (
	"log"
	"GomokuRenjuOnline-Backend/internal/hub"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有连接
	},
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func HandleConnections(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	hub.CreateAndRegisterClient(ws)
}