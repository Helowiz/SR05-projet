package main

import (
	"SR05_projet/display"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

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
		display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "upgrade failed : "+err.Error())

		return
	}

	for {
		_, message, err := cnx.ReadMessage()
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "cnx.ReadMessage() failed : "+err.Error())
			return
		}
		//fmt.Println("réception : " + string(message))
		switch string(message) {

		case "section_critique":
			fmt.Println("debut_sc")

		case "f_section_critique":
			fmt.Println("fin_sc")

		}
	}
}

func ws_send(msg string) {
	if ws == nil {
		display.Warning("SERVER :"+strconv.Itoa(os.Getpid()), "ws_send()", "Websocket non ouverte")

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
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_send()", "Scanln failed : "+err.Error())
			return
		}
		ws_send(msg)
	}
}

func main() {
	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", do_websocket)
	//go do_send()
	http.ListenAndServe(*addr+":"+*port, nil)
}
