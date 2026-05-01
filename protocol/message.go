package protocol

import (
	"SR05_projet/display"
	"strconv"
	"strings"
)

var fieldsep = "/"
var keyvalsep = "="

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
