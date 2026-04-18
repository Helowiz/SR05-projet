package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn = nil

func do_webserver(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bonjour depuis le serveur web en Go !")
}

func do_websocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	cnx, err := upgrader.Upgrade(w, r, nil)
	ws = cnx
	if err != nil {
		fmt.Println("upgrade:", err)
		return
	}

	for {
		_, message, err := cnx.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}
		fmt.Println("réception : " + string(message))
	}
}

func ws_send(msg string) {
	if ws == nil {
		fmt.Println("ws_send", "websocket non ouverte")
	} else {
		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			fmt.Println("ws_send", "WriteMessage : "+string(err.Error()))
		} else {
			fmt.Println("ws_send", "sending "+msg)
		}
	}
}

func do_send() {
	for {
		ws_send("Hello depuis le serveur")
		time.Sleep(time.Duration(4) * time.Second)
	}
}

func main() {
	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", do_websocket)
	go do_send()
	http.ListenAndServe(*addr+":"+*port, nil)
}
