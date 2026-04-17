package main

import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	p_nom := flag.String("n", "ecrivain", "nom")
	flag.Parse()
	nom := *p_nom + "-" + strconv.Itoa(os.Getpid())
	for {
		msg := protocol.Msg_format("sender", nom) + protocol.Msg_format("msg", "prout")
		fmt.Println(msg)
		display.Info(*p_nom, "msg_send", "Émission du message : "+msg)
		time.Sleep(time.Duration(2) * time.Second)
	}
}
