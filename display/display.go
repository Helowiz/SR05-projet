package display

import (
	"log"
	"os"
)

var (
	Red    = "\033[1;31m"
	Orange = "\033[1;33m"
	Green  = "\033[1;32m"
	Reset  = "\033[0;00m"
	Purple = "\033[1;35m"
	Cyan   = "\033[1;36m"
	Pink   = "\033[1;38;5;206m"
	Blue   = "\033[1;34m"
)

var stderr = log.New(os.Stderr, "", 0)
var pid = os.Getpid()

func SimpleIdShowing(id string) string {
	if len(id) < 4 {
		return id
	}
	return id[0:4]
}

func baseDisplay(color string, level string, name string, where string, what string) {
	stderr.Printf("%s* [%-4.4s %d] %-8.8s [%-7.7s] : %s%s\n", color, SimpleIdShowing(name), pid, where, level, what, Reset)
}

func Vague(name string, where string, what string) {
	baseDisplay(Cyan, "VAGUE", name, where, what)
}

func Info(name string, where string, what string) {
	baseDisplay(Green, "DISPLAY", name, where, what)
}

func Warning(name string, where string, what string) {
	baseDisplay(Orange, "WARNING", name, where, what)
}

func Error(name string, where string, what string) {
	baseDisplay(Red, "ERROR", name, where, what)
}

func Leave(name string, where string, what string) {
	baseDisplay(Pink, "LEAVE", name, where, what)
}

func Envoie(name string, where string, what string) {
	baseDisplay(Purple, "ENVOIE", name, where, what)
}

func Recu(name string, where string, what string) {
	baseDisplay(Blue, "RECU", name, where, what)
}
