package main

import (
	// "fmt"
	"flag"
	"log"
	"net/http"
)

var config *Config = &Config{}

func init() {
	flag.StringVar(&config.listenAddress, "listen", ":6740", "http listen address")
	flag.StringVar(&config.photoFolder, "photo-folder", "", "path to the photo folder")
}

func main() {
	flag.Parse()

	done := make(chan bool)

	photoFileService := NewPhotoFileService(config.photoFolder)
	go photoFileService.loop()

	webClientHub := NewWebClientHub(photoFileService)
	go webClientHub.loop()
	go photoFileService.scanPhotoFolder()

	http.Handle("/", http.FileServer(http.Dir("photoboxrecents-web/dist")))
	http.HandleFunc("/data", webClientHub.handleWebClientConnection)
	http.Handle("/photo/", http.StripPrefix("/photo/", http.FileServer(http.Dir(config.photoFolder))))

	log.Println("Open HTTP socket at:", config.listenAddress)
	err := http.ListenAndServe(config.listenAddress, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}


	//go func() {
	//	for {
	//		select {
	//		case lel := <-(<-chan *photoFile)(photoFileService.newPhotoFilesChan):
	//			log.Printf("%#v", lel)
	//		}
	//	}
	//}()


	// watcher, err := fsnotify.NewWatcher()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer watcher.Close()


	// watcher.Add("web")

	// go func() {
	// 	for {
	// 		select {
	// 		case event := <-watcher.Events:
	// 			log.Println("event:", event)
	// 			log.Printf("Event: %v\n", event)

	// 			absolutePath, _ := filepath.Abs(event.Name)
	// 			log.Println("ABS:", absolutePath)
	// 		case err := <-watcher.Errors:
	// 			log.Println("error:", err)
	// 		}
	// 	}
	// }()
	<-done
}
