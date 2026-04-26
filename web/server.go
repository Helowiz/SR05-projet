package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var ws *websocket.Conn = nil

var data int
var in_section_critique bool = false

func demander_sc() {
	msg := protocol.Msg_format("type", "fromapp_debut_sc")
	fmt.Println(msg)
}

func liberer_sc(newData string) {
	msg := protocol.Msg_format("type", "fromapp_fin_sc") + protocol.Msg_format("data", newData)
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
	ws_send("data=" + strconv.Itoa(data))

	for {
		_, message, err := cnx.ReadMessage()
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_websocket()", "cnx.ReadMessage() failed : "+err.Error())
			return
		}
		//fmt.Println("réception : " + string(message))
		parts := strings.Split(string(message), "=")
		prefix, suffix := parts[0], parts[1]
		display.Info("", "do_websocket", "received : "+string(message))

		switch string(prefix) {
		case "section_critique": // juste pour le debug, a enlever ce cas apres
			if suffix == "activate" {
				demander_sc()
			}
			if suffix == "deactivate" {
				liberer_sc("0")
			}

		case "data":
			newData, err := strconv.Atoi(suffix)
			if err != nil {
				display.Error("", "do_websocket", "data could not be converted")
				return
			}
			modify_data(newData, active)
			//display.Info("", "do_websocket", "new data : "+strconv.Itoa(newData))

		}
	}
}

func ws_send(msg string) {
	if ws == nil {
		display.Warning("SERVER :"+strconv.Itoa(os.Getpid()), "ws_send()", "Websocket non ouverte")

	} else {
		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			//fmt.Println("ws_send", "WriteMessage : "+string(err.Error()))
		} else {
			//fmt.Println("ws_send", "sending "+msg)
		}
	}
}

func handle_ctl_msgs(active chan<- bool) {
	var msg string
	for {
		_, err := fmt.Scanln(&msg)
		if err != nil {
			display.Error("SERVER :"+strconv.Itoa(os.Getpid()), "do_send()", "Scanln failed : "+err.Error())
			return
		}
		parts := strings.Split(msg, "=")
		prefix, suffix := parts[0], parts[1]

		switch prefix {
		case "section_critique": // message sur la section critique
			if suffix == "true" {
				in_section_critique = true
				ws_send("info=debut section critique")
				active <- true
			} else {
				in_section_critique = false
				ws_send("info=fin section critique")
			}

		case "data": // message sur l'update des données
			newData, err := strconv.Atoi(suffix)
			if err != nil {
				display.Error("", "handle_ctl_msgs", "data could not be converted")
				return
			}
			data = newData
			// indique a l'interface d'update les donnes
			ws_send("data=" + strconv.Itoa(newData))
		}

	}
}

/* demande et attend la section critique */
func wait_for_sc(active <-chan bool) bool {
	demander_sc()
	if <-active { // on attend que quelqu'un envoi que la section critique est active
		display.Info("", "wait_for_sc", "active channel read, returning True")
		return true
	}
	return false
}

func modify_data(newData int, active <-chan bool) {
	wait_for_sc(active)
	//
	// give back section critique
	liberer_sc(strconv.Itoa(newData))

}

func main() {
	data = 0
	var active chan bool = make(chan bool)

	var port = flag.String("port", "4444", "n° de port")
	var addr = flag.String("addr", "localhost", "nom/adresse machine")

	flag.Parse()

	http.HandleFunc("/", do_webserver)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { do_websocket(w, r, active) })
	go handle_ctl_msgs(active)
	http.ListenAndServe(*addr+":"+*port, nil)
}
