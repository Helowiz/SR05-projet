package main

import "net"

const MESSAGE = "message"
const ALARM = "alarm"
const CONNECTION = "connection"
const CTL_MESSAGE = "ctlmessage"

type Event interface {
	EventType() string
}
type MessageEvent struct {
	Content string
	Conn    net.Conn
}

type NewConnEvent struct {
	Conn net.Conn
}

type AlarmEvent struct {
}

type CtlMessageEvent struct {
	Content string
}

func (e MessageEvent) EventType() string    { return MESSAGE }
func (e AlarmEvent) EventType() string      { return ALARM }
func (e NewConnEvent) EventType() string    { return CONNECTION }
func (e CtlMessageEvent) EventType() string { return CTL_MESSAGE }
