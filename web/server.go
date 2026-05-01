package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"SR05_projet/shape"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn = nil

var whiteboard shape.WhiteBoard
var lastOpe string = ""
var in_section_critique bool = false
var is_snaped = false
var pendingOpe = ""
var port *string
var addr *string

//func do_local_snapshot() {
//
//		display.Info("SNAP", "snapshot", whiteboard.String())
//		fmt.Println(protocol.Msg_format("type", "snapshot"))
//
//}

func demander_sc() {
	msg := protocol.Msg_format("type", "fromapp_debut_sc")
	fmt.Println(msg)
}

func liberer_sc(newOpe string) {
	msg := protocol.Msg_format("type", "fromapp_fin_sc") + protocol.Msg_format("data", newOpe)
	fmt.Println(msg)
}

func do_webserver(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bonjour depuis le serveur web en Go !")
}

func do_websocket(w http.ResponseWriter, r *http.Request, active chan bool) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	cnx, err := upgrader.Upgrade(w, r, nil)
	ws = cnx
	if err != nil {
		display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "upgrade failed : "+err.Error())

		return
	}
	ws_send("data=" + lastOpe)

	for {
		_, message, err := cnx.ReadMessage()
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "cnx.ReadMessage() failed : "+err.Error())
			return
		}
		//fmt.Println("réception : " + string(message))
		parts := strings.Split(string(message), "=")
		prefix, suffix := parts[0], strings.TrimSpace(parts[1])
		display.Info("", "do_websocket", "received : "+string(message))

		switch prefix {
		case "section_critique": // juste pour le debug, a enlever ce cas apres
			if suffix == "activate" {
				demander_sc()
			}
			if suffix == "deactivate" {
				liberer_sc("0")
			}
		case "data":
			pendingOpe = suffix // en attente
			wait_for_sc(active)
			//case "snapshot":
			//	do_local_snapshot()
		}
	}
}

func ws_send(msg string) {
	if ws == nil {
		display.Warning("SERVER :"+strconv.Itoa(os.Getpid()), "ws_send()", "Websocket non ouverte")

	} else {
		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			return
		}
	}
}

func handle_ctl_msgs(active chan bool) {
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "handle_ctl_msgs", "Scanln failed : "+err.Error())
			return
		}
		msg_type := protocol.Findval(msg, "type", "server")
		msg_val := protocol.Findval(msg, "value", "server")
		switch msg_type {
		case "section_critique": // message sur la section critique
			if msg_val == "true" {
				in_section_critique = true
				ws_send("info=debut section critique")
				active <- true
				if pendingOpe != "" { //si j'ai la section critique
					modify_data(pendingOpe, active)
					liberer_sc(pendingOpe) // je la libère
					pendingOpe = ""
				}
			} else {
				in_section_critique = false
				ws_send("info=fin section critique")
			}
		case "data": // message sur l'update des données
			lastOpe = msg_val
			ws_send("data=" + msg_val) // update les données dans le whiteboard
			modify_data(msg_val, active)
		}

	}
}

/* demande et attend la section critique */
func wait_for_sc(active <-chan bool) bool {
	demander_sc()
	if <-active { // on attend que quelqu'un envoi que la section critique est active
		// display.Info("", "wait_for_sc", "active channel read, returning True")
		return true
	}
	return false
}

func modify_data(newOpe string, active chan bool) {
	op := protocol.Findval(newOpe, "op", "server")
	id := protocol.Findval(newOpe, "id", "server")
	switch op {
	case "create":
		parseShape, err := shape.ParseShape(newOpe)
		if err != nil {
			return
		}
		whiteboard.AddShape(id, parseShape)
	case "update":
		updates := shape.GetUpdateFields(newOpe)
		for key, value := range updates {
			whiteboard.UpdateShape(id, key, value)
		}
	case "delete":
		whiteboard.RemoveShape(id)
	case "clear":
		whiteboard.Clear()
	default:
		display.Warning("server", "modify", "OP inconnue : "+op)
	}
}

func main() {
	whiteboard = shape.Empty_board()
	var active = make(chan bool)

	port = flag.String("port", "4444", "n° de port")
	addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	display.Info("app", "main", "Démarrage du serveur "+*port)

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { do_websocket(w, r, active) })
	go handle_ctl_msgs(active)

	err := http.ListenAndServe(*addr+":"+*port, nil)
	if err != nil {
		return
	}
}
