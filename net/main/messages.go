package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"fmt"
	"strconv"
)

// ======================= Communication avec le controleur  =======================

func sendToCTL(msg string) {
	fmt.Println(msg)
}
func sendToCrlfromNET(msg string) {
	display.Envoie(G.Id, "MAIN_NET:sendtocrtlNet", "contenu envoyé à mon controleur de la part de net : "+msg)
	msg += protocol.Msg_format_Ctrl("msg", "net")
	fmt.Println(msg)
}

/*
	is_ctl_message

Retourne vrai si c'est un message d'un controlleur faux sinon (msg app)
*/
func is_ctl_message(msg string) bool {
	sHrcv := protocol.Findval(msg, "hlg") // les message de control on une champ hlg
	if sHrcv != "" {
		return true
	}
	return false
}

// ======================= Gestion de l'envoi / du routage des messages  =======================

/*
	send_to_neigh

Envoi le message a tout les voisins sauf celui passe en exclude
// TODO - Voir comment faire pour pas envoyer un par un ???
*/
func send_to_neigh(msg string, exclude string) {
	for idneigh, adr := range G.Neighbors {
		formatted_msg := protocol.Msg_format("from", G.Id) + protocol.Msg_format("to", idneigh) + protocol.Msg_format("num_msg", strconv.Itoa(G.CurrentMsgNum)) + msg

		_, err := G.SocketDirect.WriteTo([]byte(formatted_msg), adr)
		if err != nil {
			display.Error(G.Id, "send_to_neigh", "WriteTo failed")
			return
		}
	}

	G.CurrentMsgNum++
}

/*
	broadcast

Envoi a tous les sites
*/
func broadcast(msg string) {
	formatted_msg := protocol.Msg_format("to", "all") + protocol.Msg_format("from", G.Id) + protocol.Msg_format("num_msg", strconv.Itoa(G.CurrentMsgNum)) + msg

	_, err := G.SocketDirect.WriteTo([]byte(formatted_msg), G.Broadcast_adr)

	if err != nil {
		display.Error(G.Id, "broadcast", "WriteTo failed : "+err.Error())
		return
	}
	G.CurrentMsgNum++
}

/*
	broadcast_demande

Broadcast une demande d'admission a tous
Pas de numéro de série car nous sommes pas encore admis
*/
func broadcast_demande() {
	formatted_msg := protocol.Msg_format("to", "all") + protocol.Msg_format("from", G.Id) + protocol.Msg_format("msg", DEMANDE)

	_, err := G.SocketDirect.WriteTo([]byte(formatted_msg), G.Broadcast_adr)

	if err != nil {
		display.Error(G.Id, "broadcast", "WriteTo failed : "+err.Error())
		return
	}
}

/*
	send

Envoi un message a un site ciblé
- si dans voisins envoi direct
- sinon broadcast mais avec un id "to" qui cible
*/

func send(msg string, to string) {
	formatted_msg := protocol.Msg_format("from", G.Id) + protocol.Msg_format("to", to) + protocol.Msg_format("num_msg", strconv.Itoa(G.CurrentMsgNum)) + msg

	if adr, ok := G.ConnectionMap[to]; ok { // si je connais son adresse direct
		_, err := G.SocketDirect.WriteTo([]byte(formatted_msg), adr)
		if err != nil {
			display.Error(G.Id, "send", "display.Error writing to conn in send()")
		}
	} else { // sinon je le broadcast a tous
		G.SocketDirect.WriteTo([]byte(formatted_msg), G.Broadcast_adr)
	}
	G.CurrentMsgNum++
}

//======================= Gestion des intervalles des messages =======================

func decalerGauche(l []Interval, index int) []Interval {
	for i := index; i < (len(l) - 1); i++ {
		l[i] = l[i+1]
	}
	return l[:len(l)-1]
}

func decalerDroite(l []Interval, index int) []Interval {
	l = append(l, l[len(l)-1]) // extension de 1
	for i := len(l) - 2; i > index; i-- {
		l[i] = l[i-1]

	}
	return l
}

/* Update la map des intervalles, retourne vrai le message est nouveau faux deja present */
func UpdateInterval(interval_map map[string][]Interval, id_site string, num int) bool {

	if _, ok := interval_map[id_site]; !ok { // 1er msg de ce site
		interval_map[id_site] = make([]Interval, 0)
		interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
		return true
	}
	for index, inter := range interval_map[id_site] {

		if num >= inter.debut && num <= inter.fin { // dans un interval, deja recu
			return false
		}

		// extension a gauche
		if inter.debut-1 == num {
			interval_map[id_site][index] = Interval{inter.debut - 1, inter.fin}
			return true

		}

		// extension a droite tout vas bien
		if inter.fin+1 == num {
			interval_map[id_site][index] = Interval{inter.debut, inter.fin + 1}

			// merge avc prochain ?
			if index+1 < len(interval_map[id_site]) {
				newInterval, ok := mergeIntervals(interval_map[id_site][index], interval_map[id_site][index+1])
				if ok { // merge reussi
					interval_map[id_site][index] = newInterval
					interval_map[id_site] = decalerGauche(interval_map[id_site], index+1)
				}

			}
			return true
		}

		if num < inter.debut { // insertion entre deux intervalles
			interval_map[id_site] = decalerDroite(interval_map[id_site], index)
			interval_map[id_site][index] = Interval{num, num}
			return true
		}
	}
	// arrive en fin de liste
	interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
	return true

}

/*
	verifie si c'est possible de joindre deux intervalles et le fait si c'est possible

args : deux intervales
sortie : (intervale merged, true) si joignable, (intervalle nulle, false) sinon
*/
func mergeIntervals(a Interval, b Interval) (Interval, bool) {

	if (a.debut < b.debut && a.fin < b.debut-1) || (a.debut > b.debut && a.debut-1 > b.fin) {
		fmt.Printf("\"Not mergeable\": %v\n", "Not mergeable")
		return Interval{0, 0}, false
	} else {
		return Interval{min(a.debut, b.debut), max(a.fin, b.fin)}, true
	}
}
