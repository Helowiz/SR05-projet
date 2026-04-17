package main

import (
	"SR05_projet/display"
	"fmt"
)

func main() {
	var rcvmsg string

	for {
		_, err := fmt.Scanln(&rcvmsg)
		if err != nil {
			display.Error("Reader", "erreur", "Lecture stdin terminée ou en erreur: "+err.Error())
			return
		}
		fmt.Printf("message reçu : %s \n", rcvmsg)
	}
}
