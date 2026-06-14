/*
	HYPOTHESES

- Au debut un seul site admis, le reste sont admis apres demande et admission
*/

package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/google/uuid"
)

func makeUniqueId() string {
	return uuid.New().String()
}

/*
handleMessage

Gère un evenement message, vérifie qu'il faut bien le recevoir et le traite en conséquence
// TODO - Clean cette fonction au karcher
*/
func handleMessage(content string, adresse *net.UDPAddr, eventFile chan<- Event) {

	msgContent := protocol.Findval(content, "msg")

	elu := protocol.Findval(content, "elu")
	from := protocol.Findval(content, "from")
	to := protocol.Findval(content, "to")
	not := protocol.Findval(content, "not")

	display.Info(G.Id, "handleMessage", "reception de : "+msgContent)

	if from == G.Id { // on ne traite pas nos propres messages
		return
	}

	// demande -> non admis -> pas de id -> on ne peut pas tracke les numeros de série
	if msgContent != DEMANDE {
		num_msg, err := strconv.Atoi(protocol.Findval(content, "num_msg"))
		if err != nil {
			display.Warning(G.Id, "handleMessage", "Bad message : no num_msg could not be found or converted")
			return
		}

		if !UpdateInterval(G.IntervallesMsg, from, num_msg) {

			return
		}
	}

	if to == "all" {

		//forward(content, from) // TODO voir si c'est nescessaire

	} else if to != "" && to != G.Id {

		// je foward pour que la personne concernée le recoive
		//forward(content, from) // TODO voir si c'est nescessaire
		return
	}

	if not != "" && not == G.Id {
		display.Warning(G.Id, "handleMessage", "je dois pas le recevoir : "+msgContent)
		return
	}

	if !G.Admis && msgContent != ADMIS {
		display.Warning(G.Id, "handleMessage", "je suis pas admis : "+msgContent)
		return
	}

	display.Info(G.Id, "handleMessage", "prise en compte de : "+msgContent)

	switch msgContent {
	case ADMIS:
		{
			if !G.Admis {
				nbSites, err := strconv.Atoi(protocol.Findval(content, "nb_sites"))
				if err != nil {
					display.Error(G.Id, "handleMessage:ADMIS", "Erreur nb_sites recu "+err.Error())
					panic(err)
				}

				my_id := protocol.Findval(content, "your_id")

				// recallage du nb de sites

				G.Id = my_id

				G.NbSites = nbSites // le nb site envoyé est déjà incrémenté
				G.Admis = true
				G.Neighbors[from] = adresse // mise a jour des vosisin
				display.Info("", "", "Je suis admis ! Mon id est : "+G.Id)
				display.Info(G.Id, "DEBUG:handleNewMember", "envoie nb admis:if!admis : ")
				tellAppIAmAdmis()

			}
		}
	case BLEU:
		recptionMsgBleu(from, elu)
	case ROUGE:
		receptionMsgRouge(from, elu, eventFile)
	case NEW_MEMBER:
		handleNewMember(protocol.Findval(content, "his_id"))
	case LEAVE:
		handleOtherLeave(adresse, content, eventFile)
	case DEMANDE:
		handleDemande(adresse, eventFile)
	case CTL_MESSAGE:
		handleOtherCtlMessage(content)
	}
}

func tellAppIAmAdmis() {
	sendToCrlfromNET(protocol.Msg_format_Ctrl("type", ADMIS) + protocol.Msg_format_Ctrl("nb_sites", strconv.Itoa(G.NbSites)) + protocol.Msg_format_Ctrl("our_id", G.Id))
}

func handleOtherCtlMessage(msg string) {
	content := protocol.Findval(msg, "content")
	display.Recu(G.Id, "MAIN_NET", "DEBUG:contenu reçu par un autre controleur et que j'envoie au mien : "+msg)
	display.Recu(G.Id, "MAIN_NET", "DEBUG:contenu reçu par un autre controleur et que j'envoie au mien findval.content : "+content)
	sendToCTL(content)
}

func handleControlMessages(rcvmsg string) {
	if is_ctl_message(rcvmsg) {
		to := protocol.Findval(rcvmsg, "target") // TODO - BIEN NOTE CES CONVENTIONS
		if to == "my_net_ctl" {
			if protocol.Findval(rcvmsg, "msg") == "fromctl_wanna_leave" {
				display.Recu(G.Id, "MAIN_NET", "contenu reçu par mon controleur destiné à mon net (je veux partir) : "+rcvmsg)
				wannaLeave()
			} else if protocol.Findval(rcvmsg, "msg") == "fromctl_demande_admission" {
				display.Recu(G.Id, "MAIN_NET", "contenu reçu par mon controleur destiné à mon net (demande d'admission) : "+rcvmsg)
				if !G.Admis {
					display.Envoie(G.Id, "handleControlMessaged", "Broadcast demande")
					broadcast_demande()
				} else {
					display.Info(G.Id, "MAIN_NET", "Je demande à être admis mais je le suis déjà (ça ne devrait que lorsque le premier admis se connecte à la websocket sur l'interface).")
					tellAppIAmAdmis()
				}
			} else {
				display.Error(G.Id, "MAIN_NET", "contenu reçu par mon controleur destiné à mon net mais msg inconnu : "+rcvmsg)
			}
		} else {
			display.Recu(G.Id, "MAIN_NET", "contenu reçu par mon controleur et que je repartage : "+rcvmsg)

			if to == "all" || to == "" { // si c'est pour tous on broadcast
				broadcast(protocol.Msg_format("msg", CTL_MESSAGE) + protocol.Msg_format("content", rcvmsg))
			} else {
				send(protocol.Msg_format("msg", CTL_MESSAGE)+protocol.Msg_format("content", rcvmsg), to)
			}
		}
	}
}

/*
lirestdin
Lis les message provenant de STDIN et les ajoutent dans la file
*/
func lireStdin(eventFile chan<- Event) {
	var rcvmsg string
	for {

		//la on gère le message de la part de notre propre controlleur qu'on retransmet
		_, err := fmt.Scanln(&rcvmsg)
		display.Recu(G.Id, "lireStdin", "Recu de stdin :"+rcvmsg)
		if err != nil {
			display.Error("net", "erreur", "Lecture stdin terminée ou en erreur, arret lecture: "+err.Error())
			return
		}
		eventFile <- CtlMessageEvent{Content: rcvmsg}
	}

}

func main() {

	// arguments
	p_broadcast_port := flag.String("p", "8080", "le port de de broadcast de convention")
	p_admis := flag.Bool("a", false, "si je suis admis au départ. Par défaut non.")
	p_dev_mode := flag.Bool("dev", false, "faire tourner en local (authorise le partage de ports)")
	flag.Parse()

	// etat init
	G.Admis = *p_admis
	G.Dev_mode = *p_dev_mode
	if *p_dev_mode {
		display.Warning(G.Id, "main", "Partage des ports activé, peut causé des issues sur certains OS")
	}
	//p_firstApp = p_firstApp
	eventFile := make(chan Event, 100)

	if G.Admis { // si admis j'ai un id
		G.Id = makeUniqueId()
		tellAppIAmAdmis() // envoi de mon id
	}

	// init et écoute des sockets (broadcast et ciblée)
	init_sockets(*p_broadcast_port, eventFile)

	go lireStdin(eventFile)

	for event := range eventFile {
		switch ev := event.(type) {
		case MessageEvent:
			{
				handleMessage(ev.Content, ev.Adress, eventFile)
			}
		case CtlMessageEvent:
			{
				handleControlMessages(ev.Content)
			}
		}

	}
}
