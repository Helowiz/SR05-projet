package main

/* HYPOTHESES
- Au debut un seul site admis, le reste sont admis apres demande et admission
*/
import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"log"
	"maps"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var stderr = log.New(os.Stderr, "", 0)

var fieldsep = "/"
var keyvalsep = "="

// config serveur

var server_port string

/* Gestion messages */

var intervallesMsg = make(map[string][]protocol.Interval, 100)
var current_msg_num int = 0

/* VAGUE */
var parent = ""
var voisins = make(map[string]struct{}, 100)

var connectionMap = make(map[string]net.Conn, 10)
var NbVoisinsAttendus int

var nbSites int = 1

/* ELECTION */

var elu = ""

// gloval vars pour le site

var id string = ""
var admis = false

func makeUniqueId() string {
	return uuid.New().String()
}

var pid = os.Getpid()

func getVoisinsList() []string {
	voisinsList := slices.Collect(maps.Keys(voisins))
	for i, v := range voisinsList {
		voisinsList[i] = simpleIdShowing(v)
	}
	return voisinsList
}

func doPeriodique(eventFile chan<- Event) {
	alarm := time.NewTicker(1000 * time.Millisecond)
	for {
		if admis { // si je suis admis je stop l'alarme
			return
		}
		<-alarm.C
		eventFile <- AlarmEvent{}
	}

}

func debutVagueElection() {
	if parent == "" { // pas deja atteint => peut commencer election
		parent = id
		elu = id
		NbVoisinsAttendus = len(connectionMap)
		Vague(id, "debutDiffusionVague", "DEBUT DE LA VAGUE")
		msg := protocol.Msg_format_NET("msg", BLEU) + protocol.Msg_format_NET("elu", id) //envoie aux voisins
		send_to_neigh(msg, "")

	}
}

/* Retourne vrai si c'est un message d'un controlleur faux sinon (msg app) */
func is_ctl_message(msg string) bool {
	sHrcv := protocol.FindvalLight(msg, "hlg") // les message de control on une champ hlg
	if sHrcv != "" {
		return true
	}
	return false
}

func lirestdin(eventFile chan<- Event) {
	var rcvmsg string
	for {

		//la on gère le message de la part de notre propre controlleur qu'on retransmet
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error("net", "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			//return
		}

		if is_ctl_message(rcvmsg) {
			display.Recu("MAIN_NET", id, "contenu reçu par mon controleur et que je boradcast aux autres"+rcvmsg)
			broadcast(protocol.Msg_format_NET("msg", CTL_MESSAGE) + protocol.Msg_format_NET("content", rcvmsg))

		}
	}
}

func recptionMsgBleu(from string, elu_recu string) {
	Vague(id, "recptionMsgBleu", "from : "+simpleIdShowing(from)+" mon_parent : "+simpleIdShowing(parent))
	msg := ""
	if parent == "" || elu_recu < elu { // Première vague ou id de l’élu est plus petite que la précédente
		parent = from
		elu = elu_recu
		NbVoisinsAttendus = len(connectionMap) - 1
		if NbVoisinsAttendus > 0 {
			msg = protocol.Msg_format_NET("msg", BLEU) + protocol.Msg_format_NET("elu", elu_recu)
			Vague(id, "recsBleu", "NB VOISIN ATTENDU: "+strconv.Itoa(NbVoisinsAttendus)+" Mes voisins :"+strings.Join(getVoisinsList(), ", ")+" MESSAGE"+msg)

			send_to_neigh(msg, parent)
		} else {
			msg = protocol.Msg_format_NET("msg", ROUGE) + protocol.Msg_format_NET("elu", elu_recu)
			send(msg, parent)
			Vague(id, "recBleu", "NB VOISIN ATTENDU: "+strconv.Itoa(NbVoisinsAttendus)+" Mes voisins :"+strings.Join(getVoisinsList(), ", ")+" MESSAGE"+msg)
		}
	} else {

		if elu == elu_recu { // msg de la mm vague, => remontée vers parent
			msg = protocol.Msg_format_NET("msg", ROUGE) + protocol.Msg_format_NET("elu", elu)
			send(msg, parent)
			Vague(id, "recBleu", "NB VOISIN ATTENDU: "+strconv.Itoa(NbVoisinsAttendus)+" MESSAGE"+msg)
		}

	}

}

func receptionMsgRouge(from string, elu_recu string, eventFile chan<- Event) {
	Vague(id, "recRouge", "from : "+from+" mon elu : "+elu+" elurecu : "+elu_recu[:10])
	if elu_recu == elu { // on accept que la vague courante

		NbVoisinsAttendus--
		Vague(id, "recRouge", "C'est ma vague nbVoisins attendus : "+strconv.Itoa(NbVoisinsAttendus))
		if NbVoisinsAttendus == 0 {
			if elu == id { // on a gagne l'election
				Vague(id, "receptionMsgRouge", "FIN JE SUIS ELU !!! j'admet le site qui demande")
				parent = ""
				elu = ""
				NbVoisinsAttendus = len(voisins)
				handleAdmit(*pendingAjout, eventFile)

			} else {
				msg := protocol.Msg_format_NET("msg", ROUGE) + protocol.Msg_format_NET("elu", elu)
				Vague(id, "recRouge", "NB VOISIN ATTENDU: "+strconv.Itoa(NbVoisinsAttendus)+"MESSAGE"+msg)
				send(msg, parent)
			}
		}
	}
}

/*
handleNewMember

- Gere le signal qu'un nouveau member a ete ajouté
- Sert aussi a reset les variables sachant on a perdu l'election si on participait
*/
func handleNewMember(new_member_id string) {
	if pendingAjout != nil {
		(*pendingAjout).Close()
		pendingAjout = nil
	}

	if parent != "" || elu != "" { // on etait dans une election, on la perd
		parent = ""
		elu = ""
		nbSites += 1
		display.Info(id, "handleNewMember", "envoie nb admis:handleNewMember : "+new_member_id[:10])
		sendToCrlfromNET(protocol.Msg_format_Ctrl("nb_sites", strconv.Itoa(nbSites)))

	}
}

func handleMessage(content string, conn net.Conn, eventFile chan<- Event) {
	num_msg, err := strconv.Atoi(protocol.FindvalLight(content, "num_msg"))
	if err != nil {
		Warning(id, "handleMessage", "Bad message : no num_msg could not be found or converted")
		return
	}

	msgContent := protocol.FindvalLight(content, "msg")

	elu := protocol.FindvalLight(content, "elu")
	from := protocol.FindvalLight(content, "from")
	to := protocol.FindvalLight(content, "to")
	not := protocol.FindvalLight(content, "not")

	Info(id, "handleMessage", "reception de : "+msgContent)

	if from == id { // on ne traite pas nos propres messages
		return
	}

	if !protocol.UpdateIntervalString(intervallesMsg, from, num_msg) {
		return
	}

	if to == "all" {

		forward(content, from)

	} else if to != "" && to != id {

		// je foward pour que la personne concernée le recoive
		forward(content, from)
		return
	}

	if not != "" && not == id {
		Warning(id, "handleMessage", "je dois pas le recevoir : "+msgContent)
		return
	}

	if !admis && msgContent != ADMIS {
		Warning(id, "handleMessage", "je suis pas admis : "+msgContent)
		return
	}

	Info(id, "handleMessage", "prise en compte de : "+msgContent)

	switch msgContent {
	case ADMIS:
		{
			my_id := protocol.FindvalLight(content, "your_id")
			nbSites, err = strconv.Atoi(protocol.FindvalLight(content, "nb_sites"))
			if err != nil {
				display.Error(id, "handleMessage:ADMIS", "Erreur nb_sites recu "+err.Error())
				panic(err)
			}
			if !admis {
				admis = true
				id = my_id
				connectionMap[from] = conn // mise a jour du connection map
				// je commence mon serveur
				go server("localhost", server_port, connectionMap, eventFile)
				nbSites += 1
				display.Info(id, "DEBUG:handleNewMember", "envoie nb admis:if!admis : ")
				sendToCrlfromNET(protocol.Msg_format_Ctrl("nb_sites", strconv.Itoa(nbSites)))
				go lirestdin(eventFile)
				Info("", "", "Je suis admis ! Mon id est : "+id)
			}
		}
	case BLEU:
		recptionMsgBleu(from, elu)
	case ROUGE:
		receptionMsgRouge(from, elu, eventFile)
	case NEW_MEMBER:
		handleNewMember(protocol.FindvalLight(content, "his_id"))

	default:
		{
			display.Recu("MAIN_NET", id, "DEBUG:contenu reçu par un autre controleur et que j'envoie au mien all"+content)
			display.Recu("MAIN_NET", id, "DEBUG:contenu reçu par un autre controleur et que j'envoie au mien findval.content"+protocol.FindvalLight(content, "content"))

			sendToCrl(protocol.FindvalLight(content, "content"))
		}

	}
}

func handleAlarm(admitted_adress string, admitted_port string, eventFile chan<- Event) {
	if !admis {
		// demande l'admission, attend le resultat
		client(admitted_adress, admitted_port, connectionMap, eventFile, false)
	}
}

func sendToCrl(msg string) {

	fmt.Println(msg)
}
func sendToCrlfromNET(msg string) {
	display.Envoie("MAIN_NET:sendtocrtlNet", id, "contenu envoyé à mon controleur de la part de net"+msg)
	msg += protocol.Msg_format_Ctrl("msg", "net")
	fmt.Println(msg)
}
func main() {

	p_server_port := flag.String("p", "8081", "le port de mon contrôleur NET quand je serai admis")
	p_admitted_port := flag.String("sp", "8081", "le port du contrôleur NET a qui je demande l'admission, ou à qui je suis déjà connecté si je suis déjà admis")
	p_admitted_adress := flag.String("sa", "localhost", "l'adresse du contrôleur NET a qui je demande l'admission, ou à qui je suis déjà connecté si je suis déjà admis")
	p_admis := flag.Bool("a", false, "si je suis admis au départ. Par défaut non.")

	flag.Parse()
	admis = *p_admis
	server_port = *p_server_port

	//p_firstApp = p_firstApp
	eventFile := make(chan Event, 100)

	//go lire_msg(eventFile)

	if admis {
		id = makeUniqueId()

		go server("localhost", server_port, connectionMap, eventFile)
		if *p_admitted_port != server_port {
			go client(*p_admitted_adress, *p_admitted_port, connectionMap, eventFile, true)
			go lirestdin(eventFile)

		}
	} else {
		go doPeriodique(eventFile)
	}

	for event := range eventFile {
		switch ev := event.(type) {
		case MessageEvent:
			{
				handleMessage(ev.Content, ev.Conn, eventFile)
			}
		case NewConnEvent:
			{
				// TODO - mettre la vague ici ? (au lieu de dans handleClient)
				handleClient(ev.Conn, connectionMap, eventFile)
			}
		case AlarmEvent:
			{
				handleAlarm(*p_admitted_adress, *p_admitted_port, eventFile)
			}

		}

	}
}
