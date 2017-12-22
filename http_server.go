package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
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
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		h.graph.Write(w)
	})

	serveMux.HandleFunc("/graph.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(h.graph); err != nil {
			log.Fatalf("An error occurred while serving JSON endpoint: %v", err)
		}
	})

	server := http.Server{
		Handler:      serveMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.Serve(h.Listener())
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
