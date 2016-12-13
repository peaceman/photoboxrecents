package main

import (
	"github.com/gorilla/websocket"
	"time"
	"log"
	"fmt"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type WebClient struct {
	unregisterWebClientChan chan<- *WebClient
	webSocket *websocket.Conn
	SendChan chan []byte
}

func NewWebClient(webSocket *websocket.Conn, unregisterWebClientChan chan<- *WebClient) *WebClient {
	webClient := new(WebClient)
	webClient.webSocket = webSocket
	webClient.unregisterWebClientChan = unregisterWebClientChan
	maxMessageSize := 1024
	webClient.SendChan = make(chan []byte, maxMessageSize * 2)

	return webClient
}

func (webClient *WebClient) closeConnection() {
	webClient.unregisterWebClientChan <- webClient
	webClient.webSocket.Close()
}

func (webClient *WebClient) readLoop() {
	defer webClient.closeConnection()

	ws := webClient.webSocket
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error {
		return webClient.webSocket.SetReadDeadline(time.Now().Add(pongWait));
	});

	for {
		_, _, err := webClient.webSocket.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (webClient *WebClient) writeLoop() {
	defer webClient.closeConnection()

	pingTicker := time.NewTicker(pingPeriod)

	for {
		select {
		case <-pingTicker.C:
			if err := webClient.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println("PingTicker Error", err)
				return
			}

		case message, ok := <-webClient.SendChan:
			if !ok {
				webClient.write(websocket.CloseMessage, []byte{})
				return
			}

			if err := webClient.write(websocket.TextMessage, message); err != nil {
				log.Println("WebClient SendChan Error", err)
				return
			}
		}
	}
}

func (webClient *WebClient) write(messageType int, payload []byte) error {
	webClient.webSocket.SetWriteDeadline(time.Now().Add(writeWait))
	return webClient.webSocket.WriteMessage(messageType, payload)
}

func (webClient *WebClient) StartServing() {
	go webClient.writeLoop()
	webClient.readLoop()
}

func (webClient *WebClient) String() string {
	return fmt.Sprintf("%v", webClient.webSocket.RemoteAddr())
}