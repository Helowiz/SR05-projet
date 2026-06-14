package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"maps"
	"slices"
	"strconv"
	"strings"
)

/*
getVoisinsList
Retourne une liste des voisins a partir de la map globale
*/
func getVoisinsList() []string {
	voisinsList := slices.Collect(maps.Keys(G.Neighbors))
	for i, v := range voisinsList {
		voisinsList[i] = display.SimpleIdShowing(v)
	}
	return voisinsList
}

/*
debutVagueElection
Commence une vague pour élir un site
*/
func debutVagueElection() {
	if G.Parent == "" { // pas deja atteint => peut commencer election
		G.Parent = G.Id
		G.Elu = G.Id
		G.NbVoisinsAttendus = len(G.Neighbors)
		display.Vague(G.Id, "debutDiffusionVague", "DEBUT DE LA VAGUE")
		msg := protocol.Msg_format("msg", BLEU) + protocol.Msg_format("elu", G.Id) //envoie aux voisins
		send_to_neigh(msg, "")

	}
}

func recptionMsgBleu(from string, elu_recu string) {
	display.Vague(G.Id, "recptionMsgBleu", "from : "+display.SimpleIdShowing(from)+" mon_parent : "+display.SimpleIdShowing(G.Parent))
	msg := ""
	if G.Parent == "" || elu_recu < G.Elu { // Première vague ou G.Id de l’élu est plus petite que la précédente
		display.Vague(G.Id, "receptionMsgBleu", "Perdu")
		G.Parent = from
		G.Elu = elu_recu
		G.NbVoisinsAttendus = len(G.Neighbors) - 1
		if G.NbVoisinsAttendus > 0 {
			msg = protocol.Msg_format("msg", BLEU) + protocol.Msg_format("elu", elu_recu)
			display.Vague(G.Id, "recsBleu", "NB VOISIN ATTENDU: "+strconv.Itoa(G.NbVoisinsAttendus)+" Mes voisins : "+strings.Join(getVoisinsList(), ", ")+" MESSAGE"+msg)

			send_to_neigh(msg, G.Parent)
		} else {
			msg = protocol.Msg_format("msg", ROUGE) + protocol.Msg_format("elu", elu_recu)
			send(msg, G.Parent)
			display.Vague(G.Id, "recBleu", "NB VOISIN ATTENDU: "+strconv.Itoa(G.NbVoisinsAttendus)+" Mes voisins : "+strings.Join(getVoisinsList(), ", ")+" MESSAGE"+msg)
		}
	} else {

		if G.Elu == elu_recu { // msg de la mm vague, => remontée vers parent
			msg = protocol.Msg_format("msg", ROUGE) + protocol.Msg_format("elu", G.Elu)
			send(msg, G.Parent)
			display.Vague(G.Id, "recBleu", "NB VOISIN ATTENDU: "+strconv.Itoa(G.NbVoisinsAttendus)+" MESSAGE"+msg)
		}

	}

}

func receptionMsgRouge(from string, elu_recu string, eventFile chan<- Event) {
	display.Vague(G.Id, "recRouge", "from : "+from+" mon elu : "+G.Elu+" elurecu : "+elu_recu[:10])
	if elu_recu == G.Elu { // on accept que la vague courante

		G.NbVoisinsAttendus--
		display.Vague(G.Id, "recRouge", "C'est ma vague nbVoisins attendus : "+strconv.Itoa(G.NbVoisinsAttendus))
		if G.NbVoisinsAttendus == 0 {
			if G.Elu == G.Id { // on a gagne l'election
				display.Vague(G.Id, "receptionMsgRouge", "FIN JE SUIS ELU !!!")
				G.Parent = ""
				G.Elu = ""

				if G.PendingSelfLeave {
					handleSelfLeave(eventFile)
				} else if G.PendingAjout != nil {
					handleAdmit(eventFile)
				} else {
					display.Error(G.Id, "receptionMsgRouge", "Je suis elu mais je n'ai pas de site a admettre ni de leave a faire")
				}

			} else {
				msg := protocol.Msg_format("msg", ROUGE) + protocol.Msg_format("elu", G.Elu)
				display.Vague(G.Id, "recRouge", "NB VOISIN ATTENDU: "+strconv.Itoa(G.NbVoisinsAttendus)+"MESSAGE"+msg)
				send(msg, G.Parent)
			}
		}
	}
}
