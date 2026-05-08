package protocol

import (
	"SR05_projet/display"
	"strings"
)

var fieldsep = "/"
var keyvalsep = "="

func Msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}

// TODO maybe overwrite
func Msg_format_key(key string, keysep string, val string, fieldseparator string) string {
	return fieldseparator + keysep + key + keysep + val
}

func Recaler(x, y int) int {
	if x < y {
		return y + 1
	}
	return x + 1
}

func Findval(msg string, key string, name string) string {
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)

	for _, keyval := range tab_allkeyvals {
		equ := keyval[0:1]

		tabkeyval := strings.Split(keyval[1:], equ)
		if tabkeyval[0] == key {
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
