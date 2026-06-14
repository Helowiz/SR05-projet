package main

import (
	"net"
	"os"
)

// ---------------------- Types et constantes ------------------

// Types de messages
const DEMANDE = "demande"
const ADMIS = "admis"
const BLEU = "bleu"
const ROUGE = "rouge"
const VAGUE = "vague"
const NEW_MEMBER = "new_member"
const LEAVE = "leave_msg"
const CTL_MESSAGE = "ctlmessage"

// intervalle de message
type Interval struct {
	debut int
	fin   int
}

// Types d'evennements
type Event interface {
	EventType() string
}
type MessageEvent struct {
	Adress  *net.UDPAddr
	Content string
}

type CtlMessageEvent struct {
	Content string
}

func (e MessageEvent) EventType() string    { return "message" }
func (e CtlMessageEvent) EventType() string { return "ctlmessage" }

// ------------------ variables globales ------------------
var G = struct {
	Dev_mode bool

	IntervallesMsg    map[string][]Interval
	CurrentMsgNum     int
	ConnectionMap     map[string]*net.UDPAddr
	Parent            string
	Neighbors         map[string]*net.UDPAddr
	NbVoisinsAttendus int
	NbSites           int

	Elu string

	Id    string
	Admis bool
	Pid   int

	Broadcast_adr    *net.UDPAddr
	PendingAjout     *net.UDPAddr
	PendingSelfLeave bool
	SocketBroadcast  *net.UDPConn // socket pour les broadcast
	SocketDirect     *net.UDPConn // socket pour les comm directes
}{
	Dev_mode: true,

	IntervallesMsg: make(map[string][]Interval, 100),
	ConnectionMap:  make(map[string]*net.UDPAddr, 100), // pour savoir ou envoyer les messages
	Neighbors:      make(map[string]*net.UDPAddr, 100), // voisins logiques pour les algos

	NbSites: 1,
	Pid:     os.Getpid(),

	PendingAjout:     nil,
	PendingSelfLeave: false,
}
