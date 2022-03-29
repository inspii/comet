package internal

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

func Serve(addr string, comet Comet) {
	r := mux.NewRouter()

	h := NewHandler(comet)
	r.HandleFunc("/client/conn", h.Session)
}

type handler struct {
	comet Comet
}

func NewHandler(comet Comet) *handler {
	return &handler{comet: comet}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (p *handler) Session(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	sess := NewWSSession(r.Context(), conn, nil)

	peer := NewPeer(sess)
	if err := p.comet.AddPeer(peer); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer p.comet.RemovePeer(peer)

	sess.Wait()
}
