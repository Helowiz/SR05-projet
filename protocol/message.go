package protocol

import (
	"SR05_projet/display"
	"fmt"
	"strconv"
	"strings"
)

var fieldsep = "/"
var keyvalsep = "="

type Interval struct {
	debut int
	fin   int
}

const DEMANDE = "demande"
const ANNONCE = "annonce" // Pour annoncer au démarrage que je suis le voisin d'un site
const ADMIS = "admis"
const BLEU = "bleu"
const ROUGE = "rouge"
const VAGUE = "vague"
const NEW_MEMBER = "new_member"

func Msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

func Recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func VectToString(v map[int]int) string {
	parts := []string{}
	for pid, val := range v {
		parts = append(parts, strconv.Itoa(pid)+":"+strconv.Itoa(val))
	}
	return strings.Join(parts, ",")
}

func StringToVect(s string) map[int]int {
	v := make(map[int]int)
	for _, part := range strings.Split(s, ",") {
		kv := strings.Split(part, ":")
		if len(kv) != 2 {
			continue
		}
		pid, _ := strconv.Atoi(kv[0])
		val, _ := strconv.Atoi(kv[1])
		v[pid] = val
	}
	return v
}

func RecalerVectoriel(vect1, vect2 map[int]int) map[int]int {
	for key, val := range vect2 {
		if GetVal(vect1, key) < val {
			vect1[key] = val
		}
	}
	return vect1
}

func GetVal(vect map[int]int, key int) int {
	if val, ok := vect[key]; ok {
		return val
	}
	return 0
}

func MinVectoriel(vect1, vect2 map[int]int) bool {
	for key := range vect2 {
		if GetVal(vect1, key) < GetVal(vect2, key) {
			return false
		}
	}
	return true
}

func Concurrent(vect1, vect2 map[int]int) bool {
	return !MinVectoriel(vect1, vect2) && !MinVectoriel(vect2, vect1)
}

// Fonction utilitaire permettant de convertir une chaîne de caractères au format
// "=clé=valeur" en une paire clé-valeur
func ParseEntry(entry string, name string) (string, string) {
	sep := entry[0:1]
	kv := strings.Split(entry[1:], sep)
	if len(kv) != 2 {
		display.Warning(name, "ParseEntry", "Entrée invalide : "+entry)
		return "", ""
	}
	return kv[0], kv[1]
}

//======================= Gestion des intervalles des messages =======================

func decalerGauche(l []Interval, index int) []Interval {
	for i := index; i < (len(l) - 1); i++ {
		l[i] = l[i+1]
	}
	return l[:len(l)-1]
}
func decalerDroite(l []Interval, index int) []Interval {
	l = append(l, l[len(l)-1]) // extension de 1
	for i := len(l) - 2; i > index; i-- {
		l[i] = l[i-1]

	}
	return l
}

/* Update la map des intervalles, retourne vrai le message est nouveau faux deja present */
func UpdateInterval(interval_map map[int][]Interval, id_site int, num int) bool {

	if _, ok := interval_map[id_site]; !ok { // 1er msg de ce site
		interval_map[id_site] = make([]Interval, 0)
		interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
		return true
	}

	for index, inter := range interval_map[id_site] {

		if num >= inter.debut && num <= inter.fin { // dans un interval
			return false
		}

		// extension a gauche
		if inter.debut-1 == num {
			interval_map[id_site][index] = Interval{inter.debut - 1, inter.fin}
			return true

		}

		// extension a droite
		if inter.fin+1 == num {
			interval_map[id_site][index] = Interval{inter.debut, inter.fin + 1}

			// merge avc prochain ?
			if index+1 < len(interval_map[id_site]) {
				newInterval, ok := mergeIntervals(interval_map[id_site][index], interval_map[id_site][index+1])
				if ok { // merge reussi
					interval_map[id_site][index] = newInterval
					interval_map[id_site] = decalerGauche(interval_map[id_site], index+1)
				}

			}
			return true
		}

		if num < inter.debut { // insertion entre deux intervalles
			interval_map[id_site] = decalerDroite(interval_map[id_site], index)
			interval_map[id_site][index] = Interval{num, num}
			return true
		}
	}
	// arrive en fin de liste
	interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
	return true

}

/* Update la map des intervalles, retourne vrai le message est nouveau faux deja present */
func UpdateIntervalString(interval_map map[string][]Interval, id_site string, num int) bool {

	if _, ok := interval_map[id_site]; !ok { // 1er msg de ce site
		interval_map[id_site] = make([]Interval, 0)
		interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
		return true
	}

	for index, inter := range interval_map[id_site] {

		if num >= inter.debut && num <= inter.fin { // dans un interval
			return false
		}

		// extension a gauche
		if inter.debut-1 == num {
			interval_map[id_site][index] = Interval{inter.debut - 1, inter.fin}
			return true

		}

		// extension a droite
		if inter.fin+1 == num {
			interval_map[id_site][index] = Interval{inter.debut, inter.fin + 1}

			// merge avc prochain ?
			if index+1 < len(interval_map[id_site]) {
				newInterval, ok := mergeIntervals(interval_map[id_site][index], interval_map[id_site][index+1])
				if ok { // merge reussi
					interval_map[id_site][index] = newInterval
					interval_map[id_site] = decalerGauche(interval_map[id_site], index+1)
				}

			}
			return true
		}

		if num < inter.debut { // insertion entre deux intervalles
			interval_map[id_site] = decalerDroite(interval_map[id_site], index)
			interval_map[id_site][index] = Interval{num, num}
			return true
		}
	}
	// arrive en fin de liste
	interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
	return true

}

/*
	verifie si c'est possible de joindre deux intervalles et le fait si c'est possible

args : deux intervales
sortie : (intervale merged, true) si joignable, (intervalle nulle, false) sinon
*/
func mergeIntervals(a Interval, b Interval) (Interval, bool) {

	if (a.debut < b.debut && a.fin < b.debut-1) || (a.debut > b.debut && a.debut-1 > b.fin) {
		fmt.Printf("\"Not mergeable\": %v\n", "Not mergeable")
		return Interval{0, 0}, false
	} else {
		return Interval{min(a.debut, b.debut), max(a.fin, b.fin)}, true
	}
}

func Findval(msg string, key string, name string) string {
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		equ := keyval[0:1]

		tabkeyval := strings.Split(keyval[1:], equ)
		if tabkeyval[0] == key && len(tabkeyval) == 2 {
			return tabkeyval[1]
		}

		if len(msg) < 4 {
			display.Warning(name, "FindVal", "Message trop court ou mal formaté : "+msg)
			return ""
		}
	}
	return ""
}

func FindvalLight(msg string, key string) string {
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		equ := keyval[0:1]

		tabkeyval := strings.Split(keyval[1:], equ)
		if tabkeyval[0] == key && len(tabkeyval) == 2 {
			return tabkeyval[1]
		}

		if len(msg) < 4 {
			return ""
		}
	}
	return ""
}
