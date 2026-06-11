package main

/* Code pour le serveur des sites

chaque site ecoute sur un serveur pour admettre des nouveaux membress


*/
import (
	"SR05_projet/protocol"
	"net"
	"strconv"
	"time"
)

var pendingAjout *net.Conn = nil

// simule une election, a avoir comment on fait
func fausse_election() bool {
	time.Sleep(5 * time.Second)
	return true
}

/*
commence un serveur pour recevoir des nouvelles connections
Quand un client se connect il créer la socket pour ce client et la donne a  handleConnection

  - args :
    adress - adresse de la machine ou on veut mettre le serveur
    port - port du serveur
    wg - un waitgroup pour attendre la fin de serveur
*/
func server(adress string, port string, connectionMap map[string]net.Conn, eventQueue chan<- Event) {
	// Listen for incoming connections
	listener, err := net.Listen("tcp", adress+":"+port)
	if err != nil {
		Error(id, "server", "Error listening on "+adress+":"+port+" :"+err.Error())
		return
	}
	defer listener.Close()

	Info(id, "server", "Server is listening at : "+adress+":"+port)

	for {
		// Accept incoming connections
		conn, err := listener.Accept()
		if err != nil {
			Error(id, "server", "Error accepting connection :"+err.Error())
			continue
		}

		if pendingAjout == nil { // on ne traite qu'un requete a la fois
			// Handle client connection in a goroutine
			handleClient(conn, connectionMap, eventQueue)
		} else {
			conn.Close()
		}
	}
}

/*
gère une connection
- lit le premier message qui doit être une demande d'admission
- fait une election
- Si c'est ok donne un id au client et continue à lire la connection
*/
func handleClient(conn net.Conn, connectionMap map[string]net.Conn, eventQueue chan<- Event) {

	// Create a buffer to read data into
	buffer := make([]byte, 1024)

	// read first message, should be a request
	n, err := conn.Read(buffer)
	if err != nil {
		Error(id, "handleClient", "Error reading from connection :"+err.Error())
		return
	}
	fmsg := string(buffer[:n])
	Info(id, "handleClient", "Premier message : "+fmsg)
	msg_type := protocol.FindvalLight(fmsg, "msg")

	switch msg_type {
	case ANNONCE:
		{
			from := protocol.FindvalLight(fmsg, "from")
			Info(id, "handleClient", "Annonce de mon voisin : "+from+" !")
			neighbor_id := from
			if _, exists := connectionMap[neighbor_id]; !exists {
				connectionMap[neighbor_id] = conn
				go read_messages(conn, eventQueue)
			} else {
				Warning(id, "handleClient", "Annonce d'un voisin déjà connu : "+neighbor_id)
				conn.Close()
			}
		}
	case DEMANDE:
		{
			Info(id, "handleClient", "Election pour voir si je peux l'admettre")
			if len(connectionMap) == 0 { // si je suis tout seul j'admet direct
				handleAdmit(conn, eventQueue)
			} else { // sinon election
				debutVagueElection()
				pendingAjout = &conn
			}
		}
	default:
		{
			Warning(id, "handleClient", "First message not request nor announce")
			return
		}
	}
}

func handleAdmit(conn net.Conn, eventQueue chan<- Event) {
	new_id := makeUniqueId()

	// TODO - mettre ca dans une fonction speciale ?
	tosend := protocol.Msg_format("num_msg", strconv.Itoa(current_msg_num)) + protocol.Msg_format("from", id) + protocol.Msg_format("msg", ADMIS) + protocol.Msg_format("your_id", new_id) + "\n"
	current_msg_num++

	_, err := conn.Write([]byte(tosend))

	if err != nil {
		Error(id, "handleAdmit", "Error writting to connection :"+err.Error())
	}
	connectionMap[new_id] = conn

	pendingAjout = nil
	broadcast(protocol.Msg_format("msg", NEW_MEMBER) + protocol.Msg_format("his_id", new_id))
	go read_messages(conn, eventQueue)
}
