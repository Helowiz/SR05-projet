package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"strconv"
)

func main() {

	p_nom := flag.String("n", "controler", "nom")
	flag.Parse()

	display.Info(*p_nom, "main", "Démarrage du contrôleur...")

	var rcvmsg string
	var h = 0
	var hrcv int
	var sndmsg string

	for {
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error(*p_nom, "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			return
		}
		display.Info(*p_nom, "reception", "Reçu brut : "+rcvmsg)
		sHrcv := protocol.Findval(rcvmsg, "hlg", *p_nom)
		if sHrcv != "" {
			oldH := h
			hrcv, _ = strconv.Atoi(sHrcv)
			h = protocol.Recaler(h, hrcv)
			display.Info(*p_nom, "recalage", fmt.Sprintf("H_locale=%d, H_recue=%d -> Nouvelle H=%d", oldH, hrcv, h))
		} else {
			h = h + 1
			display.Info(*p_nom, "horloge", fmt.Sprintf("Pas de 'hlg', incrémentation locale -> H=%d", h))
		}

		sndmsg = protocol.Findval(rcvmsg, "msg", *p_nom)
		if sndmsg == "" {
			display.Info(*p_nom, "format", "Formatage du message complet avec horloge ajoutée")
			newMsg := protocol.Msg_format("msg", rcvmsg) + protocol.Msg_format("hlg", strconv.Itoa(h))
			fmt.Println(newMsg)
			display.Info(*p_nom, "emission", "Transmis : "+newMsg)
		} else {
			display.Info(*p_nom, "format", "Message extrait avec succès")
			fmt.Println(sndmsg)
			display.Info(*p_nom, "emission", "Transmis : "+sndmsg)
		}
	}
}
