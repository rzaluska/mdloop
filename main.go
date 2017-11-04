package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func withWatcherAndFile(watcher *fsnotify.Watcher, fileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		b, err := ReadAndRender(fileName)
		if err != nil {
			log.Println(err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
			log.Println(err)
		}

		for {
			select {
			case ev := <-watcher.Events:
				log.Println("event:", ev)
				b, err := ReadAndRender(fileName)
				if err != nil {
					log.Println(err)
				}
				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					log.Println(err)
				}
				if err := watcher.Add(fileName); err != nil {
					log.Println(err)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}
}

func ReadAndRender(fileName string) ([]byte, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	md := blackfriday.Run(b)
	return md, nil
}

func withFile(fileName string) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		b, err := ReadAndRender(fileName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if _, err := w.Write(b); err != nil {
			log.Println(err)
		}
	}
	return f
}

func main() {
	wdir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Can't get current working directory: %s", err)
		return
	}
	log.Printf("Working dir: %s", wdir)
	fileName := flag.String("f", "README.md", "")
	flag.Parse()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	wsHandler := withWatcherAndFile(watcher, *fileName)
	http.Handle("/", http.FileServer(http.Dir(wdir)))
	http.HandleFunc("/reload", wsHandler)
	log.Println("Running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
	watcher.Close()
}
