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

func (s subscription) readPump() {
	c := s.Con
	defer func() {
		h.unregister <- s
		c.ws.Close()
	}()
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
		h.broadcast <- message{string(msg), s}
	}
}

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
