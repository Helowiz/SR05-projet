package protocol

import (
	"SR05_projet/display"
	"strconv"
	"strings"
)

var fieldsepCtrl = "/"
var keyvalsepCtrl = "="

const DEMANDE = "demande"
const ANNONCE = "annonce" // Pour annoncer au démarrage que je suis le voisin d'un site
const ADMIS = "admis"
const BLEU = "bleu"
const ROUGE = "rouge"
const VAGUE = "vague"
const NEW_MEMBER = "new_member"
const LEAVE = "leave_msg"

func Findval(msg string, key string) string {
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

func Msg_format_Ctrl(key string, val string) string {

	return fieldsepCtrl + keyvalsepCtrl + key + keyvalsepCtrl + val
}

func Recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func VectToString(v map[string]int) string {
	parts := []string{}
	for id, val := range v {
		parts = append(parts, id+":"+strconv.Itoa(val))
	}
	return strings.Join(parts, ",")
}

func StringToVect(s string) map[string]int {
	v := make(map[string]int)
	for _, part := range strings.Split(s, ",") {
		kv := strings.Split(part, ":")
		if len(kv) != 2 {
			continue
		}
		id := kv[0]
		val, _ := strconv.Atoi(kv[1])
		v[id] = val
	}
	return v
}

func RecalerVectoriel(vect1, vect2 map[string]int) map[string]int {
	for key, val := range vect2 {
		if GetVal(vect1, key) < val {
			vect1[key] = val
		}
	}
	return vect1
}

func GetVal(vect map[string]int, key string) int {
	if val, ok := vect[key]; ok {
		return val
	}
	return 0
}

func MinVectoriel(vect1, vect2 map[string]int) bool {
	for key := range vect2 {
		if GetVal(vect1, key) < GetVal(vect2, key) {
			return false
		}
	}
	return true
}

func Concurrent(vect1, vect2 map[string]int) bool {
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
