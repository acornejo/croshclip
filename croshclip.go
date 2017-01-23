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

const (
	Version = "V2"
	Vok     = "VOK"
	Read    = "R"
	Write   = "W"
	Wok     = "WOK"
)

var fromConsole = make(chan []byte)
var toConsole = make(chan []byte)

func chromeReader(ws *websocket.Conn, dataCh chan []byte) {
	msg := make([]byte, 1024*10)
	for {
		n, err := ws.Read(msg)
		if err != nil {
            close(dataCh)
			return
		}
		dataCh <- msg[:n]
	}
}

func wsHandler(ws *websocket.Conn) {
    defer ws.Close()
    _, err := ws.Write([]byte(Version))
    if err != nil {
        log.Println(err)
        return
    }
	chromeData := make(chan []byte)
	go chromeReader(ws, chromeData)
	for {
		select {
		case chromeMsg, ok := <-chromeData:
            if !ok {
                log.Println("closed.")
                return
            }
			if chromeMsg[0] == 'C' {
                // skip control messages
				continue
            } else if bytes.Equal(chromeMsg, []byte(Vok)) {
                log.Println("connected.")
            } else if bytes.Equal(chromeMsg, []byte(Wok)) {
				log.Println("copy completed.")
			} else if bytes.Equal(chromeMsg[0:1], []byte(Read)) {
				log.Println("paste completed.")
				toConsole <- chromeMsg[1:]
            } else {
                log.Println("unknown extension message ", chromeMsg)
            }
		case consoleMsg := <-fromConsole:
			if bytes.Equal(consoleMsg, []byte(Read)) {
				_, err = ws.Write([]byte(Read))
				if err != nil {
					log.Println(err)
					break
				}
			} else if bytes.Equal(consoleMsg[0:1], []byte(Write)) {
				_, err = ws.Write(consoleMsg)
				if err != nil {
					log.Println(err)
					break
				}
			} else {
                log.Println("unkown console message ", consoleMsg)
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
	body = append([]byte(Write), body...)
	fromConsole <- body
	w.WriteHeader(http.StatusOK)
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	fromConsole <- []byte(Read)
	msg := <-toConsole
	w.Write(msg)
}

func main() {
	const address = "localhost:30001"
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
