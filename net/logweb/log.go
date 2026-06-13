package logweb

import (
	"SR05-etude/protocol"
	"fmt"
)

const LOG = "log"

func sendToLog(msg string) {
	fmt.Println(protocol.Msg_format_Ctrl("to", LOG) + msg)
}

func Vague(id string, from string, to string, parent string, elu string, color string, nbvoisinsAttendus string) {
	msg := protocol.Msg_format_Ctrl("id", id) +
		protocol.Msg_format_Ctrl("from", from) +
		protocol.Msg_format_Ctrl("to", to) +
		protocol.Msg_format_Ctrl("parent", parent) +
		protocol.Msg_format_Ctrl("elu", elu) +
		protocol.Msg_format_Ctrl("color", color) +
		protocol.Msg_format_Ctrl("nbvoisinsAttendus", nbvoisinsAttendus)
	sendToLog(msg)
}

func Admis() {}
