package signaling

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/ice"
	"github.com/pions/webrtc/pkg/media"
	"github.com/pions/webrtc/pkg/media/samplebuilder"
	"github.com/pions/webrtc/pkg/rtcp"
	"github.com/pions/webrtc/pkg/rtp/codecs"
)

//TODO: Auslagern in separates package
type rtcsessiondescription struct {
	Type     string `json:"type"`
	Sdp      string `json:"sdp"`
	ClientID string `json:"clientId"`
}

const (
	rtcpPLIInterval = time.Second * 3
)

var peerConnectionConfig = webrtc.RTCConfiguration{
	IceServers: []webrtc.RTCIceServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

var receiver = make(map[string][]chan<- media.RTCSample)
var receiverLock = make(map[string]*sync.RWMutex)

func buildPeerConnection(sub subscription) *webrtc.RTCPeerConnection {
	webrtc.RegisterCodec(webrtc.NewRTCRtpVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	connection, err := webrtc.New(peerConnectionConfig)

	if err != nil {
		panic(err)
	}

	connection.OnTrack = func(track *webrtc.RTCTrack) {
		log.Println("new Track")
		log.Println(track.PayloadType)
		log.Println(track.ID)
		connections := h.rooms[sub.Room]

		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for {
				select {
				case <-ticker.C:
					err := connection.SendRTCP(&rtcp.PictureLossIndication{MediaSSRC: track.Ssrc})
					if err != nil {
						log.Println(err)
					}
				}
			}
		}()
		//TODO: Entfernen / In separate Funktion auslagern
		log.Println("Create Receiver Tracks")
		log.Println(connections)
		for con := range connections {
			log.Println("Peer Client ID: ", sub.Con.ClientID)
			log.Println("Potential Receiver Client ID: ", con.ClientID)
			if _, ok := con.receiverPeers[sub.Con.ClientID]; !ok && con.ClientID != sub.Con.ClientID {
				log.Println("Create new Receiver Track for Client ", con.ClientID)
				receiverTrack, err := con.peer.NewRTCTrack(track.PayloadType, "video", sub.Con.ClientID)

				if err != nil {
					log.Println(err)
				}

				c := buildPeerConnection(sub)

				c.AddTrack(receiverTrack)

				receiverLock[sub.Con.ClientID].Lock()
				receiver[sub.Con.ClientID] = append(receiver[sub.Con.ClientID], receiverTrack.Samples)
				receiverLock[sub.Con.ClientID].Unlock()

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
		builder := samplebuilder.New(256, &codecs.VP8Packet{})
		packetCount := 0
		for {
			receiverLock[sub.Con.ClientID].RLock()
			builder.Push(<-track.Packets)
			for s := builder.Pop(); s != nil; s = builder.Pop() {
				//log.Println("New Packet")
				//log.Println(outboundSamples)
				for i, outChan := range receiver[sub.Con.ClientID] {
					packetCount = packetCount + 1
					log.Println(i, ": Send Packet ", packetCount)
					outChan <- *s
				}
			}
			receiverLock[sub.Con.ClientID].RUnlock()
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
			log.Println("register")

			msgType := "joined"

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

			offer, offErr := sub.Con.peer.CreateOffer(nil)
			if offErr != nil {
				log.Fatal(offErr)
			}

			a := rtcsessiondescription{offer.Type.String(), offer.Sdp, sub.Con.ClientID}
			b, err = json.Marshal(a)
			if err != nil {
				log.Println(err)
			}

			if err := sub.Con.writeJSON(message{string(b), sub}); err != nil {
				log.Fatal(err)
			}
			//TODO: Broadcast an bestehende Peers per Join-Message
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
			if strings.Contains(msg.Data, "answer") {
				log.Println("answer")
				var tsd rtcsessiondescription
				if err := json.Unmarshal([]byte(msg.Data), &tsd); err != nil {
					log.Fatal(err)
				}
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
				log.Println("offer")
				//TODO: Unterscheidung Receive / Send Connection (Unterschied ClientID)
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

				a := rtcsessiondescription{answer.Type.String(), answer.Sdp, tsd.ClientID}
				b, err := json.Marshal(a)
				if err != nil {
					log.Println(err)
				}

				if err := msg.Sub.Con.writeJSON(message{string(b), msg.Sub}); err != nil {
					log.Fatal(err)
				}
				//TODO: Aufsetzen der Verbindungen für die bestehenden Peers per Join-Message
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
