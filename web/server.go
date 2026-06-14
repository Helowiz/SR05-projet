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
	snap, err := snapshot.ToString(snapshot.Shot(whiteboard, *id)) // prend la snapshot et la mets en String pour l'envoyer
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
	msg := protocol.Msg_format_Ctrl("type", msgType) + protocol.Msg_format_Ctrl("snap", snap)
	//display.Info("SNAP", "snapshot", msg)
	fmt.Println(msg)
}

// func reloadSnapshot() {
// 	msg := protocol.Msg_format_Ctrl("type", "reload")
// 	fmt.Println(msg)
// }

func demander_sc() {
	msg := protocol.Msg_format_Ctrl("type", "fromapp_debut_sc")
	fmt.Println(msg)
}

func liberer_sc(newOpe string) {
	msg := protocol.Msg_format_Ctrl("type", "fromapp_fin_sc") + protocol.Msg_format_Ctrl("data", newOpe)
	fmt.Println(msg)
}

func wanna_leave() {
	msg := protocol.Msg_format_Ctrl("type", "fromapp_wanna_leave")
	fmt.Println(msg)
}

func demande_admission() {
	msg := protocol.Msg_format_Ctrl("type", "fromapp_demande_admission")
	fmt.Println(msg)
}

func do_webserver(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/client.html" {
		http.FileServer(http.Dir("../web/")).ServeHTTP(w, r)
		return
	}
	content, err := os.ReadFile("../web/client.html")
	if err != nil {
		http.Error(w, "Impossible de lire le fichier HTML", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(content)
}

/* Gere un message de la websocket */
func handle_ws_msg(message string) {
	//fmt.Println("réception : " + string(message))
	//display.Info("", "handle_ws_msg", "received : "+string(message))
	entries := strings.Split(message, "/")[1:] // split par / et ignore la première entrée vide
	prefix, suffix := protocol.ParseEntry(entries[0], "handle_ws_msg")
	//display.Info("", "handle_ws_msg", "received : "+string(message))

	switch prefix {
	case "init":
		ws_send("data=" + suffix)

	case "section_critique":
		if suffix == "activate" {
			demander_sc()
		}
		if suffix == "deactivate" {
			if len(entries) < 2 {
				display.Warning("", "handle_ws_msg", "Message pour libération section critique mal formaté : "+message)
				return
			}
			prefix, suffix = protocol.ParseEntry(entries[1], "handle_ws_msg")
			if prefix != "data" {
				display.Warning("", "handle_ws_msg", "Message pour libération section critique mal formaté : "+message)
				return
			}
			liberer_sc(suffix)
		}
	case "data":
		if suffix != "" {
			pendingOpe = suffix // en attente
			demander_sc()
			//wait_for_sc(active)
		}

	case "snapshot":
		initiate = true
		doLocalSnapshot()

	// case "reload":
	// 	reloadSnapshot()

	case "wanna_leave":
		if suffix == "true" {
			wanna_leave()
		} else {
			display.Warning("", "handle_ws_msg", "Message pour quitter le réseau mal formaté : "+message)
		}
	case "demande_admission":
		if suffix == "true" {
			demande_admission()
		} else {
			display.Warning("", "handle_ws_msg", "Message pour broadcast demande mal formaté : "+message)
		}
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
	for _, ope := range whiteboard.ToOperations() {
		eventQueue <- Event{from: WEBSOCKET, content: "/=init=" + ope}
	}

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
func handle_ctl_msg(msg string) {
	msg_type := protocol.Findval(msg, "type")
	msg_val := protocol.Findval(msg, "value")
	switch msg_type {
	case protocol.LEAVE: // message de validation de départ (je peux quitter le réseau)
		if msg_val == "true" {
			display.Info("SERVER :"+strconv.Itoa(os.Getpid()), "handle_ctl_msg()", "Départ validé par le contrôleur, je quitte le réseau")
			ws_send(protocol.LEAVE + "=true")
			whiteboard.Clear()
		} else {
			display.Warning("SERVER :"+strconv.Itoa(os.Getpid()), "handle_ctl_msg()", "Départ refusé par le contrôleur, je reste dans le réseau")
		}
	case protocol.ADMIS: // message de validation d'admission (je peux être admis)
		if msg_val == "true" {
			display.Info("SERVER :"+strconv.Itoa(os.Getpid()), "handle_ctl_msg()", "Admission validée par le contrôleur, je suis admis")
			ws_send(protocol.ADMIS + "=true")
		} else {
			display.Warning("SERVER :"+strconv.Itoa(os.Getpid()), "handle_ctl_msg()", "Admission refusée par le contrôleur, je ne suis pas admis")
			ws_send(protocol.ADMIS + "=false")
		}
	case protocol.NEW_MEMBER: // message de notification d'un nouveau membre
		// sending back to ctl the actual wb state
		msg := protocol.Msg_format_Ctrl("type", "wb_ops") + protocol.Msg_format_Ctrl("new_member_id", msg_val)
		ops := whiteboard.ToOperations()
		n_ops := len(ops)
		msg += protocol.Msg_format_Ctrl("n_ops", strconv.Itoa(n_ops))
		i := 0
		for _, ope := range ops {
			msg += protocol.Msg_format_Ctrl("op"+strconv.Itoa(i), ope)
			i++
		}
		fmt.Println(msg)
	case "wb_ops": // message de mon controleur contenant les opérations du whiteboard pour le mettre à jour parce que je viens d'être admis
		n_ops_str := protocol.Findval(msg, "n_ops")
		n_ops, err := strconv.Atoi(n_ops_str)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "handle_ctl_msg()", "Erreur de conversion n_ops : "+err.Error())
			return
		}
		for i := 0; i < n_ops; i++ {
			ope := protocol.Findval(msg, "op"+strconv.Itoa(i))
			modify_data(ope)
			ws_send("data=" + ope)
		}
	case "section_critique": // message sur la section critique
		switch msg_val {
		case "true":
			in_section_critique = true
			ws_send("section_critique=debut_sc")
		case "false":
			in_section_critique = false
			ws_send("section_critique=fin_sc")
		case "other":
			ws_send("section_critique=other")
		}
	case "data": // message sur l'update des données
		lastOpe = msg_val
		ws_send("data=" + msg_val)
		if msg_val != "" {
			modify_data(msg_val) // update les données dans le whiteboard
		}
	case "snapshot_app":
		doLocalSnapshot()
	case "reload":
		// extraire la global_state
		// envoyer à l'interface sous forme de suite d'opération
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

func modify_data(newOpe string) {
	op := protocol.Findval(newOpe, "op")
	idSite := protocol.Findval(newOpe, "id")
	switch op {
	case "lock":
		// no operation
	case "create":
		parseShape, err := shape.ParseShape(newOpe)
		if err != nil {
			return
		}
		whiteboard.AddShape(idSite, parseShape)
	case "update":
		updates := shape.GetUpdateFields(newOpe)
		for key, value := range updates {
			whiteboard.UpdateShape(idSite, key, value)
		}
	case "delete":
		whiteboard.RemoveShape(idSite)
	case "clear":
		whiteboard.Clear()
	default:
		display.Warning("server", "modify", "OP inconnue : "+op)
	}
}

func main() {
	whiteboard = shape.Empty_board()
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
				handle_ctl_msg(event.content)
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
