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
	"regexp"
)

type PhotoFile struct {
	path    string
	modTime time.Time
}

func (p *PhotoFile) String() string {
	return p.path
}

type PhotoFileService struct {
	photoFolder                    string
	photoFiles                     []*PhotoFile
	pathToPhotoFile                map[string]*PhotoFile
	newPhotoFilesChan              chan *PhotoFile

	newPhotoListeners              map[chan *PhotoFile]bool

	RegisterNewPhotoListenerChan   chan chan *PhotoFile
	UnregisterNewPhotoListenerChan chan chan *PhotoFile
}

func NewPhotoFileService(photoFolder string) *PhotoFileService {
	pfs := new(PhotoFileService)
	pfs.photoFolder = photoFolder
	pfs.newPhotoFilesChan = make(chan *PhotoFile)
	pfs.pathToPhotoFile = make(map[string]*PhotoFile)
	pfs.newPhotoListeners = make(map[chan *PhotoFile]bool)

	pfs.RegisterNewPhotoListenerChan = make(chan chan *PhotoFile)
	pfs.UnregisterNewPhotoListenerChan = make(chan chan *PhotoFile)

	return pfs
}

func (pfs *PhotoFileService) GetRecentPhotoFiles() []*PhotoFile {
	return pfs.photoFiles[:10]
}

func (pfs *PhotoFileService) scanPhotoFolder() {
	log.Println("PhotoFileService: Start scanning of the photo folder at", pfs.photoFolder)
	files, err := ioutil.ReadDir(pfs.photoFolder)
	if err != nil {
		log.Fatal(err)
	}

	var photoFiles []PhotoFile
	for _, file := range files {
		path := strings.Join([]string{pfs.photoFolder, file.Name()}, "/")
		photoFile := PhotoFile{path: path, modTime: file.ModTime()}
		photoFiles = append(photoFiles, photoFile)
	}

	sort.Sort(ByModTime(photoFiles))

	for _, photoFile := range photoFiles {
		pfs.addPhotoFile(photoFile)
	}
	log.Println("PhotoFileService: Finished scanning of the photo folder at", pfs.photoFolder)
}

func (pfs *PhotoFileService) addPhotoFile(pf PhotoFile) bool {
	if _, exists := pfs.pathToPhotoFile[pf.path]; exists {
		log.Println("PhotoFileService: Skip adding photo file at path", pf.path, "is already registered!")
		return false
	}

	match, _ := regexp.MatchString("(?i).+\\.(png|jpeg|JPG)", pf.path)
	if !match {
		return false
	}

	pfs.pathToPhotoFile[pf.path] = &pf;
	pfs.photoFiles = append(pfs.photoFiles, &pf)

	log.Println("PhotoFileService: Add new photo file at path", pf.path)
	pfs.newPhotoFilesChan <- &pf

	return true
}

func (pfs *PhotoFileService) watchFolderLoop(photoFolder string) {
	photoFolderPath, err := filepath.Abs(photoFolder)
	if err != nil {
		log.Fatal("Failed to get absolute path of", photoFolder)
	}
	log.Println("PhotoFileService: Starting to watch for file changes in", photoFolderPath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	watcher.Add(photoFolderPath)

	for {
		select {
		case event := <-watcher.Events:
			//log.Println("event:", event)

			if event.Op == fsnotify.Create {
				absolutePath, _ := filepath.Abs(event.Name)

				info, err := os.Stat(absolutePath)
				if err != nil {
					log.Println("Failed to get file info for path", absolutePath)
					continue
				}

				pf := PhotoFile{path: absolutePath, modTime: info.ModTime()}
				pfs.addPhotoFile(pf)
			}


		case err := <-watcher.Errors:
			log.Println("FileWatcherError:", err)
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

func (pfs *PhotoFileService) broadCastNewPhotoFile(pf *PhotoFile) {
	log.Printf("PhotoFileService: Broadcast new photo file %s to %v listeners", pf, len(pfs.newPhotoListeners))
	for listener := range pfs.newPhotoListeners {
		listener <- pf
	}
}

func (pfs *PhotoFileService) unregisterNewPhotoListener(listener chan *PhotoFile) {
	if _, ok := pfs.newPhotoListeners[listener]; !ok {
		return // is not registered
	}

	delete(pfs.newPhotoListeners, listener)
	close(listener)
}

func (pfs *PhotoFileService) registerNewPhotoListener(listener chan *PhotoFile) {
	pfs.newPhotoListeners[listener] = true
}

func (pfs *PhotoFileService) loop() {
	go pfs.watchFolderLoop(pfs.photoFolder)
	go pfs.newPhotoListenerLoop()
}

type ByModTime []PhotoFile

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