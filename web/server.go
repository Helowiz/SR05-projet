package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"SR05_projet/shape"
	"SR05_projet/snapshot"
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

var pendingOpe = ""
var port *string
var addr *string
var id *string

var initiate = false

const WEBSOCKET string = "websocket"
const CONTROL string = "control"

type Event struct {
	from    string
	content string
}

func doLocalSnapshot() {
	snap, err := snapshot.SnapshotToString(snapshot.Shot(whiteboard, *id)) // prend la snapshot et la mets en String pour l'envoyer
	if err != nil {
		display.Error("snapshot", "doLocalSnapshot", err.Error())
		return
	}

	// envoie la snapshot
	msgType := "snapshot"
	if initiate {
		initiate = false // reset pour prochaine snapshot
		msgType = "snapshot_init"
	}
	msg := protocol.Msg_format("type", msgType) + protocol.Msg_format("snap", snap)
	display.Info("SNAP", "snapshot", msg)
	fmt.Println(msg)
}

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

/* Gere un message de la websocket */
func handle_ws_msg(message string) {
	//fmt.Println("réception : " + string(message))
	parts := strings.Split(message, "=")
	prefix, suffix := parts[0], strings.TrimSpace(parts[1])
	display.Info("", "handle_ws_msg", "received : "+string(message))

	switch prefix {
	case "init":
		ws_send("data=" + lastOpe)

	case "section_critique": // juste pour le debug, TODO - a enlever ce cas apres
		if suffix == "activate" {
			demander_sc()
		}
		if suffix == "deactivate" {
			liberer_sc("0")
		}
	case "data":
		pendingOpe = suffix // en attente
		demander_sc()
		//wait_for_sc(active)

	case "snapshot":
		initiate = true
		doLocalSnapshot()
	}
}

/* Etabli et gere le websocket, push messages sur eventQueue*/
func do_websocket(w http.ResponseWriter, r *http.Request, eventQueue chan<- Event) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	cnx, err := upgrader.Upgrade(w, r, nil)
	ws = cnx
	if err != nil {
		display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "upgrade failed : "+err.Error())

		return
	}
	eventQueue <- Event{from: WEBSOCKET, content: "init="}

	for {
		_, message, err := cnx.ReadMessage()
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "cnx.ReadMessage() failed : "+err.Error())
			return
		}

		eventQueue <- Event{from: WEBSOCKET, content: string(message)}

	}
}

/* Envoi au websocket */
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

/* Gere message de control */
func handle_ctl_msg(msg string, active chan bool) {
	msg_type := protocol.Findval(msg, "type", "server")
	msg_val := protocol.Findval(msg, "value", "server")
	switch msg_type {
	case "section_critique": // message sur la section critique
		if msg_val == "true" {
			in_section_critique = true
			ws_send("info=debut section critique")
			//active <- true
			if pendingOpe != "" { //si j'ai la section critique
				if pendingOpe != "lock" {
					//modify_data(pendingOpe, active)
					liberer_sc(pendingOpe) // je la libère
				}
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

	case "snapshot":
		doLocalSnapshot()
	}
}

/* Envoi messages de control dans eventqueue */
func listen_for_ctl_msg(eventQueue chan<- Event) {
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "listen_for_clt_msg", "Scanln failed : "+err.Error())
			return
		}
		eventQueue <- Event{from: CONTROL, content: msg}
	}
}

/* demande et attend la section critique */
func wait_for_sc(active <-chan bool) bool {
	demander_sc()
	if <-active {
		return true
	}
	return false
}

func modify_data(newOpe string, active chan bool) {
	op := protocol.Findval(newOpe, "op", "server")
	id := protocol.Findval(newOpe, "id", "server")
	switch op {
	case "lock":
		// no operation
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
	var eventQueue = make(chan Event, 100)

	port = flag.String("port", "4444", "n° de port")
	addr = flag.String("addr", "localhost", "nom/adresse machine")
	id = flag.String("id", "", "nom id")

	flag.Parse()

	display.Info("app", "main", "Démarrage du serveur "+*port)

	http.HandleFunc("/", do_webserver)

	// ecoute websocket et control
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { do_websocket(w, r, eventQueue) })
	go listen_for_ctl_msg(eventQueue)

	// loop gestion evenements
	go func() {
		for event := range eventQueue {
			switch event.from {
			case CONTROL:
				handle_ctl_msg(event.content, active)
			case WEBSOCKET:
				handle_ws_msg(event.content)
			}
		}
	}()

	err := http.ListenAndServe(*addr+":"+*port, nil)
	if err != nil {
		return
	}
}
