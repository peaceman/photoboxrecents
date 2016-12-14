package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func (r *http.Request) bool { return true },
}

type WebClientHub struct {
	webClients       map[*WebClient]bool
	register         chan *WebClient
	unregister       chan *WebClient
	newPhotoFiles    chan *PhotoFile
	photoFileService *PhotoFileService
}

func NewWebClientHub(photoFileService *PhotoFileService) *WebClientHub {
	hub := new(WebClientHub)
	hub.webClients = make(map[*WebClient]bool)
	hub.register = make(chan *WebClient)
	hub.unregister = make(chan *WebClient)
	hub.photoFileService = photoFileService

	newPhotoFiles := make(chan *PhotoFile)
	photoFileService.RegisterNewPhotoListenerChan <- newPhotoFiles

	hub.newPhotoFiles = newPhotoFiles

	return hub
}

func (hub *WebClientHub) loop() {
	for {
		select {
		case webClient := <-hub.register:
			hub.registerWebClient(webClient)
		case webClient := <-hub.unregister:
			hub.unregisterWebClient(webClient)
		case photoFile := <-hub.newPhotoFiles:
			photoFile.String()
			hub.broadcastPhotoFile(photoFile)
		}
	}
}

func (hub *WebClientHub) broadcastPhotoFile(photoFile *PhotoFile) {
	log.Printf("Broadcast PhotoFile to %d WebClients | %s", len(hub.webClients), photoFile)

	for webClient := range hub.webClients {
		select {
		case webClient.SendChan <- []byte(photoFile.String()):
		default:
			hub.unregisterWebClient(webClient)
		}
	}
}

func (hub *WebClientHub) registerWebClient(webClient *WebClient) {
	hub.webClients[webClient] = true
	log.Println("Registered WebClient with address", webClient)

	for _, photoFile := range hub.photoFileService.GetRecentPhotoFiles() {
		webClient.SendChan <- []byte(photoFile.String())
	}
}

func (hub *WebClientHub) unregisterWebClient(webClient *WebClient) {
	if _, ok := hub.webClients[webClient]; !ok {
		return // given WebClient is not registered
	}

	delete(hub.webClients, webClient)
	close(webClient.SendChan)
	log.Println("Unregistered WebClient with address", webClient)
}

func (hub *WebClientHub) handleWebClientConnection(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(responseWriter, "Method not allowed", 405)
		return
	}

	webSocket, err := upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	webClient := NewWebClient(webSocket, hub.unregister)
	hub.register <- webClient

	webClient.StartServing()
}