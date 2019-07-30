package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// "Adapted" from gorilla's websocket chat server example
// https://github.com/gorilla/websocket/blob/master/examples/chat/

const (
	writeWait = 20 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 256
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Client
type wsClient struct {
	hub *wsHub
	Conn *websocket.Conn
	send chan []byte
}

// Hub manage clients / broadcasts
type wsHub struct {
	wsClients map[*wsClient]bool
	broadcast chan []byte
	register chan *wsClient
	unregister chan *wsClient
}

func websocketUpgrade(hub *wsHub,w http.ResponseWriter, r *http.Request) {
	logger.Printf("[websocketService] *** Upgrade Request from %s",r.RemoteAddr)
	Conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[websocketService] %s Websocket Upgrade Error: %s", r.RemoteAddr, err)
		return
	}
	// DO NOT DEFER CONN CLOSE! duh... 2 hours, I'll never get back.
	client := &wsClient{hub: hub, Conn: Conn, send: make(chan []byte, 1024)}
	client.hub.register <- client
	go client.writePump()
	go client.readPump()
}

func (c *wsClient) readPump() {
	defer func() {
		logger.Printf("Unregistering: %s",c.Conn.RemoteAddr())
		c.hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		logger.Printf("[ws@%s] %s",c.Conn.RemoteAddr(),message)
		if err != nil {
			logger.Printf("Read error1: %v", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Printf("Read error2: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages
func (c *wsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		logger.Printf("Ending WritePump: %s",c.Conn.RemoteAddr())
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
			case message := <-c.send: {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				w, err := c.Conn.NextWriter(websocket.TextMessage)
				if err != nil {
					logger.Printf("Write Closed on [%s]",err);
					return
				}
				w.Write(message)
				n := len(c.send)
				for i := 0; i < n; i++ {
					w.Write(newline)
					w.Write(<-c.send)
				}
				if err := w.Close(); err != nil {
					logger.Printf("Closed on [%s]",err);
					return
				}
			}
			case <-ticker.C: {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					logger.Printf("Closed on [%s]",err);
					return
				}
			}
		}
	}
}

func newWsHub() *wsHub {
	return &wsHub{
		broadcast:  make(chan []byte),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		wsClients:    make(map[*wsClient]bool),
	}
}

func (h *wsHub) run() {
	for {
		select {
			case wsc := <-h.register: {
				h.wsClients[wsc] = true
			}
			case wsc := <-h.unregister: {
				if _, ok := h.wsClients[wsc]; ok {
					delete(h.wsClients, wsc)
					close(wsc.send)
				}
			}
			case message := <-h.broadcast: {
				for wsc := range h.wsClients {
					select {
						case wsc.send <- message:
						default:
							close(wsc.send)
							delete(h.wsClients, wsc)
					}
				}
			}
		}
	}
}
