package signaling

import (
	"encoding/json"
	"log"
)

type hub struct {
	rooms      map[string]map[*connection]bool
	broadcast  chan message
	register   chan subscription
	unregister chan subscription
}

var h = hub{
	broadcast:  make(chan message),
	register:   make(chan subscription),
	unregister: make(chan subscription),
	rooms:      make(map[string]map[*connection]bool),
}

func Run() {
	for {
		select {
		case sub := <-h.register:
			msgType := "joined"
			log.Println("register")
			if h.rooms[sub.Room] == nil {
				log.Println("create room " + sub.Room)
				h.rooms[sub.Room] = make(map[*connection]bool)
				msgType = "created"
			}
			h.rooms[sub.Room][sub.Con] = true
			b, err := json.Marshal(data{msgType, sub.Room})
			if err != nil {
				log.Println(err)
			}
			sub.Con.writeJSON(message{string(b), sub})
			b, err = json.Marshal(data{"join", sub.Room})
			if err != nil {
				log.Println(err)
			}
			connections := h.rooms[sub.Room]
			for con := range connections {
				if con != sub.Con {
					select {
					case con.send <- message{string(b), sub}:
					default:
						close(con.send)
						delete(connections, con)
						if len(connections) == 0 {
							delete(h.rooms, sub.Room)
						}
					}
				}
			}
		case sub := <-h.unregister:
			log.Println("unregister")
			connections := h.rooms[sub.Room]
			if connections != nil {
				if _, ok := connections[sub.Con]; ok {
					delete(connections, sub.Con)
					close(sub.Con.send)
					if len(connections) == 0 {
						delete(h.rooms, sub.Room)
					}
				}
			}
		case msg := <-h.broadcast:
			connections := h.rooms[msg.Sub.Room]
			for con := range connections {
				if con != msg.Sub.Con {
					select {
					case con.send <- msg:
					default:
						close(con.send)
						delete(connections, con)
						if len(connections) == 0 {
							delete(h.rooms, msg.Sub.Room)
						}
					}
				}
			}
		}
	}
}
