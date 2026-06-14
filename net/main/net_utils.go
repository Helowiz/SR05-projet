package main

import (
	"net"
	"strconv"
)

/*
getBroadcastAddress

TODO - Mieux documenter cette fonction
Retourne l'adresse de braodcast du réseau.
Pour ce faire, scan toutes les interfaces de la machine pour trouver une adresse qui convient.
*/
func getBroadcastAddress() (string, error) {
	return "255.255.255.255", nil

}

/*
getPrivatePort

Retourne le port de notre socket pour les connexions ciblées sous forme de string.
Retourne vide si socket ciblé pas encore init.
*/
func getPrivatePort() string {
	if G.SocketDirect == nil {
		return ""
	}
	return strconv.Itoa(G.SocketDirect.LocalAddr().(*net.UDPAddr).Port)
}
