package signaling

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pions/webrtc"
	uuid "github.com/satori/go.uuid"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type connection struct {
	ClientID      string `json:"clientId"`
	ws            *websocket.Conn
	peer          *webrtc.RTCPeerConnection
	receiverPeers map[string]*webrtc.RTCPeerConnection
	send          chan message
}

func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writeJSON(message message) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(message)
}

func WSConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ws, wsErr := upgrader.Upgrade(w, r, nil)
	if wsErr != nil {
		log.Println(wsErr)
		return
	}

	uID, idErr := uuid.NewV4()
	if idErr != nil {
		log.Println(idErr)
		return
	}

	c := &connection{send: make(chan message, 256), ws: ws, ClientID: uID.String()}
	s := subscription{c, vars["room"]}
	h.register <- s
	go s.writePump()
	s.readPump()
}
