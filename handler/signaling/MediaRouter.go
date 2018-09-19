package signaling

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/ice"
	"github.com/pions/webrtc/pkg/media/samplebuilder"
)

type rtcsessiondescription struct {
	Type     string `json:"type"`
	Sdp      string `json:"sdp"`
	ClientID string `json:"clientId"`
}

var receiver = make(map[string][]*webrtc.RTCTrack)

func buildPeerConnection(sub subscription) *webrtc.RTCPeerConnection {
	webrtc.RegisterCodec(webrtc.NewRTCRtpVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	connection, err := webrtc.New(webrtc.RTCConfiguration{})

	if err != nil {
		panic(err)
	}

	connection.Ontrack = func(track *webrtc.RTCTrack) {
		connections := h.rooms[sub.Room]
		for con := range connections {
			if _, ok := con.receiverPeers[sub.Con.ClientID]; !ok && con.ClientID != sub.Con.ClientID {
				receiverTrack, err := con.peer.NewRTCTrack(track.PayloadType, sub.Con.ClientID, "video")

				if err != nil {
					log.Println(err)
				}

				c := buildPeerConnection(sub)

				c.AddTrack(receiverTrack)

				receiver[sub.Con.ClientID] = append(receiver[sub.Con.ClientID], receiverTrack)
				con.receiverPeers[sub.Con.ClientID] = c

				log.Println(receiver)
				offer, offErr := c.CreateOffer(nil)
				if offErr != nil {
					log.Fatal(offErr)
				}
				log.Println(con.ClientID)
				a := rtcsessiondescription{offer.Type.String(), offer.Sdp, sub.Con.ClientID}
				b, err := json.Marshal(a)
				if err != nil {
					log.Fatal(err)
				}
				if err := con.writeJSON(message{string(b), sub}); err != nil {
					log.Fatal(err)
				}
			}
		}
		log.Println("handle Packets")
		builder := samplebuilder.New(256)

		for {
			builder.Push(<-track.Packets)
			for s := builder.Pop(); s != nil; s = builder.Pop() {
				for _, receiver := range receiver[sub.Con.ClientID] {
					receiver.Samples <- *s
				}
			}
		}
	}

	connection.OnICEConnectionStateChange = func(connectionState ice.ConnectionState) {
		log.Println("Connection State has changed", connectionState.String())
	}

	return connection
}

func RunMediaRouter() {
	for {
		select {
		case sub := <-h.register:
			msgType := "joined"
			log.Println("register")
			if h.rooms[sub.Room] == nil {
				log.Println("create room " + sub.Room)
				h.rooms[sub.Room] = make(map[*connection]bool)
			}
			h.rooms[sub.Room][sub.Con] = true
			b, err := json.Marshal(data{msgType, sub.Room})
			if err != nil {
				log.Fatal(err)
			}
			sub.Con.writeJSON(message{string(b), sub})

			sub.Con.peer = buildPeerConnection(sub)
			sub.Con.receiverPeers = make(map[string]*webrtc.RTCPeerConnection)

			log.Println("PeerConnection created")
			log.Println(sub.Con.peer.GetTransceivers())
			offer, offErr := sub.Con.peer.CreateOffer(nil)
			if offErr != nil {
				log.Fatal(offErr)
			}

			log.Println(offer.Sdp)

			a := rtcsessiondescription{offer.Type.String(), offer.Sdp, sub.Con.ClientID}
			b, err = json.Marshal(a)
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
			log.Println("ClientID: ", msg.Sub.Con.ClientID)
			log.Println(msg.Data)
			if strings.Contains(msg.Data, "answer") {
				log.Println("answer")
				var tsd rtcsessiondescription
				if err := json.Unmarshal([]byte(msg.Data), &tsd); err != nil {
					log.Fatal(err)
				}
				log.Println("Unmarshal json")
				if tsd.ClientID != msg.Sub.Con.ClientID {
					if _, ok := msg.Sub.Con.receiverPeers[tsd.ClientID]; ok {
						if err := msg.Sub.Con.receiverPeers[tsd.ClientID].SetRemoteDescription(webrtc.RTCSessionDescription{
							Type: webrtc.RTCSdpTypeAnswer,
							Sdp:  tsd.Sdp,
						}); err != nil {
							log.Fatal(err)
						}
					}
				} else {
					if err := msg.Sub.Con.peer.SetRemoteDescription(webrtc.RTCSessionDescription{
						Type: webrtc.RTCSdpTypeAnswer,
						Sdp:  tsd.Sdp,
					}); err != nil {
						log.Fatal(err)
					}
				}
			} else if strings.Contains(msg.Data, "offer") {
				var tsd rtcsessiondescription
				if err := json.Unmarshal([]byte(msg.Data), &tsd); err != nil {
					log.Fatal(err)
				}
				msg.Sub.Con.peer = buildPeerConnection(msg.Sub)
				msg.Sub.Con.peer.SetRemoteDescription(webrtc.RTCSessionDescription{
					Type: webrtc.RTCSdpTypeAnswer,
					Sdp:  tsd.Sdp,
				})
				log.Println("PeerConnection created")

				answer, ansErr := msg.Sub.Con.peer.CreateAnswer(nil)
				if ansErr != nil {
					log.Fatal(ansErr)
				}

				log.Println(answer.Sdp)

				a := rtcsessiondescription{answer.Type.String(), answer.Sdp, tsd.ClientID}
				b, err := json.Marshal(a)
				if err != nil {
					log.Println(err)
				}

				if err := msg.Sub.Con.writeJSON(message{string(b), msg.Sub}); err != nil {
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
