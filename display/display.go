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
)

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

func baseDisplay(color string, level string, name string, where string, what string) {
	stderr.Printf("%s* [%.6s %d] %-8.8s [%-7.7s] : %s%s\n", color, name, pid, where, level, what, Reset)
}

func Info(name string, where string, what string) {
	baseDisplay(Green, "INFO", name, where, what)
}

func Warning(name string, where string, what string) {
	baseDisplay(Orange, "WARNING", name, where, what)
}

func Error(name string, where string, what string) {
	baseDisplay(Red, "ERROR", name, where, what)
}

func Envoie(name string, where string, what string) {
	baseDisplay(Purple, "ENVOIE", name, where, what)
}

func Recu(name string, where string, what string) {
	baseDisplay(Cyan, "RECU", name, where, what)
}
