package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
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
var h = 0
var this_id int
var n_sites int
var messages_recus = make(map[Estampille]struct{}) // on veut just verifier l'existence

var map_file = make(map[int]EltMapFile)

/* Check si j'ai la plus petite estampille */
func smallest_estampille() bool {
	if len(map_file) != n_sites {
		display.Info("Controleur : "+strconv.Itoa(this_id), "smallest_estampille", "Pas encore decouvert tout le réseau")
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
	return true
}

func estampille_from_msg(msg string) (Estampille, error) {
	rcv_h, err := strconv.Atoi(protocol.Findval(msg, "hlg", ""))
	rcv_id, err := strconv.Atoi(protocol.Findval(msg, "id", ""))
	if err != nil {
		display.Error("", "estampille_from_msg", "Un ou plusieurs champs manquants")
		return Estampille{0, 0}, errors.New("estampille malformée")
	}
	return Estampille{rcv_id, rcv_h}, nil

}

/* En fonction du type de message recu, execute le bon traitement */
func parse_ctl_message(msg string) {
	display.Info("Controleur : "+strconv.Itoa(this_id), "parse_ctl_msg", "Parsing : "+msg)
	est, err := estampille_from_msg(msg)
	msg_content := protocol.Findval(msg, "msg", "")

	if err != nil {
		return
	}
	if msg_content == "" {
		display.Error("Message incomplet", "parse_message", "Message malforme : "+msg_content)
		return
	}

	switch msg_content {

	case "requete":
		rec_dem_sc(est)
	case "accuse":

		rec_accuse_sc(est)

	case "liberation":
		rec_fin_sc(est)

	default:
		display.Info("", "parse_message", "Message ignore"+msg_content)

		return

	}

}

func parse_app_msg(msg string) {
	display.Info("", "parse_app_msg", "Parsing : "+msg)
	switch msg {
	case "fromapp_debut_sc":
		app_dem_sc()
	case "fromapp_fin_sc":
		app_fin_sc()

	default:
		display.Info("", "parse_message", "Message ignore : "+msg)
		return
	}
}

// /* Envoi aux autres sites le message d'initialisation */
// func envoi_init() {
// 	envoyer_tous("init")
// }

// /* Traite un message d'initialisation d'un autre site */
// func rec_init(est Estampille) {
// 	map_file[est.id_site] = EltMapFile{"liberation", est.val_h}
// }

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

/* Route le message sans modifs aux successeurs*/
func forward(msg string) {
	fmt.Println(msg)
}

/* Previens l'application de base qu'on est en section critique*/
func debut_sc() {
	display.Info("Entrée en section critique", "debut_sc", "Entree SC")
	fmt.Println("toapp_debut_sc")
}

/* Previens l'application de base qu'on est en fin de section critique*/
func fin_sc() {
	display.Info("Fin section critique", "fin_sc", "Fin SC")
	fmt.Println("toapp_fin_sc")

}

/* Traite une demande d'entree en section critique de l'application de base */
func app_dem_sc() {
	h++
	map_file[this_id] = EltMapFile{"requete", h}
	envoyer_tous("requete")
}

/* Traite une demande de fin de section critique de l'application de base */
func app_fin_sc() {
	h++
	map_file[this_id] = EltMapFile{"liberation", h}
	envoyer_tous("liberation")
	fin_sc()
}

/* Reception d'une requete de section critique d'un autre site */
func rec_dem_sc(est Estampille) {
	protocol.Recaler(h, est.val_h)

	// verifier l'existence du site dans le map
	// _, ok := map_file[est.id_site]
	// if !ok {
	// 	display.Error("out of range", "rec_dem_sc", "id site not in map")
	// 	return
	// }

	map_file[est.id_site] = EltMapFile{"requete", est.id_site}

	envoyer("accuse", est.id_site)
	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc()
	}

}

/* Reception d'une fin de section critique d'un autre site */
func rec_fin_sc(est Estampille) {
	protocol.Recaler(h, est.val_h)
	map_file[est.id_site] = EltMapFile{"liberation", h}

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc()
	}
}

/* Reception d'un accuse de reception d'un autre site */
func rec_accuse_sc(est Estampille) {
	display.Info("", "rec_accuse_sc", "Reception accuse")
	protocol.Recaler(h, est.val_h)
	// si le site n'exist pas encore dans la map on l'ajoute
	if _, ok := map_file[est.id_site]; !ok {
		map_file[est.id_site] = EltMapFile{"accuse", h}

	} else {
		// on n'ecrase pas une une ancienne requete avec un accuse
		if map_file[est.id_site].msg_type != "requete" {
			map_file[est.id_site] = EltMapFile{"accuse", h}
		}
	}
	display.Info("", "rec_accuse_sc", strconv.FormatBool(smallest_estampille())+map_file[this_id].msg_type)
	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {

		debut_sc()
	}

}

func main() {

	p_nom := flag.String("n", "controler", "nom")
	p_nbsites := flag.String("nbsites", "controler", "nombre de sites")
	flag.Parse()

	n_sites, _ = strconv.Atoi(*p_nbsites)

	display.Info(*p_nom, "main", "Démarrage du contrôleur...")

	this_id = os.Getpid() // assigner notre pid a la variable global

	var rcvmsg string
	var hrcv int
	//var sndmsg string

	//envoi_init()

	for {
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error(*p_nom, "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			//return
		}
		//display.Info(*p_nom, "reception", "Reçu brut : "+rcvmsg)

		sHrcv := protocol.Findval(rcvmsg, "hlg", *p_nom)

		// reception d'un autre site
		if sHrcv != "" {
			hrcv, _ = strconv.Atoi(sHrcv)
			h = protocol.Recaler(h, hrcv)
			targetId := protocol.Findval(rcvmsg, "target", *p_nom)

			// extraire l'estampille
			est, err := estampille_from_msg(rcvmsg)
			if err != nil {
				display.Error(*p_nom, "erreur", "Erreur estampille: "+err.Error())
			}

			if _, ok := messages_recus[est]; ok { // si le message a deje ete traite on fait rien
				continue
			}

			// ajouter aux messages recus pour garder une trace
			messages_recus[est] = struct{}{}

			// si le message est pour ou pour tous nous on le traite
			if targetId == "" || targetId == strconv.Itoa(this_id) {
				// parse le message de controle
				parse_ctl_message(rcvmsg)
				// si le message est pour tous on le fait passer
				forward(rcvmsg)
			} else { // si le message n'est pas pour nous on le renvoi a nos successeurs
				forward(rcvmsg)
			}

		} else { // reception de l'application de base
			h = h + 1
			parse_app_msg(rcvmsg)
			//display.Info(*p_nom, "horloge", fmt.Sprintf("Pas de 'hlg', incrémentation locale -> H=%d", h))
		}

		/*
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
		*/
	}
}
