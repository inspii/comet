package controller

import (
	"github.com/gorilla/websocket"
	"github.com/inspii/comet/internal/service"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type peerHandler struct {
	comet service.Comet
}

func NewPeerHandler(comet service.Comet) *peerHandler {
	return &peerHandler{comet: comet}
}

type readWriter struct {
	conn             *websocket.Conn
	readTimeout      time.Duration
	writeTimeout     time.Duration
	pingPongInterval time.Duration
	pings            int
	pongs            int
}

func newReadWriter(conn *websocket.Conn) (*readWriter, error) {
	rw := &readWriter{
		conn: conn,
	}
	if err := rw.pong(); err != nil {
		return nil, err
	}
	go rw.ping()

	return rw, nil
}

func (r *readWriter) Read(p []byte) (int, error) {
	_, reader, err := r.conn.NextReader()
	if err != nil {
		return 0, err
	}
	deadline := time.Now().Add(r.readTimeout)
	if err := r.conn.SetReadDeadline(deadline); err != nil {
		return 0, err
	}

	return reader.Read(p)
}

func (r *readWriter) Write(data []byte) (int, error) {
	deadline := time.Now().Add(r.writeTimeout)
	if err := r.conn.SetWriteDeadline(deadline); err != nil {
		_ = r.conn.WriteMessage(websocket.CloseMessage, []byte{})
		return 0, err
	}

	writer, err := r.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	n, err := writer.Write(data)
	if err != nil {
		return n, err
	}

	err = writer.Close()
	return n, err
}

func (r *readWriter) pong() error {
	if err := r.conn.SetReadDeadline(time.Now().Add(r.pingPongInterval)); err != nil {
		return err
	}

	r.conn.SetPongHandler(func(string) error {
		return nil
	})
	return nil
}

func (r *readWriter) ping() {
	ticker := time.NewTicker(r.pingPongInterval)
	for {
		select {
		case <-ticker.C:
			deadline := time.Now().Add(r.writeTimeout)
			err := r.conn.SetWriteDeadline(deadline)
			if err != nil {
				return
			}
			if err := r.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			r.pings++
		}
	}
}

func (p *peerHandler) WS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	rw, err := newReadWriter(conn)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	peer := service.NewPeer(rw)
	p.comet.AddPeer(peer)
}
