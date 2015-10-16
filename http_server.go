package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

type HttpServer struct {
	graph    *Graph
	listener net.Listener
	iface    string
	port     string

	listenerMtx sync.Mutex
}

func NewHttpServer(iface string, port string, graph *Graph) *HttpServer {
	h := &HttpServer{
		graph: graph,
		iface: iface,
		port:  port,
	}

	return h
}

func (h *HttpServer) Start() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		h.graph.Write(w)
	})
	http.Serve(h.Listener(), nil)
}

func (h *HttpServer) Close() {
	h.Listener().Close()
}

func (h *HttpServer) Url() string {
	return fmt.Sprintf("http://%s/", h.Listener().Addr())
}

func (h *HttpServer) Listener() net.Listener {
	h.listenerMtx.Lock()
	defer h.listenerMtx.Unlock()

	if h.listener != nil {
		return h.listener
	}

	ifaceAndPort := fmt.Sprintf("%v:%v", h.iface, h.port)
	listener, err := net.Listen("tcp4", ifaceAndPort)
	if err != nil {
		log.Fatal(err)
	}

	h.listener = listener
	return h.listener
}
