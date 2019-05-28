package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type wsConnection struct {
	Conn *websocket.Conn
	Label string
	Nick  string
}

var (
	WSBroadcast chan string
	wsUpgrader = websocket.Upgrader {
		ReadBufferSize: 1024,
		WriteBufferSize: 1024,
	}
	connectionList []wsConnection
)

func websocketUpgrade(w http.ResponseWriter, r *http.Request) {
	var newCon wsConnection
	logger.Printf("[websocketService] *** Upgrade Request from %s",r.RemoteAddr)
	// Perform WebSocket upgrade.
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[websocketService] %s Websocket Upgrade Error: %s", r.RemoteAddr, err)
		return
	}
	defer conn.Close()
	log.Printf("[websocketService] %s is connected.",r.RemoteAddr)
	err = conn.WriteMessage(websocket.TextMessage, []byte("{\"MOTD\" : \"GeekProjex.com\"}"))
	if err != nil {
		log.Printf("[websocketService] Write Error: %s", err)
		return
	}
	newCon.Conn = conn
	newCon.Label = r.RemoteAddr
	newCon.Nick = ""
	connectionList = append(connectionList,newCon)
	logger.Printf("Number Of Connections: %d",len(connectionList))
	for {
		msgType, bytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("[websocketService] Read Error: ", err)
			break
		}
		logger.Printf("[websocketService] [%d] Reading From %s",msgType,r.RemoteAddr)
		logger.Printf("[websocketService] [%d] %s",msgType,bytes)
	}
	log.Printf("[websocketService] [%s] WebSocket connection terminated.",newCon.Label)
	tempList := connectionList
	connectionList = nil
	for _,n := range tempList {
		if n.Conn != conn {
			connectionList = append(connectionList,n)
		}
	}
	logger.Printf("Number Of Connections: %d",len(connectionList))
}