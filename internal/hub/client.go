package hub

import (
	"log"
	"GomokuRenjuOnline-Backend/pkg/protocol"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID     int
	hub    *Hub
	room   *Room
	conn   *websocket.Conn
	send   chan protocol.OutboundMessage
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg protocol.InboundMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading json: %v", err)
			break
		}
		
		if c.room != nil {
			c.room.forward <- messageFromClient{client: c, message: &msg}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			log.Printf("error writing json: %v", err)
			break
		}
	}
}