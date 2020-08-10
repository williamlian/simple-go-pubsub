package ws

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type subHandler struct{}

type wsConnType struct {
	conn     *websocket.Conn
	isClosed bool
}

var upgrader = websocket.Upgrader{}

/* ServeHTTP implements http.Handler
Subscriber handler takes last path as topic name. */
func (h *subHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.Path, "/")
	topic := paths[len(paths)-1]
	if topic == "" {
		log.Print("Error: empty topic")
		w.WriteHeader(400)
		fmt.Fprint(w, "Error: empty topic")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed upgrading to Websocket. error: %s", err.Error())
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error: failed upgrade to Websocket, error: %s", err.Error())
		return
	}

	wsConn := new(wsConnType)
	wsConn.conn = conn
	wsConn.isClosed = false

	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("client request close. [%d] %s", code, text)
		wsConn.isClosed = true
		return nil
	})

	pubsubSingle.Subscribe(topic, wsConn)

	// start a periodical reader to response with ping or close
	for {
		_, _, err := conn.NextReader()
		if err != nil {
			// we are done serving this connection
			conn.Close()
			return
		}
	}
}

// OnData implements Subscriber interface
func (h *wsConnType) OnData(data string) bool {
	if h.isClosed {
		log.Printf("Connection has been marked as closed")
		return true
	}
	err := h.conn.WriteMessage(websocket.TextMessage, []byte(data))
	if err != nil {
		log.Printf("Failed writing WS, error: %s. Unsubscribe.", err.Error())
		h.conn.Close()
		return true
	}
	return false
}
