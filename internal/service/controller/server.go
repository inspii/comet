package controller

import (
	"github.com/gorilla/mux"
	"github.com/inspii/comet/internal/service"
)

func Serve(addr string, comet service.Comet) {
	r := mux.NewRouter()

	h := NewPeerHandler(comet)
	r.HandleFunc("/client/conn", h.WS)
}
