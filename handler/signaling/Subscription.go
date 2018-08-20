package signaling

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type data struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type message struct {
	Data string       `json:"data"`
	Sub  subscription `json:"sub"`
}

type subscription struct {
	Con  *connection `json:"con"`
	Room string      `json:"room"`
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump() {
	c := s.Con
	log.Println(c)
	defer func() {
		h.unregister <- s
		c.ws.Close()
	}()
	//c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("UnexplectedCloseError: %v", err)
			}
			break
		}
		//h.broadcast <- message{msg, s.room}
		h.broadcast <- message{string(msg), s}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.Con
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			log.Println("Send: " + msg.Data)
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.writeJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
