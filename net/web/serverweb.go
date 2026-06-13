package main

import (
	"SR05-etude/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn = nil

type Event struct {
	from    string
	content string
}

const NET string = "net"

var sauvMsg []string

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

	for _, msg := range sauvMsg {
		ws_send(msg)
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

	msg_type := protocol.Findval(msg, "to", "LogNet")
	if msg_type != "log" {
		return
	}

	if ws == nil {
		sauvMsg = append(sauvMsg, msg)
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

func listen_for_net_msg(eventQueue chan<- Event) {
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "listen_for_clt_msg", "Scanln failed : "+err.Error())
			return
		}
		ws_send(msg)
	}
}

func main() {
	var eventQueue = make(chan Event, 100)

	var port = flag.String("port", "12345", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", do_websocket)
	go listen_for_net_msg(eventQueue)
	http.ListenAndServe(*addr+":"+*port, nil)
}
