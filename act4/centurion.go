// Le centurion décide de rester dans sa tente à surveiller un tableau
//  sur lequel on vient lui écrire les événements :
//  arrivée d'un messager à l'entrée du camp,
//  évolution de la situation sur le champ de bataille...
//  Il prend toujours le premier événement,
//  le traite et revient dans sa tente à surveiller le tableau.

//go build -o centurion centurion.go
//./centurion

package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func lireMsg(canalLecture chan<- string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {

		if scanner.Scan() {
			m1 := scanner.Text()
			canalLecture <- m1
		} else {

			close(canalLecture)
			return
		}
	}
}

func main() {
	canalLecture := make(chan string)
	canalAlarme := make(chan time.Time)

	go lireMsg(canalLecture)

	delai := 5 * time.Second
	ticker := time.NewTicker(delai)
	defer ticker.Stop()

	go func() {
		for t := range ticker.C {
			canalAlarme <- t
		}
	}()

	for {

		select {
		case m1, ok := <-canalLecture:
			if !ok {
				fmt.Println("Canal lecture ferme, fin du programme.")
				return
			}

			fmt.Printf("[LECTURE] Message reçu : %s\n", m1)

		case t := <-canalAlarme:

			m2 := fmt.Sprintf("[ALARME] Délai écoulé à %s !!!!!!!!!!!!!!!!!", t.Format(time.TimeOnly))
			fmt.Fprintln(os.Stdout, m2)

		}
	}
}
