package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const address = "localhost:30001"

var Version = []byte("V2")
var Vok = []byte("VOK")
var Read = []byte("R")
var Write = []byte("W")
var Wok = []byte("WOK")

var fromConsole = make(chan []byte)
var toConsole = make(chan []byte)

func chromeReader(ws *websocket.Conn, dataCh chan []byte, errCh chan error) {
	msg := make([]byte, 1024*10)
	for {
		n, err := ws.Read(msg)
		if err != nil {
			errCh <- err
			return
		}
		dataCh <- msg[:n]
	}
}

func wsHandler(ws *websocket.Conn) {
	ws.Write(Version)
	msg := make([]byte, 1024*10)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(msg[:n], Vok) {
		log.Fatal("version mismatch")
	}
	log.Println("connected.")
	chromeData := make(chan []byte)
	chromeErr := make(chan error)
	go chromeReader(ws, chromeData, chromeErr)
	for {
		select {
		case chromeMsg := <-chromeData:
			if chromeMsg[0] == 'C' {
				continue
			} else if bytes.Equal(chromeMsg, Wok) {
				log.Println("copy completed.")
			} else if bytes.Equal(chromeMsg[0:1], Read) {
				log.Println("paste completed.")
				toConsole <- chromeMsg[1:]
			}
		case err := <-chromeErr:
			fmt.Println(err)
			break
		case consoleMsg := <-fromConsole:
			if bytes.Equal(consoleMsg, Read) {
				_, err = ws.Write(Read)
				if err != nil {
					log.Println(err)
					break
				}
			} else if bytes.Equal(consoleMsg[0:1], Write) {
				_, err = ws.Write(consoleMsg)
				if err != nil {
					log.Println(err)
					break
				}

			}
		}
	}
}

func copyHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	body = append(Write, body...)
	fromConsole <- body
	w.WriteHeader(http.StatusOK)
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	fromConsole <- Read
	msg := <-toConsole
	w.Write(msg)
}

func main() {
	serve := flag.Bool("serve", false, "Start clipboard server.")
	copy := flag.Bool("copy", false, "Copy to crouton clipboard.")
	paste := flag.Bool("paste", false, "Paste from crouton clipboard.")
	flag.Parse()

	if *serve {
		http.Handle("/", websocket.Handler(wsHandler))
		http.HandleFunc("/copy", copyHandler)
		http.HandleFunc("/paste", pasteHandler)
		err := http.ListenAndServe(address, nil)
		if err != nil {
			panic(err.Error())
		}
	} else if *copy {
		resp, err := http.Post(fmt.Sprintf("http://%s/copy", address), "application/octet-stream", bufio.NewReader(os.Stdin))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp.Body.Close()
	} else if *paste {
		resp, err := http.Get(fmt.Sprintf("http://%s/paste", address))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Print(string(body))
	} else {
		fmt.Println("Must specify either -serve or -copy or -paste")
		os.Exit(1)
	}
}
