package internal

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

func Serve(addr string, comet Comet) {
	r := mux.NewRouter()

	h := NewHandler(comet)
	r.HandleFunc("/peer/conn", h.HandlePeer)
	r.HandleFunc("/service/conn", h.HandleService)
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

func (p *handler) HandlePeer(w http.ResponseWriter, r *http.Request) {
	protocol := r.Header.Get("Comet-Protocol")
	service := r.Header.Get("Comet-Service")
	serviceToken := r.Header.Get("Comet-Service-Token")
	clientID := r.Header.Get("Comet-Client-ID")
	ip := r.RemoteAddr

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	sess := NewWSSession(r.Context(), conn, nil)

	info := PeerInfo{
		Protocol:     protocol,
		ClientID:     clientID,
		IP:           ip,
		Service:      service,
		ServiceToken: serviceToken,
	}
	peer, err := p.comet.NewPeer(sess, info)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if err := p.comet.AddPeer(peer); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer p.comet.RemovePeer(peer)

	sess.Wait()
}

func (p *handler) HandleService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.Header.Get("Comet-Service")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	service, ok := p.comet.GetService(serviceName)
	if !ok {
		w.Write([]byte(err.Error()))
		return
	}

	sess := NewWSSession(r.Context(), conn, nil)
	serviceSess := NewServiceWorker(sess)
	if err := service.AddWorker(serviceSess); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer service.RemoveWorker(serviceSess)

	sess.Wait()
}
