package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"strings"
	"sort"
	"os"
)

type photoFile struct {
	path    string
	modTime time.Time
}

type PhotoFileService struct {
	photoFolder       string
	photoFiles        []*photoFile
	pathToPhotoFile   map[string]*photoFile
	newPhotoFilesChan chan *photoFile

	newPhotoListeners map[chan *photoFile]bool

	RegisterNewPhotoListenerChan chan chan *photoFile
	UnregisterNewPhotoListenerChan chan chan *photoFile
}

func NewPhotoFileService(photoFolder string) *PhotoFileService {
	pfs := new(PhotoFileService)
	pfs.photoFolder = photoFolder
	pfs.newPhotoFilesChan = make(chan *photoFile)
	pfs.pathToPhotoFile = make(map[string]*photoFile)
	pfs.newPhotoListeners = make(map[chan *photoFile]bool)

	pfs.RegisterNewPhotoListenerChan = make(chan chan *photoFile)
	pfs.UnregisterNewPhotoListenerChan = make(chan chan *photoFile)

	return pfs
}

func (pfs *PhotoFileService) scanPhotoFolder() {
	log.Println("Start scanning of the photo folder at", pfs.photoFolder)
	files, err := ioutil.ReadDir(pfs.photoFolder)
	if err != nil {
		log.Fatal(err)
	}

	var photoFiles []photoFile
	for _, file := range files {
		path := strings.Join([]string{pfs.photoFolder, file.Name()}, "/")
		photoFile := photoFile{path: path, modTime: file.ModTime()}
		photoFiles = append(photoFiles, photoFile)
	}

	sort.Sort(ByModTime(photoFiles))

	for _, photoFile := range photoFiles {
		pfs.addPhotoFile(photoFile)
	}
	log.Println("Finished scanning of the photo folder at", pfs.photoFolder)
}

func (pfs *PhotoFileService) addPhotoFile(pf photoFile) bool {
	if _, exists := pfs.pathToPhotoFile[pf.path]; exists {
		log.Println("Skip adding photo file at path", pf.path, "is already registered!")
		return false
	}

	pfs.pathToPhotoFile[pf.path] = &pf;
	pfs.photoFiles = append(pfs.photoFiles, &pf)

	log.Println("Add new photo file at path", pf.path)
	pfs.newPhotoFilesChan <- &pf

	return true
}

func (pfs *PhotoFileService) watchFolderLoop(photoFolder string) {
	photoFolderPath, err := filepath.Abs(photoFolder)
	if err != nil {
		log.Fatal("Failed to get absolute path of", photoFolder)
	}
	log.Println("Starting to watch for file changes in", photoFolderPath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	watcher.Add(photoFolderPath)

	for {
		select {
		case event := <-watcher.Events:
			log.Println("event:", event)

			if event.Op == fsnotify.Create {
				absolutePath, _ := filepath.Abs(event.Name)

				info, err := os.Stat(absolutePath)
				if err != nil {
					log.Println("Failed to get file info for path", absolutePath)
					continue
				}

				pf := photoFile{path: absolutePath, modTime: info.ModTime()}
				pfs.addPhotoFile(pf)
			}


		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func (pfs *PhotoFileService) newPhotoListenerLoop() {
	for {
		select {
		case photoFile := <-pfs.newPhotoFilesChan:
			pfs.broadCastNewPhotoFile(photoFile)
		case listener := <-pfs.RegisterNewPhotoListenerChan:
			pfs.registerNewPhotoListener(listener)
		case listener := <-pfs.UnregisterNewPhotoListenerChan:
			pfs.unregisterNewPhotoListener(listener)
		}
	}
}

func (pfs *PhotoFileService) broadCastNewPhotoFile(pf *photoFile) {
	for listener := range pfs.newPhotoListeners {
		select {
		case listener <- pf:
			default:
			pfs.unregisterNewPhotoListener(listener)
		}
	}
}

func (pfs *PhotoFileService) unregisterNewPhotoListener(listener chan *photoFile) {
	if _, ok := pfs.newPhotoListeners[listener]; !ok {
		return // is not registered
	}

	delete(pfs.newPhotoListeners, listener)
	close(listener)
}

func (pfs *PhotoFileService) registerNewPhotoListener(listener chan *photoFile) {
	pfs.newPhotoListeners[listener] = true

	for _, photoFile := range pfs.photoFiles[0:10] {
		listener <- photoFile
	}
}

func (pfs *PhotoFileService) loop() {
	go pfs.watchFolderLoop(pfs.photoFolder)
	go pfs.newPhotoListenerLoop()
}

type ByModTime []photoFile

func (s ByModTime) Len() int {
	return len(s)
}

func (s ByModTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByModTime) Less(i, j int) bool {
	iTime, jTime := s[i].modTime, s[j].modTime

	return iTime.Before(jTime)
}