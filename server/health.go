package server

import (
	"net/http"

	gorillaMux "github.com/gorilla/mux"
	k8sApiServerHealthz "k8s.io/apiserver/pkg/server/healthz"
)

func registerHealth(router *gorillaMux.Router) {
	k8sApiServerHealthz.InstallHandler(muxWrapper{mux: router})
	router.Handle("/ping", http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("pong"))
	}))
}

type muxWrapper struct {
	mux *gorillaMux.Router
}

func (m muxWrapper) Handle(path string, handler http.Handler) {
	m.mux.Handle(path, handler)
}
