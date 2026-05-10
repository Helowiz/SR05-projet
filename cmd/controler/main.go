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
var n_msg = 0        // num serie messages

var intervalles_recus = make(map[int][]protocol.Interval) // check si on a deja recu

var app_en_sc bool = false // indique si l'app est en section critique

var map_file = make(map[int]EltMapFile) // map pour la file d'attente
var horloge_vect = make(map[int]int)    // map pour l'horloge vectorielles

// Snapshot
const WHITE string = "blanc"
const RED string = "rouge"

var stopSnapshot = false
var sauvMsg []string

var color = WHITE
var initiator = false
var total = 0
var localStat *snapshot.Snapshot
var globalState *snapshot.GlobalSnapshot

// seulement utile pour initiateur
var nbStateExpected int
var nbMsgExpected int

var totalEnvoyer = 0
var totalRecu = 0

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
	if len(map_file) != n_sites {
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

/*mets à jour l'horloge vectorielle à l'aide de celle reçue */
func vectorial_from_msg(msg string, vect map[int]int) error {
	rcv_h_vect_string := protocol.Findval(msg, "hlgvect", proc_name)
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

	//creer la variable de l'horloge vectorielle quand besoin
	err_vect := vectorial_from_msg(msg, horloge_vect)

	msg_content := protocol.Findval(msg, "msg", "")
	receiveColor := protocol.Findval(msg, "color", color)

	// si le msg est bien un msg de l'app (et pas de la snapshot)
	isAppMsg := msg_content == "requete" || msg_content == "accuse" || msg_content == "liberation"

	if receiveColor == RED && color == WHITE { // signal pour prendre une snapshot
		color = RED
		stopSnapshot = true
		sendToApp("snapshot_app", protocol.VectToString(horloge_vect))
	}

	if isAppMsg && receiveColor == WHITE && color == RED { // msg prepost
		msgToSend := "prepost" + protocol.Msg_format("value", protocol.Findval(msg, "msg", proc_name))
		envoyer_tous(msgToSend)
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
		rec_accuse_sc(est)

	case "liberation":
		// on update les donnes avec le champ data recu
		//(un message de liberation devrai toujours avoir un champ data)
		total--
		newData := protocol.Findval(msg, "data", proc_name)
		rec_fin_sc(est)
		sendToApp("data", newData)

	case "state": //global_state=blabla bilan=0
		if initiator {
			receiveGlobalState := protocol.Findval(msg, "global_state", proc_name)
			display.Info("STATE", "état", "état reçu")
			receiveTotal, _ := strconv.Atoi(protocol.Findval(msg, "total", proc_name))
			receiveSnapshot, _ := snapshot.ToSnapshot(receiveGlobalState)

			globalState = snapshot.Merge(globalState, receiveSnapshot)
			nbStateExpected--
			nbMsgExpected = nbMsgExpected + receiveTotal
			if nbStateExpected == 0 && nbMsgExpected == 0 {
				endSnapshot()
			}
		}
	case "prepost":
		if initiator {
			nbMsgExpected--
			globalState = snapshot.MergeMsg(globalState, msg)
			if nbStateExpected == 0 && nbMsgExpected == 0 {
				endSnapshot()
			}
		}
	case "reset_snapshot":
		resetSnapshot()
	default:
		return
	}
}

func endSnapshot() {
	snapshot.SaveSnapshot(globalState)
	envoyer_tous("reset_snapshot")
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

/* Traite un message recu de l'application de base */
func parse_app_msg(msg string) {
	type_msg := protocol.Findval(msg, "type", proc_name) // s'il retourne vide on ignore le message de toute facon

	switch type_msg {
	case "fromapp_debut_sc":
		app_dem_sc()
	case "fromapp_fin_sc":
		newData := protocol.Findval(msg, "data", proc_name)
		app_fin_sc(newData)
	case "snapshot_init":
		color = RED
		initiator = true
		localStat, _ = snapshot.ToSnapshot(protocol.Findval(msg, "snap", proc_name))

		localStat.HorlogeVect = make(map[int]int) // copie HV
		for k, v := range horloge_vect {
			localStat.HorlogeVect[k] = v
		}
		globalState = snapshot.Merge(nil, localStat)

		nbStateExpected = n_sites - 1
		nbMsgExpected = total
	case "snapshot": // Snapshot reçu de l'APP
		receiveSnapshot := protocol.Findval(msg, "snap", proc_name)
		localStat, _ = snapshot.ToSnapshot(receiveSnapshot)

		localStat.HorlogeVect = make(map[int]int)
		for k, v := range horloge_vect {
			localStat.HorlogeVect[k] = v
		}
		snapshotStr, _ := snapshot.ToString(localStat)

		display.Info(proc_name, "ROUGE", proc_name+" devient Rouge et prend une snapshot : "+snapshotStr)
		msgToSend := "state" + protocol.Msg_format("global_state", snapshotStr) + protocol.Msg_format("total", strconv.Itoa(total))
		envoyer_tous(msgToSend)

		stopSnapshot = false
		for _, m := range sauvMsg {
			handleMsg(m)
		}
		sauvMsg = nil
	default:
		return
	}
}

func envoyer(msg string, id int) {
	n_msg++
	est := Estampille{this_id, h}

	horloge_vect_str := protocol.VectToString(horloge_vect)
	sendToCtl(protocol.Msg_format("n_msg", strconv.Itoa(n_msg)) + protocol.Msg_format("target", strconv.Itoa(id)) + protocol.Msg_format("id", strconv.Itoa(est.id_site)) + protocol.Msg_format("hlg", strconv.Itoa(est.val_h)) + protocol.Msg_format("hlgvect", horloge_vect_str) + protocol.Msg_format("msg", msg) + protocol.Msg_format("color", color))
}

func envoyer_tous(msg string) {
	n_msg++
	est := Estampille{this_id, h}

	horloge_vect_str := protocol.VectToString(horloge_vect)
	// id=id_site hlg=val_h msg=msg
	sendToCtl(protocol.Msg_format("n_msg", strconv.Itoa(n_msg)) + protocol.Msg_format("id", strconv.Itoa(est.id_site)) + protocol.Msg_format("hlg", strconv.Itoa(est.val_h)) + protocol.Msg_format("hlgvect", horloge_vect_str) + protocol.Msg_format("msg", msg) + protocol.Msg_format("color", color))
}

/* Envoi aux autres le signal de liberation avec les donnes a jour*/
func envoyer_liberation(newData string) {
	n_msg++
	est := Estampille{this_id, h}
	horloge_vect_str := protocol.VectToString(horloge_vect)
	sendToCtl(protocol.Msg_format("n_msg", strconv.Itoa(n_msg)) + protocol.Msg_format("id", strconv.Itoa(est.id_site)) + protocol.Msg_format("hlg", strconv.Itoa(est.val_h)) + protocol.Msg_format("hlgvect", horloge_vect_str) + protocol.Msg_format("msg", "liberation") + protocol.Msg_format("data", newData))

}

/* Envoi un message a l'app (just stdout car l'app y est connectee) */
func sendToApp(msg_type string, value string) {
	fmt.Println(protocol.Msg_format("type", msg_type) + protocol.Msg_format("value", value))
}

/* Route le message sans modifs aux successeurs*/
func forward(msg string) {
	sendToCtl(msg)
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
	total += (n_sites - 1) // on envoie des messages à tous le monde sauf soi
	envoyer_tous("requete")
}

/* Traite une demande de fin de section critique de l'application de base */
func app_fin_sc(newData string) {
	h++
	horloge_vect[this_id]++
	map_file[this_id] = EltMapFile{"liberation", h}
	total += (n_sites - 1) // on envoie des messages à tous le monde sauf soi
	envoyer_liberation(newData)
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

func handleMsg(msg string) {
	targetId := protocol.Findval(msg, "target", proc_name)

	est, err := estampille_from_msg(msg)
	if err != nil {

	}

	// traite pas nos propre msg
	if est.id_site == this_id {
		return
	}

	nmsg_recu_str := protocol.Findval(msg, "n_msg", proc_name)
	nmsg_recu, err := strconv.Atoi(nmsg_recu_str)
	if err != nil {
		display.Error(proc_name, "erreur", "Erreur n_msg recu "+err.Error())
		return
	}

	// check si deja recu, update intervalles au passage
	if !protocol.UpdateInterval(intervalles_recus, est.id_site, nmsg_recu) {
		return
	}

	if targetId == "" || targetId == strconv.Itoa(this_id) { // si le message est pour nous ou pour tous nous on le traite
		if targetId == "" { // si le message est pour tous on le fait passer
			forward(msg)
		}
		parse_ctl_message(msg)
	} else { // si le message n'est pas pour nous on le renvoi a nos successeurs
		forward(msg)
	}
}

func sendToCtl(msg string) {
	fmt.Println(msg)
}

func main() {

	// arguments en entree
	p_nom := flag.String("n", "controler", "nom")
	p_nbsites := flag.Int("nbsites", 1, "nombre de sites")
	flag.Parse()

	// init de l'identite du site
	proc_name = *p_nom
	this_id = os.Getpid() // assigner notre pid a la variable global
	n_sites = *p_nbsites  // nombre de sites

	//initialisation de l'horloge vectorielle
	horloge_vect[this_id] = 0

	display.Info(proc_name, "main", "Démarrage du contrôleur...")

	var rcvmsg string // message recu

	for {
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error(*p_nom, "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			//return
		}
		//display.Info(*p_nom, "main", "recu : "+rcvmsg)
		if is_ctl_message(rcvmsg) { // reception d'un autre site
			if stopSnapshot {
				sauvMsg = append(sauvMsg, rcvmsg)
			} else {
				handleMsg(rcvmsg)
			}
		} else { // reception de l'application de base
			h++
			horloge_vect[this_id]++
			parse_app_msg(rcvmsg)
		}
	}
}
