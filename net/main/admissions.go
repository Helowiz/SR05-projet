package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"net"
	"strconv"
)

func handleDemande(adresse *net.UDPAddr, eventQueue chan<- Event) {
	display.Info(G.Id, "handleDemande", "handling demande, port recu : "+strconv.Itoa(adresse.Port))
	G.PendingAjout = adresse
	if len(G.Neighbors) == 0 { // si je suis tout seul j'admet direct
		display.Info(G.Id, "handleClient", "C'est le premier donc ok direct")
		handleAdmit(eventQueue)
	} else { // sinon election
		display.Info(G.Id, "handleClient", "Election pour voir si je peux l'admettre")
		debutVagueElection()
	}

}

func handleAdmit(eventQueue chan<- Event) {
	display.Info(G.Id, "handleAdmit", "J'admet le site qui demande")
	new_id := makeUniqueId()
	G.NbSites += 1
	// TODO - mettre ca dans une fonction speciale ?
	tosend := protocol.Msg_format("num_msg", strconv.Itoa(G.CurrentMsgNum)) + protocol.Msg_format("from", G.Id) + protocol.Msg_format("msg", ADMIS) + protocol.Msg_format("your_id", new_id) + protocol.Msg_format("nb_sites", strconv.Itoa(G.NbSites))
	G.CurrentMsgNum++

	_, err := G.SocketDirect.WriteTo([]byte(tosend), G.PendingAjout)

	if err != nil {
		display.Error(G.Id, "handleAdmit", "display.Error writting to connection :"+err.Error())
	}
	G.Neighbors[new_id] = G.PendingAjout

	broadcast(protocol.Msg_format("msg", NEW_MEMBER) + protocol.Msg_format("his_id", new_id))

	sendToCrlfromNET(protocol.Msg_format_Ctrl("type", NEW_MEMBER) + protocol.Msg_format_Ctrl("nb_sites", strconv.Itoa(G.NbSites)) + protocol.Msg_format_Ctrl("i_am_the_parent", "true") + protocol.Msg_format_Ctrl("new_member_id", new_id)) // previens mon controleur

}

/*
handleNewMember

- Gere le signal qu'un nouveau member a ete ajouté
- Sert aussi a reset les variables sachant on a perdu l'election si on participait
*/
func handleNewMember(new_member_id string) {
	if new_member_id == G.Id {
		// c'est moi qui ai été admis, je ne fais rien.
		// A noter qu'il arrive que je recoive ce message après avoir été admis,
		// ce qui peut poser problème si une autre vague a été lancée entre temps,
		// c'est pour cela que je ne fais rien ici.
		return
	}
	if G.Parent != "" || G.Elu != "" { // on etait dans une election, on la perd
		G.Parent = ""
		G.Elu = ""
		G.PendingAjout = nil
		G.PendingSelfLeave = false
		G.NbSites += 1
		sendToCrlfromNET(protocol.Msg_format_Ctrl("type", NEW_MEMBER) + protocol.Msg_format_Ctrl("nb_sites", strconv.Itoa(G.NbSites)) + protocol.Msg_format_Ctrl("i_am_the_parent", "false") + protocol.Msg_format_Ctrl("new_member_id", new_member_id)) // previens mon controleur

	} else {
		display.Warning(G.Id, "handleNewMember", "J'ai reçu un signal de nouvel admis alors que je n'étais pas dans une election, je l'ignore")
	}
}
