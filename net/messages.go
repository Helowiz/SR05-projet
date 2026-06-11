package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"strconv"
	"strings"
)

const DEMANDE = "demande"
const ANNONCE = "annonce" // Pour annoncer au démarrage que je suis le voisin d'un site
const ADMIS = "admis"
const BLEU = "bleu"
const ROUGE = "rouge"
const VAGUE = "vague"
const NEW_MEMBER = "new_member"

var (
	Red    = "\033[1;31m"
	Orange = "\033[1;33m"
	Green  = "\033[1;32m"
	Reset  = "\033[0;00m"
	Purple = "\033[1;35m"
	Cyan   = "\033[1;36m"
)

func simpleIdShowing(id string) string {
	if len(id) < 4 {
		return id
	}
	return id[0:4]
}

func baseDisplay(color string, level string, name string, where string, what string) {
	stderr.Printf("%s* [%-4.4s %d] %-8.8s [%-7.7s] : %s%s\n", color, simpleIdShowing(name), pid, where, level, what, Reset)
}

func Vague(name string, where string, what string) {
	baseDisplay(Cyan, "VAGUE", name, where, what)
}

func Info(name string, where string, what string) {
	baseDisplay(Green, "DISPLAY", name, where, what)
}

func Warning(name string, where string, what string) {
	baseDisplay(Orange, "WARNING", name, where, what)
}

func Error(name string, where string, what string) {
	baseDisplay(Red, "ERROR", name, where, what)
}

func Findval(msg string, key string, name string) string {
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		equ := keyval[0:1]

		tabkeyval := strings.Split(keyval[1:], equ)
		if tabkeyval[0] == key && len(tabkeyval) == 2 {
			return tabkeyval[1]
		}

		if len(msg) < 4 {
			display.Warning(name, "FindVal", "Message trop court ou mal formaté : "+msg)
			return ""
		}
	}
	return ""
}

// ======================= Gestion de l'envoi / du routage des messages

/*
	send_to_neigh

Envoi le message a tout les voisins sauf celui passe en exclude
*/
func send_to_neigh(msg string, exclude string) {
	msg = protocol.Msg_format_NET("from", id) + protocol.Msg_format_NET("num_msg", strconv.Itoa(current_msg_num)) + msg

	for id_neigh, _ := range connectionMap {
		if id_neigh != exclude {
			send(msg, id_neigh)
		}
	}

	current_msg_num++
}

/*
	broadcast

Envoi le message a tout les voisins en leur disant de l'envoyer a leur tour aux voisins
*/
func broadcast(msg string) {

	msg = protocol.Msg_format_NET("to", "all") + protocol.Msg_format_NET("from", id) + protocol.Msg_format_NET("num_msg", strconv.Itoa(current_msg_num)) + msg
	send_to_neigh(msg, "")
	current_msg_num++
}

/*
	send

Envoi un message a un site ciblé
*/

func send(msg string, to string) {
	msg = protocol.Msg_format_NET("from", id) + protocol.Msg_format_NET("to", to) + protocol.Msg_format_NET("num_msg", strconv.Itoa(current_msg_num)) + msg
	// si c'est pour mon voisin, j'envoi que a lui
	if conn, ok := connectionMap[to]; ok {
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			Error(id, "send", "Error writing to conn in send()")
		}
	} else { // sinon j'envoi a tous mes voisins
		send_to_neigh(msg, "")
	}
	current_msg_num++
}

/*
forward
Transmet un message a tous les voisins tel qu'il est reçu
*/
func forward(msg string, sender string) {
	send_to_neigh(msg, sender)
}
