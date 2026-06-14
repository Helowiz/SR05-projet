package main

/* Code pour le serveur des sites

chaque site ecoute sur un serveur pour admettre des nouveaux membress


*/
import (
	"SR05_projet/display"
	"SR05_projet/protocol"
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// config serveur

func init_sockets(broadcast_port string, eventQueue chan<- Event) {
	var err error
	display.Info(G.Id, "init_socket", "Init du socket")

	// setup targeted receiving

	direct_socket_adr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:0") // laisse le OS choisir le port
	if err != nil {
		display.Error("", "initSockets", "Failed to resolve adress")
		panic(err)
	}

	G.SocketDirect, err = net.ListenUDP("udp4", direct_socket_adr)
	if err != nil {
		display.Error("", "initSockets", "Failed to listen to socketDirect")
		panic(err)
	}
	// si on est en mode dev, il faut qu'on puisse réutiliser le port 8080
	// setup broadcast receiving
	if G.Dev_mode {
		lc := net.ListenConfig{
			Control: func(network, address string, c syscall.RawConn) error {
				return c.Control(func(fd uintptr) {
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
					syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				})
			},
		}

		conn, err := lc.ListenPacket(context.Background(), "udp4", "0.0.0.0"+":"+broadcast_port)
		if err != nil {
			panic(err)
		}
		G.SocketBroadcast = conn.(*net.UDPConn)
		//broadcast_adr, err = net.ResolveUDPAddr("udp", "localhost"+":"+port)

	} else {

		broadcast_socket_adr, err := net.ResolveUDPAddr("udp4", "0.0.0.0"+":"+broadcast_port)
		if err != nil {
			display.Error("", "initSockets", "Failed to resolve adress")
			panic(err)
		}
		G.SocketBroadcast, err = net.ListenUDP("udp4", broadcast_socket_adr)

		if err != nil {
			display.Error("", "initSockets", "Failed to start listening")
			panic(err)
		}

	}

	G.Broadcast_adr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+broadcast_port)
	if err != nil {
		display.Error("", "init_socket", "Failed to resolve broadcast adress")
		panic(err)
	}
	// on lit des deux sockets
	go read_messages(G.SocketBroadcast, eventQueue)
	go read_messages(G.SocketDirect, eventQueue)
	display.Info(G.Id, "init_socket", "Socket init good")

}

func read_messages(connection *net.UDPConn, eventQueue chan<- Event) {
	defer connection.Close()
	display.Info(G.Id, "read_messages", "En attente de messages sur : "+connection.LocalAddr().String())
	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			display.Error("", "read_messages", "ReadFromUDP failed : "+err.Error())
			panic(err)
		}
		line := string(buffer[:n])
		from := protocol.Findval(line, "from")

		if from == G.Id { // traite pas mes propres msg
			continue
		}

		// maj map de connection pour pas devoir broadcast la prochaine fois que je lui envoi
		_, connu := G.ConnectionMap[from]
		if from != "" && !connu {
			G.ConnectionMap[from] = clientAddr
		}
		display.Info(G.Id, "read_messages", "Lu : "+line+" de "+clientAddr.IP.String())
		eventQueue <- MessageEvent{Adress: clientAddr, Content: line}

	}

}
