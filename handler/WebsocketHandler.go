package handler

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
	ws   *websocket.Conn
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var queue = make(chan Message)
var upgrader = websocket.Upgrader{}

func WSConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		msg.ws = ws
		queue <- msg
	}
}

func WSMessages() {
	for {
		msg := <-queue
		log.Println(msg.Type)
		switch msg.Type {
		case "create or join":
			createOrJoin(msg)
		case "ipaddr":
			sendIPAddr(msg)
		case "message":
			broadcast(msg)
		}
	}
}

func broadcast(msg Message) {
	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Fatal(err)
			client.Close()
			delete(clients, client)
		}
	}
}

func sendIPAddr(msg Message) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil {
				msg.ws.WriteJSON(Message{
					"ipaddr",
					string(ip),
					nil,
				})
			}
		}
	}
}

func createOrJoin(msg Message) {
	log.Println("Func createOrJoin")
	var msgType string
	if len(clients) == 0 {
		log.Println("0 clients")
		msgType = "created"
	} else if len(clients) == 1 {
		log.Println("1 client")
		broadcast(Message{
			"join",
			msg.Text,
			nil,
		})
		msgType = "joined"
	} else {
		log.Println("too many clients")
		msgType = "full"
	}
	log.Println("send Msg")
	msg.ws.WriteJSON(Message{
		msgType,
		msg.Text,
		nil,
	})
}
