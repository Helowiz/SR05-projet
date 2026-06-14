package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"maps"
	"net"
	"slices"
	"strconv"
)

func wannaLeave() {
	if len(G.Neighbors) == 0 {
		display.Warning(G.Id, "wannaLeave", "Je suis tout seul dans le réseau, je ne peux pas quitter")
	} else { // sinon election
		display.Info(G.Id, "wannaLeave", "Je souhaite quitter le réseau, je lance une vague pour savoir si je peux quitter")
		debutVagueElection()
		G.PendingSelfLeave = true
	}
}

// On supposera ici que j'ai au moins un voisin (que je ne suis pas tout seul dans le réseau), normalement
// cette vérification a été faite plus tôt.
// Dire à tous les voisins que je quitte le réseau. Donc :
// - mon parent coupera sa connexion avec moi
// - les autres voisins se connecteront directement à mon parent et couperont leur connexion avec moi
func handleSelfLeave(eventFile chan<- Event) {

	parent_to_connect_to_id := slices.Collect(maps.Keys(G.Neighbors))[0]
	adr, ok := G.Neighbors[parent_to_connect_to_id]
	if !ok {
		display.Warning(G.Id, "handleSelfLeave", "Je n'ai pas de connexion avec mon parent admis, je ne peux pas quitter proprement")
		return
	}
	private_port := strconv.Itoa(adr.Port)
	ip := adr.IP.String()
	display.Leave(G.Id, "handleSelfLeave", "Je quitte le réseau, j'informe mes voisins et je coupe mes connexions. Mon parent admis est "+display.SimpleIdShowing(parent_to_connect_to_id)+" à l'adresse "+ip+":"+private_port)

	sendToCrlfromNET(protocol.Msg_format_Ctrl("type", LEAVE))

	formatted_msg := protocol.Msg_format("num_msg", strconv.Itoa(G.CurrentMsgNum)) + protocol.Msg_format("from", G.Id) + protocol.Msg_format("msg", LEAVE) + protocol.Msg_format("parent_to_connect_to_id", parent_to_connect_to_id) + protocol.Msg_format("parent_admitted_IP", ip) + protocol.Msg_format("parent_admitted_private_port", private_port)
	G.CurrentMsgNum++
	n_children := len(G.Neighbors) - 1
	formatted_msg += protocol.Msg_format("nb_children", strconv.Itoa(n_children))
	i := 0
	for child_id, child_adr := range G.Neighbors {
		if child_id == parent_to_connect_to_id {
			continue
		}
		formatted_msg += protocol.Msg_format("child_"+strconv.Itoa(i)+"_id", child_id) + protocol.Msg_format("child_"+strconv.Itoa(i)+"_IP", child_adr.IP.String()) + protocol.Msg_format("child_"+strconv.Itoa(i)+"_private_port", strconv.Itoa(child_adr.Port))
		i++
	}
	_, err := G.SocketBroadcast.WriteTo([]byte(formatted_msg), G.Broadcast_adr)
	if err != nil {
		display.Error(G.Id, "broadcast", "WriteTo failed : "+err.Error())
		return
	}

	G.Admis = false
	G.Id = ""
	G.PendingSelfLeave = false
	G.Parent = ""
	G.Elu = ""
	G.NbVoisinsAttendus = 0
	G.Neighbors = make(map[string]*net.UDPAddr, 100)

	// On peut tenter de se réadmettre directement
	// go doPeriodique(eventFile)
}

func handleOtherLeave(adresse *net.UDPAddr, content string, eventFile chan<- Event) {
	leaving_id := protocol.Findval(content, "from")

	if leaving_id == G.Id {
		display.Leave(G.Id, "handleOtherLeave", "J'ai reçu mon propre message de leave")
		return
	}

	if G.Parent != "" || G.Elu != "" { // on etait dans une election, on la perd
		G.Parent = ""
		G.Elu = ""
		G.PendingSelfLeave = false
		G.PendingAjout = nil
		G.NbSites -= 1
		sendToCrlfromNET(protocol.Msg_format_Ctrl("type", "other_leave") + protocol.Msg_format_Ctrl("nb_sites", strconv.Itoa(G.NbSites)) + protocol.Msg_format_Ctrl("its_id", leaving_id))
	} else {
		display.Warning(G.Id, "handleOtherLeave", "J'ai reçu un signal de leave alors que je n'étais pas dans une élection, je l'ignore")
		return
	}

	if _, exists := G.Neighbors[leaving_id]; !exists {
		display.Leave(G.Id, "handleOtherLeave", "Message de leave d'un site qui n'est pas mon voisin : "+display.SimpleIdShowing(leaving_id)+", je l'ignore")
		return
	}

	display.Leave(G.Id, "handleOtherLeave", "Le site "+display.SimpleIdShowing(leaving_id)+" quitte le réseau, je coupe la connexion avec lui et je vérifie si je dois me reconnecter à son parent pour rester connecté au réseau")
	delete(G.Neighbors, leaving_id)

	parent_to_connect_to_id := protocol.Findval(content, "parent_to_connect_to_id")
	if parent_to_connect_to_id == G.Id {
		display.Leave(G.Id, "handleOtherLeave", "Je suis le parent admis du site qui quitte, j'initie les connexions avec ses enfants pour qu'ils restent connectés au réseau")
		n_children := protocol.Findval(content, "nb_children")
		n_children_int, err := strconv.Atoi(n_children)
		if err != nil {
			display.Error(G.Id, "handleOtherLeave", "Erreur lors de la conversion du nombre d'enfants : "+err.Error())
			return
		}
		for i := 0; i < n_children_int; i++ {
			child_id := protocol.Findval(content, "child_"+strconv.Itoa(i)+"_id")
			if _, exists := G.Neighbors[child_id]; exists {
				display.Leave(G.Id, "handleOtherLeave", "L'enfant "+display.SimpleIdShowing(child_id)+" est déjà dans mes voisins, je ne fais rien")
				continue
			}
			child_private_port := protocol.Findval(content, "child_"+strconv.Itoa(i)+"_private_port")
			child_ip := protocol.Findval(content, "child_"+strconv.Itoa(i)+"_IP")
			child_address, err := net.ResolveUDPAddr("udp4", child_ip+":"+child_private_port)
			if err != nil {
				display.Error(G.Id, "handleOtherLeave", "failed to resolve private adress for child "+display.SimpleIdShowing(child_id)+", error: "+err.Error())
				continue
			}
			G.Neighbors[child_id] = child_address
			display.Leave(G.Id, "handleOtherLeave", "Connexion établie avec l'enfant "+display.SimpleIdShowing(child_id)+" à l'adresse "+child_address.IP.String()+":"+strconv.Itoa(child_address.Port))
		}
	} else {

		display.Leave(G.Id, "handleOtherLeave", "Je ne suis pas le parent admis du site qui quitte, je dois me reconnecter à son parent, s'il n'est pas déjà dans mes voisins.")
		if _, exists := G.Neighbors[parent_to_connect_to_id]; exists {
			display.Leave(G.Id, "handleOtherLeave", "Le parent admis du site qui quitte est déjà dans mes voisins, je ne fais rien")
			return
		}

		private_port := protocol.Findval(content, "parent_admitted_private_port")
		ip := protocol.Findval(content, "parent_admitted_IP")
		new_address, err := net.ResolveUDPAddr("udp4", ip+":"+private_port)
		if err != nil {
			display.Error(G.Id, "handleOtherLeave", "failed to resolve private adress, refus du client")
			return
		}
		display.Leave(G.Id, "handleOtherLeave", "Le parent admis du site qui quitte n'est pas dans mes voisins, je dois me connecter à lui. Son id est : "+display.SimpleIdShowing(parent_to_connect_to_id)+" et son adresse est : "+new_address.IP.String()+":"+strconv.Itoa(new_address.Port))
		G.Neighbors[parent_to_connect_to_id] = new_address
	}
}
