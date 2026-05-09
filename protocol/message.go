package protocol

import (
	"SR05_projet/display"
	"strconv"
	"strings"
)

var fieldsep = "/"
var keyvalsep = "="

type Interval struct {
	debut int
	fin   int
}

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

func UpdateIntervals(interval_map map[int][]Interval, id_site int, num int) bool {
	if len(interval_map[id_site]) == 0 {
		interval_map[id_site] = append(interval_map[id_site], Interval{num, num})
		return false
	}

	last_interval := interval_map[id_site][0]
	
	for index, inter := range interval_map[id_site] {
		if num >= inter.debut && num <= inter.fin { // dans un interval
			return true
		}

		if inter.debut - 1== num {
			inter.debut = num
			if index > 0 { // au moins 2 intervalles pr merge
					newInterval, ok := mergeIntervals(last_interval)
			}
		
			

		}


	}

}

/* verifie si c'est possible de joindre deux intervalles et le fait si c'est possible
args : deux intervales
sortie : (intervale merged, true) si joignable, (intervalle nulle, false) sinon
*/
func mergeIntervals(a Interval, b Interval) (Interval, bool){
	
	if (a.debut < b.debut && a.fin < b.debut) || (a.debut > b.debut && a.debut > b.fin) {
		return Interval{0,0}, false
	} else {
		return Interval{min(a.debut, b.debut), max(a.fin, b.fin)} , true
	}
}
// [(1,3), (5,5), (7,8)] 6 and 4 missing
// reveive 4
//  [(1,5), (7,8)]

[1, 4] [2, 5]

