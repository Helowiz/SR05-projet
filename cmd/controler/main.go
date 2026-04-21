package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type EltTabFile struct {
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
var map_file map[int]EltTabFile

/* Check si j'ai la plus petite estampille */
func smallest_estampille() bool {
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

// au debut il faudrait que tous les sites s'envoi un message pour se connaitre
// comme ça on peut initialiser le map_file
func init_map_file() {

}

func envoyer(msg string, id int) {
	fmt.Println(protocol.Msg_format("target", strconv.Itoa(id))+protocol.Msg_format("id", strconv.Itoa(this_id))+protocol.Msg_format("hlg", strconv.Itoa(h)), protocol.Msg_format("msg", msg))
}
func envoyer_tous(msg string) {
	fmt.Println(protocol.Msg_format("id", strconv.Itoa(this_id))+protocol.Msg_format("hlg", strconv.Itoa(h)), protocol.Msg_format("msg", msg))
}

/* Previens l'application de base qu'on est en section critique*/
func debut_sc() {}

/* Previens l'application de base qu'on est en fin de section critique*/
func fin_sc() {}

/* Traite une demande d'entree en section critique de l'application de base */
func app_dem_sc() {
	h++
	map_file[this_id] = EltTabFile{"requete", h}
	envoyer_tous("requete")
}

/* Traite une demande de fin de section critique de l'application de base */
func app_fin_sc() {
	h++
	map_file[this_id] = EltTabFile{"liberation", h}
	envoyer_tous("liberation")
}

/* Reception d'une requete de section critique d'un autre site */
func rec_dem_sc(est Estampille) {
	protocol.Recaler(h, est.val_h)

	// verifier l'existence du site dans le map
	_, ok := map_file[est.id_site]
	if !ok {
		display.Error("out of range", "rec_dem_sc", "id site not in map")
		return
	}

	map_file[est.id_site] = EltTabFile{"requete", est.id_site}

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
	map_file[est.id_site] = EltTabFile{"liberation", h}

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
		map_file[est.id_site] = EltTabFile{"accuse", h}
	}

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc()
	}

}

func main() {

	p_nom := flag.String("n", "controler", "nom")
	flag.Parse()

	display.Info(*p_nom, "main", "Démarrage du contrôleur...")

	this_id = os.Getpid() // assigner notre pid a la variable global

	var rcvmsg string
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

		// reception d'un autre site
		if sHrcv != "" {
			oldH := h
			hrcv, _ = strconv.Atoi(sHrcv)
			h = protocol.Recaler(h, hrcv)
			display.Info(*p_nom, "recalage", fmt.Sprintf("H_locale=%d, H_recue=%d -> Nouvelle H=%d", oldH, hrcv, h))

		} else { // reception de l'application de base
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
