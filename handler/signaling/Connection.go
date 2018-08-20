package signaling

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	ClientID string `json:"clientId"`

	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan message
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *connection) writeJSON(message message) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(message)
}

// WSConnection handles websocket requests from the peer.
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
