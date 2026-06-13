package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"encoding/json"
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

var sauvMsg []string

func format_msg_json(msg string) string {
	data := make(map[string]string)

	keys := []string{"to", "info", "id", "from", "parent", "elu", "color", "nbvoisinsAttendus", "address", "port", "id_new", "neighbor_id", "add_member"}

	for _, key := range keys {
		val := protocol.Findval(msg, key, "")
		if val != "" {
			data[key] = val
		}
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Erreur JSON:", err)
		return "{}"
	}
	return string(jsonBytes)
}

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
	sauvMsg = nil

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

func listen_for_net_msg() {
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "listen_for_clt_msg", "Scanln failed : "+err.Error())
			return
		}
		msg_type := protocol.Findval(msg, "to", "LogNet")
		if msg_type == "log" {
			ws_send(format_msg_json(msg))
		}

	}
}

func main() {

	var port = flag.String("port", "12345", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", do_websocket)
	go listen_for_net_msg()
	http.ListenAndServe(*addr+":"+*port, nil)
}
