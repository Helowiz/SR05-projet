package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
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
			if elt.val_h < h || (elt.val_h == h && site_id < this_id) {
				return false
			}
		}
	}
	return true
}

// /* Ajout un element au map s'il n'y est pas deja */

// func add_if_nexists(m map[string]EltMapFile, key, liberation string) {
// 	if _, ok := m[key]; !ok {
// 		m[key] = EltMapFile{liberation, 0}
// 	}
// }

/* En fonction du type de message recu, execute le bon traitement */
func parse_ctl_message(msg string) {
	display.Info("Controleur : "+strconv.Itoa(this_id), "parse_ctl_msg", "Parsing : "+msg)

	msg_content := protocol.Findval(msg, "msg", "")
	rcv_h, err := strconv.Atoi(protocol.Findval(msg, "hlg", ""))
	rcv_id, err := strconv.Atoi(protocol.Findval(msg, "id", ""))
	if msg_content == "" || err != nil {
		display.Error("Message incomplet", "parse_message", "Un ou plusieurs champs manquants")
	}

	est := Estampille{rcv_id, rcv_h}
	switch msg_content {

	case "requete":
		rec_dem_sc(est)

	case "accuse":
		rec_accuse_sc(est)

	case "liberation":
		rec_fin_sc(est)

	}

}

func parse_app_msg(msg string) {
	display.Info("PARSING APP MESSAGE", "parse_app_msg", "Parsing : "+msg)
	switch msg {
	case "debut_sc":
		app_dem_sc()
	case "fin_sc":
		app_fin_sc()
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
	fmt.Println(protocol.Msg_format("target", strconv.Itoa(id)) + protocol.Msg_format("id", strconv.Itoa(this_id)) + protocol.Msg_format("hlg", strconv.Itoa(h)) + protocol.Msg_format("msg", msg))
}
func envoyer_tous(msg string) {
	fmt.Println(protocol.Msg_format("id", strconv.Itoa(this_id)) + protocol.Msg_format("hlg", strconv.Itoa(h)) + protocol.Msg_format("msg", msg))
}

/* Previens l'application de base qu'on est en section critique*/
func debut_sc() {
	display.Info("Entrée en section critique", "debut_sc", "Entree SC")
}

/* Previens l'application de base qu'on est en fin de section critique*/
func fin_sc() {
	display.Info("Fin section critique", "fin_sc", "Fin SC")
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
	protocol.Recaler(h, est.val_h)

	// On n’ecrase pas la date d’une requête par celle d’un accuse
	if map_file[est.id_site].msg_type != "requete" {
		map_file[est.id_site] = EltMapFile{"accuse", h}
	}

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
		display.Info(*p_nom, "reception", "Reçu brut : "+rcvmsg)

		sHrcv := protocol.Findval(rcvmsg, "hlg", *p_nom)

		// reception d'un autre site
		if sHrcv != "" {
			hrcv, _ = strconv.Atoi(sHrcv)
			h = protocol.Recaler(h, hrcv)
			targetId := protocol.Findval(rcvmsg, "target", *p_nom)

			// on ne traite que le message d'un autre controleur s'il est adresse a nous
			if targetId == "" || targetId == strconv.Itoa(this_id) {

				// parse le message de controle
				parse_ctl_message(rcvmsg)
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
