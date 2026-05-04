package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"SR05_projet/snapshot"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type EltMapFile struct {
	msg_type string
	val_h    int
}

type Estampille struct {
	id_site int
	val_h   int
}

// global vars
var h = 0            // horloge du site
var this_id int      // id du site (passe en param)
var proc_name string // nom du site (passe en param)
var n_sites int      // nombre de sites (passe en param)

var messages_recus = make(map[Estampille]struct{}) // map juste pour verifier l'existence des messages
var app_en_sc bool = false                         // indique si l'app est en section critique

var map_file = make(map[int]EltMapFile) // map pour la file d'attente

// Snapshot
const WHITE string = "blanc"
const RED string = "rouge"

var color = WHITE
var initiator = false
var total = 0
var globalState []snapshot.Snapshot

var nbStateExpected int
var nbMsgExpected int

var initiatorID int

/* Fonction utilitaire juste pour print la map file*/
func map_file_to_string() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-10s %-15s %-10s\n", "ID", "MSG_TYPE", "VAL_H"))
	sb.WriteString(strings.Repeat("-", 35) + "\n")
	for k, v := range map_file {
		sb.WriteString(fmt.Sprintf("%-10d %-15s %-10d\n", k, v.msg_type, v.val_h))
	}
	return sb.String()
}

/* Check si j'ai la plus petite estampille et que j'ai bien tout les sites dans la map*/
func smallest_estampille() bool {
	//display.Info(proc_name, "smallest_estampille", "Test estampille : \n"+map_file_to_string())

	if len(map_file) != n_sites {
		//display.Info(proc_name, "smallest_estampille", "Pas encore decouvert tout le réseau")
		return false
	}

	// pour chaque autre site dans mon map
	for site_id, elt := range map_file {
		if site_id != this_id { // si c'est pas moi même
			// comparaison avec les estampilles (on check le numero de site en cas d'egalite des horloges)
			if elt.val_h < map_file[this_id].val_h || (elt.val_h == map_file[this_id].val_h && site_id < this_id) {
				return false
			}
		}
	}
	//display.Info(proc_name, "smallest_estampille", "TEST REUSSI")

	return true
}

/* Retourne vrai si c'est un message d'un controlleur faux sinon (msg app) */
func is_ctl_message(msg string) bool {
	sHrcv := protocol.Findval(msg, "hlg", proc_name) // les message de control on une champ hlg
	if sHrcv != "" {
		return true
	}
	return false
}

/* Extrait l'estampille d'un message, retourne aussi une erreur en cas de probleme */
func estampille_from_msg(msg string) (Estampille, error) {
	rcv_h, err := strconv.Atoi(protocol.Findval(msg, "hlg", proc_name))
	rcv_id, err := strconv.Atoi(protocol.Findval(msg, "id", proc_name))
	if err != nil {
		display.Error("", "estampille_from_msg", "Un ou plusieurs champs manquants")
		return Estampille{0, 0}, errors.New("estampille malformée")
	}
	return Estampille{rcv_id, rcv_h}, nil

}

/* Traite un message recu d'une autre application de controle */
func parse_ctl_message(msg string) {
	// display.Info(proc_name, "parse_ctl_msg", "Parsing : "+msg)
	est, err := estampille_from_msg(msg)
	msg_content := protocol.Findval(msg, "msg", "")

	if err != nil {
		display.Error(proc_name, "parse_ctl_msg", "Estampille extraction failed")
		return
	}
	if msg_content == "" {
		display.Error(proc_name, "parse_message", "Message malforme : "+msg_content)
		return
	}

	switch msg_content {

	case "requete":
		rec_dem_sc(est)

	case "accuse":
		rec_accuse_sc(est)

	case "liberation":
		// on update les donnes avec le champ data recu
		//(un message de liberation devrai toujours avoir un champ data)
		newData := protocol.Findval(msg, "data", proc_name)
		rec_fin_sc(est)
		send_to_app("data", newData)

	case "state":
		h = protocol.Recaler(h, est.val_h) // on se synchronise
		if isInit {                        // c'est l'initateur qui rassemble toutes les snapshot
			receiveSnapshot, _ := snapshot.StringToSnapshot(protocol.Findval(msg, "snap", proc_name)) // Snapshot prise
			globalSnapshot = append(globalSnapshot, *receiveSnapshot)
			nbEtatAttendus = nbEtatAttendus - 1
			if nbEtatAttendus == 0 {
				snapshot.SaveGlobalSnapshot(globalSnapshot)
				isInit = false
				isSnapshot = false
				globalSnapshot = nil
				initiateID = -1
				h++
				envoyer_tous("snapshot_reset")
			}
		}
	case "snapshot_reset":
		h = protocol.Recaler(h, est.val_h)
		isSnapshot = false
		initiateID = -1
	default:
		//display.Info(proc_name, "parse_message", "Message ignore"+msg_content)
		return

	}

}

/* Traite un message recu de l'application de base */
func parse_app_msg(msg string) {
	//display.Info(proc_name, "parse_app_msg", "Parsing : "+msg)
	type_msg := protocol.Findval(msg, "type", proc_name) // s'il retourne vide on ignore le message de toute facon
	appColor := protocol.Findval(msg, "color", color)

	if appColor == RED && color == WHITE { // signal pour prendre une snapshot
		color = RED
		localSnapshot, _ := snapshot.StringToSnapshot(protocol.Findval(msg, "snap", proc_name)) // Snapshot prise
		globalState = append(globalState, *localSnapshot)
		msgSend := protocol.Msg_format("type", "state") + protocol.Msg_format("global_state", globalState) + protocol.Msg_format("total", strconv.Itoa(total))
		envoyer(msgSend, initiatorID)
	}

	if appColor == WHITE && color == RED { // msg prepost
		msgSend := protocol.Msg_format("type", "prepost") + protocol.Msg_format("value", msg)
		envoyer_tous(msgSend)
	}

	switch type_msg {
	case "fromapp_debut_sc":
		app_dem_sc()
	case "fromapp_fin_sc":
		newData := protocol.Findval(msg, "data", proc_name)
		if newData == "" {
			display.Error(proc_name, "parse_app_message", "Nouvelles donnees non trouvees")
			return
		}
		app_fin_sc(newData)
	case "snapshot_init":
		initiator = true
		localSnapshot, _ := snapshot.StringToSnapshot(protocol.Findval(msg, "snap", proc_name)) // Snapshot prise
		globalState = append(globalState, *localSnapshot)
		nbStateExpected = n_sites - 1
		nbMsgExpected = total
		initiatorID = this_id
		envoyer_tous("snapshot" + protocol.Msg_format("initiator", strconv.Itoa(this_id))) // on dit à tout le monde de faire une snapshot
	default:
		//display.Info(proc_name, "parse_app_message", "Message ignore : "+msg)
		return
	}
}

func envoyer(msg string, id int) {
	est := Estampille{this_id, h}
	fmt.Println(protocol.Msg_format("target", strconv.Itoa(id)) + protocol.Msg_format("id", strconv.Itoa(est.id_site)) + protocol.Msg_format("hlg", strconv.Itoa(est.val_h)) + protocol.Msg_format("msg", msg))

	// on ne veux pas traiter nos propres messages donc c'est comme si on l'avait deja recu
	messages_recus[est] = struct{}{}
}
func envoyer_tous(msg string) {
	est := Estampille{this_id, h}

	fmt.Println(protocol.Msg_format("id", strconv.Itoa(est.id_site)) + protocol.Msg_format("hlg", strconv.Itoa(est.val_h)) + protocol.Msg_format("msg", msg))

	// on ne veux pas traiter nos propres messages donc c'est comme si on l'avait deja recu
	messages_recus[est] = struct{}{}
}

/* Envoi aux autres le signal de liberation avec les donnes a jour*/
func envoyer_liberation(newData string) {
	est := Estampille{this_id, h}
	fmt.Println(protocol.Msg_format("id", strconv.Itoa(est.id_site)) + protocol.Msg_format("hlg", strconv.Itoa(est.val_h)) + protocol.Msg_format("msg", "liberation") + protocol.Msg_format("data", newData))
	// on ne veux pas traiter nos propres messages donc c'est comme si on l'avait deja recu
	messages_recus[est] = struct{}{}

}

/* Envoi un message a l'app (just stdout car l'app y est connectee) */
func send_to_app(msg_type string, value string) {
	fmt.Println(protocol.Msg_format("type", msg_type) + protocol.Msg_format("value", value) + protocol.Msg_format("color", color)) //ajout couleur pour la snapshot
	total++                                                                                                                        // pour la snapshot
}

/* Route le message sans modifs aux successeurs*/
func forward(msg string) {
	fmt.Println(msg)
}

/* Previens l'application de base qu'on est en section critique*/
func debut_sc_app() {
	if !app_en_sc {
		//display.Info(proc_name, "debut_sc", "Entree SC")
		send_to_app("section_critique", "true")
		app_en_sc = true
	}
}

/* Previens l'application de base qu'on est en fin de section critique*/
func fin_sc_app(newData string) {
	if app_en_sc {
		//display.Info(proc_name, "fin_sc", "Fin SC")

		// envoi le message a l'app que c'est la fin de la sc
		send_to_app("section_critique", "false")

		// envoi les donnes modifies a l'app pour qu'elle confirme son changement
		send_to_app("data", newData)

		app_en_sc = false
	}
}

/* Traite une demande d'entree en section critique de l'application de base */
func app_dem_sc() {
	h++
	map_file[this_id] = EltMapFile{"requete", h}
	envoyer_tous("requete")
}

/* Traite une demande de fin de section critique de l'application de base */
func app_fin_sc(newData string) {
	h++
	map_file[this_id] = EltMapFile{"liberation", h}
	envoyer_liberation(newData)
	fin_sc_app(newData)
}

/* Reception d'une requete de section critique d'un autre site */
func rec_dem_sc(est Estampille) {

	h = protocol.Recaler(h, est.val_h)

	map_file[est.id_site] = EltMapFile{"requete", est.val_h}

	envoyer("accuse", est.id_site)

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc_app()
	}

}

/* Reception d'une fin de section critique d'un autre site */
func rec_fin_sc(est Estampille) {
	//display.Info(proc_name, "rec_fin_sc", "")

	h = protocol.Recaler(h, est.val_h)
	map_file[est.id_site] = EltMapFile{"liberation", est.val_h}

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	//display.Info(proc_name, "rec_fin_sc", strconv.FormatBool(smallest_estampille())+map_file[this_id].msg_type)

	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc_app()
	}
}

/* Reception d'un accuse de reception d'un autre site */
func rec_accuse_sc(est Estampille) {
	//display.Info(proc_name, "rec_accuse_sc", "Reception accuse")
	h = protocol.Recaler(h, est.val_h)
	// si le site n'exist pas encore dans la map on l'ajoute
	if _, ok := map_file[est.id_site]; !ok {
		map_file[est.id_site] = EltMapFile{"accuse", est.val_h}

	} else {
		// on n'ecrase pas une une ancienne requete avec un accuse
		if map_file[est.id_site].msg_type != "requete" {
			map_file[est.id_site] = EltMapFile{"accuse", est.val_h}
		}
	}
	//display.Info(proc_name, "rec_accuse_sc", strconv.FormatBool(smallest_estampille())+map_file[this_id].msg_type)

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {

		debut_sc_app()
	}
}

func main() {

	// arguments en entree
	p_nom := flag.String("n", "controler", "nom")
	p_nbsites := flag.Int("nbsites", 1, "nombre de sites")
	flag.Parse()

	// init de l'identite du site
	proc_name = *p_nom
	this_id = os.Getpid() // assigner notre pid a la variable global

	// on note le nombre de sites
	n_sites = *p_nbsites

	display.Info(proc_name, "", "Démarrage du contrôleur...")

	var rcvmsg string // message recu

	// boucle infini qui scan stdin pour des messages
	for {
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error(*p_nom, "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			//return
		}

		display.Info(*p_nom, "reception", "Reçu brut : "+rcvmsg)

		// reception d'un autre site
		if is_ctl_message(rcvmsg) {

			targetId := protocol.Findval(rcvmsg, "target", *p_nom)

			// extraire l'estampille
			est, err := estampille_from_msg(rcvmsg)
			if err != nil {
				display.Error(*p_nom, "erreur", "Erreur estampille: "+err.Error())
				return
			}

			// si le message a deje ete traite on fait rien
			if _, ok := messages_recus[est]; ok {
				continue
			}

			// ajouter aux messages recus pour garder une trace
			messages_recus[est] = struct{}{}

			// si le message est pour nous ou pour tous nous on le traite
			if targetId == "" || targetId == strconv.Itoa(this_id) {
				// parse le message de controle
				parse_ctl_message(rcvmsg)
				// si le message est pour tous on le fait passer
				if targetId == "" {
					forward(rcvmsg)
				}
			} else { // si le message n'est pas pour nous on le renvoi a nos successeurs
				forward(rcvmsg)
			}

		} else { // reception de l'application de base
			h++
			total-- // pour la snapshot
			parse_app_msg(rcvmsg)
		}
	}
}
