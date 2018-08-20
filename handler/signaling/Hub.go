package signaling

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/pions/webrtc"
)

type rtcsessiondescription struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}

// h maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

var peerConnection *webrtc.RTCPeerConnection

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
			log.Println("broadcast")
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

func RunMediaServer() {
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
			peerConnection, conErr := webrtc.New(webrtc.RTCConfiguration{
				ICEServers: []webrtc.RTCICEServer{
					{
						URLs: []string{"stun:stun.l.google.com:19302"},
					},
				},
			})
			if conErr != nil {
				log.Fatal(conErr)
			}
			offer, offerErr := peerConnection.CreateOffer(nil)
			if offerErr != nil {
				log.Fatal(offerErr)
			}
			log.Println(offer)
			o := rtcsessiondescription{offer.Type.String(), offer.Sdp}
			b, err = json.Marshal(o)
			if err != nil {
				log.Println(err)
			}

			if err := sub.Con.writeJSON(message{string(b), sub}); err != nil {
				log.Fatal(err)
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
			log.Println("broadcast")
			log.Println(msg.Data)
			if strings.Contains(msg.Data, "answer") {
				var tsd rtcsessiondescription
				if err := json.Unmarshal([]byte(msg.Data), &tsd); err != nil {
					log.Fatal(err)
				}

			} else {
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
}
