package main

import (
	"bufio"
	"net"
	"strings"
	"time"

	"SR05_projet/protocol"
)

func read_messages(conn net.Conn, eventQueue chan<- Event) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		read := scanner.Text()
		Info(id, "read_messages", "Received : "+read)
		eventQueue <- MessageEvent{read, conn}
	}
	Info(id, "read_messages", "Connection closed, stopping read_messages")

}

func client(adress string, port string, connectionMap map[string]net.Conn, eventQueue chan<- Event, alreadyAdmitted bool) {

	msg_type := ""
	if alreadyAdmitted {
		msg_type = ANNONCE
	} else {
		msg_type = DEMANDE
	}
	Info(id, "client", "Client : connection au ctl NET "+adress+":"+port+" et envoi de "+msg_type)
	// Connect to the server
	conn, err := net.Dial("tcp", adress+":"+port)
	for err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			Warning(id, "client", "Ctl NET "+adress+":"+port+" not up yet, retrying later...")
		}
		Error(id, "client", "Dial error in client :"+err.Error())
		if !alreadyAdmitted {
			return
		} else {
			time.Sleep(2 * time.Second)
			conn, err = net.Dial("tcp", adress+":"+port)
		}
	}

	demande_msg := protocol.Msg_format("msg", msg_type)
	if alreadyAdmitted {
		demande_msg = demande_msg + protocol.Msg_format("from", id)
	}
	conn.Write([]byte(demande_msg))

	go read_messages(conn, eventQueue)
}
