package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"SR05_projet/snapshot"
	"errors"
	"flag"
	"fmt"
	"strconv"
)

type EltMapFile struct {
	msg_type string
	val_h    int
}

type Estampille struct {
	id_site string
	val_h   int
}

// global vars
var h = 0               // horloge du site
var this_id string = "" // id du site (passe en param)
var proc_name string    // nom du site (passe en param)
var nbSites = 1         // mis à jour par le NET

var app_en_sc bool = false // indique si l'app est en section critique

var map_file = make(map[string]EltMapFile) // map pour la file d'attente
var horloge_vect = make(map[string]int)    // map pour l'horloge vectorielles

// Snapshot
const WHITE = "blanc"
const RED = "rouge"
const APP_SNAPSHOT = "snapshot"
const ADMIS = "admis"

var stopSnapshot = false
var sauvMsg []string

// Pour l'algo
var color = WHITE
var initiator = false
var total = 0
var localStat *snapshot.Snapshot
var globalState *snapshot.GlobalSnapshot

// seulement utile pour initiateur
var nbStateExpected int
var nbMsgExpected int

var idCurrentSnap = 0

/* Check si j'ai la plus petite estampille et que j'ai bien tout les sites dans la map*/
func smallest_estampille() bool {
	if len(map_file) != nbSites {
		display.Warning(this_id, "smallest_estampille", "Je n'ai pas tout les sites map file : "+fmt.Sprint(map_file)+"nb sites : "+strconv.Itoa(nbSites))
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
	sHrcv := protocol.Findval(msg, "hlg") // les message de control on une champ hlg
	if sHrcv != "" {
		return true
	}
	return false
}

func is_net_message(msg string) bool {
	val := protocol.Findval(msg, "msg")
	if val == "net" {
		display.Info(proc_name, "DEBUG:is_net_message", "test is_net_message retourne vrai ")
		return true
	}
	return false
}

/* Extrait l'estampille d'un message, retourne aussi une erreur en cas de probleme */
func estampille_from_msg(msg string) (Estampille, error) {
	rcv_h, err := strconv.Atoi(protocol.Findval(msg, "hlg"))
	rcv_id := protocol.Findval(msg, "id")
	if err != nil {
		display.Error("", "estampille_from_msg", "Un ou plusieurs champs manquants")
		return Estampille{"", 0}, errors.New("estampille malformée")
	}
	return Estampille{rcv_id, rcv_h}, nil

}

/*mets à jour l'horloge vectorielle à l'aide de celle reçue */
func vectorial_from_msg(msg string, vect map[string]int) error {
	rcv_h_vect_string := protocol.Findval(msg, "hlgvect")
	if rcv_h_vect_string == "" {
		return errors.New("champ hlgvect manquant")
	}

	rcv_h_vect := protocol.StringToVect(rcv_h_vect_string)
	vect = protocol.RecalerVectoriel(vect, rcv_h_vect)

	//TODO je suis pas sûre qu'on doive réincrémenter ici
	vect[this_id]++
	return nil
}

/* Traite un message recu d'une autre application de controle */
func parse_ctl_message(msg string) {

	est, err := estampille_from_msg(msg)
	display.Info(this_id, "parse_ctl_message", "parsing message : "+msg+" de : "+est.id_site[0:20])
	//creer la variable de l'horloge vectorielle quand besoin
	err_vect := vectorial_from_msg(msg, horloge_vect)

	msg_content := protocol.Findval(msg, "msg")
	receiveColor := protocol.Findval(msg, "color")
	receiveSnapId, _ := strconv.Atoi(protocol.Findval(msg, "snap_id"))

	// si le msg est bien un msg de l'app (et pas de la snapshot)
	isAppMsg := msg_content == "requete" || msg_content == "accuse" || msg_content == "liberation"

	if isAppMsg && receiveColor == WHITE && color == RED { // msg prepost
		if initiator {
			handlePrepostMsg(msg)
		} else {
			msgToSend := "prepost" + protocol.Msg_format_Ctrl("value", protocol.Findval(msg, "msg"))
			envoyer_tous(msgToSend)
		}
	}

	if err != nil {
		display.Error(proc_name, "parse_ctl_msg", "Estampille extraction failed")
		return
	}
	if err_vect != nil {
		display.Error(proc_name, "parse_ctl_msg", "Horloge vectorielle extraction failed")
		return
	}
	if msg_content == "" {
		display.Error(proc_name, "parse_message", "Message malforme : "+msg_content)
		return
	}

	switch msg_content {

	case "requete":
		total--
		rec_dem_sc(est)

	case "accuse":
		total--
		display.Info(this_id, "parse_ctl_message", "ACCUSE RECEIVED")

		rec_accuse_sc(est)

	case "liberation":
		// on update les donnes avec le champ data recu
		//(un message de liberation devrai toujours avoir un champ data)
		total--
		newData := protocol.Findval(msg, "data")
		rec_fin_sc(est)
		sendToApp("data", newData)

	case "state": //global_state=blabla bilan=0
		if initiator && receiveSnapId == idCurrentSnap {
			receiveGlobalState := protocol.Findval(msg, "global_state")
			display.Info("STATE", "état", "état reçu")
			receiveTotal, _ := strconv.Atoi(protocol.Findval(msg, "total"))
			receiveSnapshot, _ := snapshot.ToSnapshot(receiveGlobalState)

			globalState = snapshot.Merge(globalState, receiveSnapshot)
			nbStateExpected--
			nbMsgExpected = nbMsgExpected + receiveTotal
			display.Info("", "STATE", "Etat attendu : "+strconv.Itoa(nbStateExpected)+" message attentu : "+strconv.Itoa(nbMsgExpected))
			if nbStateExpected == 0 && nbMsgExpected == 0 {
				endSnapshot()
			}
		}
	case "prepost":
		if initiator && receiveSnapId == idCurrentSnap {
			handlePrepostMsg(msg)
		}
	case "reset_snapshot":
		if receiveSnapId == idCurrentSnap { // On ne reset que si c'est le bon snapshot
			resetSnapshot()
		}
	case "reload_snapshot":
		// demande la dernière global state
		// envoie la global state à l'APP
	default:
		return
	}
}

func endSnapshot() {
	hvmap := snapshot.GetHVMap(*globalState)
	if snapshot.CheckCoherenceSnap(hvmap) {
		snapshot.SaveSnapshot(globalState)
	} else {
		display.Error(proc_name, "endSnapshot", "Snapshot incohérente !")
	}
	envoyer_tous("reset_snapshot") // TODO
	resetSnapshot()
}

func resetSnapshot() {
	display.Info("RESET", "", "")
	initiator = false
	color = WHITE
	localStat = nil
	globalState = nil
	sauvMsg = nil
}

func handlePrepostMsg(msg string) {
	nbMsgExpected--
	globalState = snapshot.MergeMsg(globalState, msg)

	display.Info("", "PREPOST", "Etat attendu : "+strconv.Itoa(nbStateExpected)+" message attentu : "+strconv.Itoa(nbMsgExpected))
	if nbStateExpected == 0 && nbMsgExpected == 0 {
		endSnapshot()
	}
}

/* Traite un message recu de l'application de base */
func parse_app_msg(msg string) {
	type_msg := protocol.Findval(msg, "type") // s'il retourne vide on ignore le message de toute facon

	switch type_msg {
	case "fromapp_debut_sc":
		app_dem_sc()
	case "fromapp_fin_sc":
		newData := protocol.Findval(msg, "data")
		app_fin_sc(newData)
	case "fromapp_wanna_leave":
		wanna_leave()
	case "fromapp_demande_admission":
		demande_admission()
	case "snapshot_init":
		color = RED
		initiator = true
		localStat, _ = snapshot.ToSnapshot(protocol.Findval(msg, "snap"))

		localStat.HorlogeVect = make(map[string]int) // copie HV
		for k, v := range horloge_vect {
			localStat.HorlogeVect[k] = v
		}
		globalState = snapshot.Merge(nil, localStat)
		idCurrentSnap++
		nbStateExpected = nbSites - 1
		nbMsgExpected = total
	case "snapshot": // Snapshot reçu de l'APP
		receiveSnapshot := protocol.Findval(msg, "snap")
		localStat, _ = snapshot.ToSnapshot(receiveSnapshot)

		localStat.HorlogeVect = make(map[string]int) // TODO faire une fonction de ça
		for k, v := range horloge_vect {
			localStat.HorlogeVect[k] = v
		}
		snapshotStr, _ := snapshot.ToString(localStat)
		msgToSend := "state" + protocol.Msg_format_Ctrl("global_state", snapshotStr) + protocol.Msg_format_Ctrl("total", strconv.Itoa(total))
		envoyer_tous(msgToSend)

		for _, m := range sauvMsg {
			parse_ctl_message(m)
		}
		sauvMsg = nil
		stopSnapshot = false
	case "reload":
		// demande la dernière global state
		// envoie la global state à l'APP
		// faire passer le message
	default:
		return
	}
}

func envoyer(msg string, to string) {

	horloge_vect_str := protocol.VectToString(horloge_vect)
	sendToCtl(protocol.Msg_format_Ctrl("target", to) + protocol.Msg_format_Ctrl("id", this_id) + protocol.Msg_format_Ctrl("hlg", strconv.Itoa(h)) + protocol.Msg_format_Ctrl("hlgvect", horloge_vect_str) + protocol.Msg_format_Ctrl("msg", msg) + protocol.Msg_format_Ctrl("color", color))
}

func envoyer_tous(msg string) {
	horloge_vect_str := protocol.VectToString(horloge_vect)
	// id=id_site hlg=val_h msg=msg
	sendToCtl(protocol.Msg_format_Ctrl("id", this_id) + protocol.Msg_format_Ctrl("hlg", strconv.Itoa(h)) + protocol.Msg_format_Ctrl("hlgvect", horloge_vect_str) + protocol.Msg_format_Ctrl("msg", msg) + protocol.Msg_format_Ctrl("color", color))
}

/* Envoi aux autres le signal de liberation avec les donnes a jour*/
func envoyer_liberation(newData string) {
	horloge_vect_str := protocol.VectToString(horloge_vect)
	sendToCtl(protocol.Msg_format_Ctrl("id", this_id) + protocol.Msg_format_Ctrl("hlg", strconv.Itoa(h)) + protocol.Msg_format_Ctrl("hlgvect", horloge_vect_str) + protocol.Msg_format_Ctrl("msg", "liberation") + protocol.Msg_format_Ctrl("data", newData))

}

/* Envoi un message a l'app (just stdout car l'app y est connectee) */
func sendToApp(msg_type string, value string) {
	fmt.Println(protocol.Msg_format_Ctrl("type", msg_type) + protocol.Msg_format_Ctrl("value", value))
}

/* Previens l'application de base qu'on est en section critique*/
func debut_sc_app() {
	if !app_en_sc {
		sendToApp("section_critique", "true")
		app_en_sc = true
	}
}

/* Previens l'application de base qu'on est en fin de section critique*/
func fin_sc_app(newData string) {
	if app_en_sc {
		sendToApp("section_critique", "false") // envoi le message a l'app que c'est la fin de la sc
		if newData != "" {
			sendToApp("data", newData) // envoi les donnes modifies a l'app pour qu'elle confirme son changement
		}
		app_en_sc = false
	}
}

/* Préviens l'application de base qu'un autre site veut entrer en section critique*/
func other_in_sc_app() {
	if !app_en_sc {
		sendToApp("section_critique", "other")
	}
}

/* Traite une demande d'entree en section critique de l'application de base */
func app_dem_sc() {
	h++
	horloge_vect[this_id]++
	map_file[this_id] = EltMapFile{"requete", h}
	if nbSites == 1 { // si je suis seul je peux entrer direct en SC
		debut_sc_app()
	} else {
		total += (nbSites - 1) // on envoie des messages à tous le monde sauf soi
		envoyer_tous("requete")
	}
}

/* Traite une demande de fin de section critique de l'application de base */
func app_fin_sc(newData string) {
	h++
	horloge_vect[this_id]++
	map_file[this_id] = EltMapFile{"liberation", h}
	if nbSites != 1 { // si je suis seul je peux entrer direct en SC
		total += (nbSites - 1) // on envoie des messages à tous le monde sauf soi
		envoyer_liberation(newData)
	}
	fin_sc_app(newData)
}

/* Reception d'une requete de section critique d'un autre site */
func rec_dem_sc(est Estampille) {

	h = protocol.Recaler(h, est.val_h)

	map_file[est.id_site] = EltMapFile{"requete", est.val_h}
	total++
	envoyer("accuse", est.id_site)

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc_app()
	} else {
		other_in_sc_app()
	}

}

/* Reception d'une fin de section critique d'un autre site */
func rec_fin_sc(est Estampille) {
	h = protocol.Recaler(h, est.val_h)
	map_file[est.id_site] = EltMapFile{"liberation", est.val_h}

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc_app()
	}
}

/* Reception d'un accuse de reception d'un autre site */
func rec_accuse_sc(est Estampille) {
	h = protocol.Recaler(h, est.val_h)
	display.Info(this_id, "rec_accuse_sc", "Gestion accuse")
	if _, ok := map_file[est.id_site]; !ok { // si le site n'exist pas encore dans la map on l'ajoute
		map_file[est.id_site] = EltMapFile{"accuse", est.val_h}

	} else {
		// on n'ecrase pas une une ancienne requete avec un accuse
		if map_file[est.id_site].msg_type != "requete" {
			map_file[est.id_site] = EltMapFile{"accuse", est.val_h}
		}
	}

	// verifier si l'arrivee de ce message nous permet de passer en SC
	// j'ai envoye une requete et j'ai la plus petite estampille
	if (map_file[this_id].msg_type == "requete") && smallest_estampille() {
		debut_sc_app()
	}
}

func wanna_leave() {
	envoyer("fromctl_wanna_leave", "my_net_ctl")
}

func demande_admission() {
	envoyer("fromctl_demande_admission", "my_net_ctl")
}

func sendToCtl(msg string) {
	if color == RED {
		msg += protocol.Msg_format_Ctrl("snap_id", strconv.Itoa(idCurrentSnap))
	}
	display.Envoie(proc_name, "sendToCtl", "envoi : "+msg)
	fmt.Println(msg)
}

/*
handleAdmitted

Init du site complet avec son nouvel ID et les informations nescessaires
*/
func handleAdmitted(net_msg string) {
	//initialisation de l'horloge vectorielle
	var err error

	nbSites, err = strconv.Atoi(protocol.Findval(net_msg, "nb_sites"))
	if err != nil {
		display.Error(proc_name, "handleAdmitted", "Erreur nb_sites recu "+err.Error())
		panic(err)
	}
	our_id := protocol.Findval(net_msg, "our_id")
	if our_id == "" {
		display.Error(proc_name, "handleAdmitted", "id recu vide")
		return
	} else if our_id == this_id {
		display.Info(proc_name, "handleAdmitted", "Je suis déjà admis, je préviens l'appli !")
	} else {
		this_id = our_id
		horloge_vect[this_id] = 0
		map_file[this_id] = EltMapFile{"liberation", h}
	}

	admis_app() // on préviens l'app qu'on est admis
}

/*
admis_app

Previens l'application de base qu'on est admis dans le réseau
*/
func admis_app() {
	sendToApp(protocol.ADMIS, "true")
}

func handleNewSite(net_msg string) {
	var err error
	nbSites, err = strconv.Atoi(protocol.Findval(net_msg, "nb_sites"))
	if err != nil {
		display.Error(proc_name, "handleNewSite", "Erreur nb_sites recu "+err.Error())
		panic(err)
	}
}

func handleSelfLeave() {
	this_id = ""
	horloge_vect = make(map[string]int)
	map_file = make(map[string]EltMapFile)
	sendToApp(protocol.LEAVE, "true")
}

func handleOtherLeave(net_msg string) {
	var err error
	nbSites, err = strconv.Atoi(protocol.Findval(net_msg, "nb_sites"))
	if err != nil {
		display.Error(proc_name, "handleOtherLeave", "Erreur nb_sites recu "+err.Error())
		panic(err)
	}
	leaving_id := protocol.Findval(net_msg, "its_id")
	if leaving_id == this_id {
		display.Error(this_id, "handleOtherLeave", "J'ai reçu mon propre message de leave (ne devrait JAMAIS arriver)")
		return
	}
	if _, ok := map_file[leaving_id]; ok {
		delete(map_file, leaving_id)
	} else {
		display.Warning(this_id, "handleOtherLeave", "J'ai reçu un signal de leave d'un site que je ne connaissais pas : "+leaving_id)
	}
}

func parse_net_message(net_msg string) {
	type_msg := protocol.Findval(net_msg, "type")
	switch type_msg {
	case protocol.ADMIS:
		handleAdmitted(net_msg)
	case protocol.NEW_MEMBER:
		handleNewSite(net_msg)
	case protocol.LEAVE:
		handleSelfLeave()
	case "other_leave":
		handleOtherLeave(net_msg)
	}

}

func main() {

	// arguments en entree
	p_nom := flag.String("n", "controler", "nom")
	flag.Parse()

	// init de l'identite du site
	proc_name = *p_nom

	display.Info(proc_name, "main", "Démarrage du contrôleur...")

	var rcvmsg string // message recu

	for {
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error(*p_nom, "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			//return
		}
		// display.Info(*p_nom, "main", "recu : "+rcvmsg)
		if is_ctl_message(rcvmsg) { // reception d'un autre site

			if this_id == "" { // si je suis pas admis je fais rien
				continue
			}

			receiveColor := protocol.Findval(rcvmsg, "color")
			receiveIdSnap, _ := strconv.Atoi(protocol.Findval(rcvmsg, "snap_id"))

			if receiveIdSnap > idCurrentSnap && receiveColor == RED && color == WHITE { // signal pour prendre une snapshot
				color = RED
				stopSnapshot = true
				idCurrentSnap = receiveIdSnap // pour le reset et pas recevoir des messages rouge alors qu'on est redevenu blanc et réenclencher une snapshot
				display.Envoie(proc_name, "MAIN", proc_name+"devient ROUGE TOTAL "+strconv.Itoa(total))
				sendToApp("snapshot_app", protocol.VectToString(horloge_vect))
			}

			if stopSnapshot { // pour bloquer le controleur pendant la prise de snap depuis l'app
				sauvMsg = append(sauvMsg, rcvmsg)
			} else {
				parse_ctl_message(rcvmsg)
			}

		} else if is_net_message(rcvmsg) { // reception d'un message du NET
			display.Recu("MAIN_APP", proc_name, "contenu reçu de la part de mon NET : "+rcvmsg)
			parse_net_message(rcvmsg)
		} else { // reception de l'application de base
			h++
			horloge_vect[this_id]++
			parse_app_msg(rcvmsg)
		}

	}
}
