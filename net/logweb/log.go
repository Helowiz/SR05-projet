package logweb

import (
	"SR05_projet/protocol"
	"fmt"
)

const LOG = "log"
const ADMET = "ADMET"
const ADMIS = "ADMIS"
const DEMANDE = "DEMANDE"
const NEW_NEIGHBOR = "VOISINS"
const ADD_MEMBER = "MEMBERS"
const VAGUE = "VAGUE"
const RECOIS_DEMANDE = "RECOIS_DEMANDE"

func sendToLog(msg string) {
	fmt.Println(protocol.Msg_format_Ctrl("to", LOG) + msg)
}

func Vague(id string, from string, to string, parent string, elu string, color string, nbvoisinsAttendus string) {
	msg := protocol.Msg_format_Ctrl("info", VAGUE) +
		protocol.Msg_format_Ctrl("id", id) +
		protocol.Msg_format_Ctrl("from", from) +
		protocol.Msg_format_Ctrl("to", to) +
		protocol.Msg_format_Ctrl("parent", parent) +
		protocol.Msg_format_Ctrl("elu", elu) +
		protocol.Msg_format_Ctrl("color", color) +
		protocol.Msg_format_Ctrl("nbvoisinsAttendus", nbvoisinsAttendus)
	sendToLog(msg)
}

func Admis(id string, adress string, port string) {
	msg := protocol.Msg_format_Ctrl("info", ADMIS) +
		protocol.Msg_format_Ctrl("id", id) +
		protocol.Msg_format_Ctrl("address", adress) +
		protocol.Msg_format_Ctrl("port", port)
	sendToLog(msg)
}

func DemandeAdmission(id string) {
	msg := protocol.Msg_format_Ctrl("info", RECOIS_DEMANDE) +
		protocol.Msg_format_Ctrl("id", id)
	sendToLog(msg)
}

func Admet(id string, id_new string) {
	msg := protocol.Msg_format_Ctrl("info", ADMET) +
		protocol.Msg_format_Ctrl("id", id) +
		protocol.Msg_format_Ctrl("id_new", id_new)
	sendToLog(msg)
}

func Annonce(id string, neighbor_id string) {
	msg := protocol.Msg_format_Ctrl("info", NEW_NEIGHBOR) +
		protocol.Msg_format_Ctrl("id", id) +
		protocol.Msg_format_Ctrl("neighbor_id", neighbor_id)
	sendToLog(msg)
}

func Demande() {
	msg := protocol.Msg_format_Ctrl("info", DEMANDE)
	sendToLog(msg)
}
