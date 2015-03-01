// gzvis is a tool to assist you visualising the operation of
// the go runtime garbage collector.
//
// usage:
//
//     gcvis program [arguments]...
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/browser"
)

var (
	gcvisGraph Graph

	listener net.Listener
)

func indexHandler(w http.ResponseWriter, req *http.Request) {
	gcvisGraph.write(w)
}

func main() {
	var err error

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s command <args>...", os.Args[0])
	}

	listener, err = net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}

	pr, pw, _ := os.Pipe()
	gcChan := make(chan *gctrace, 1)
	scvgChan := make(chan *scvgtrace, 1)

	parser := Parser{
		reader:   pr,
		gcChan:   gcChan,
		scvgChan: scvgChan,
	}

	gcvisGraph = NewGraph(strings.Join(os.Args[1:], " "), GCVIS_TMPL)

	go startSubprocess(pw)
	go parser.Run()

	http.HandleFunc("/", indexHandler)

	go http.Serve(listener, nil)

	url := fmt.Sprintf("http://%s/", listener.Addr())
	log.Printf("opening browser window, if this fails, navigate to %s", url)
	browser.OpenURL(url)

	for {
		select {
		case gcTrace := <-gcChan:
			gcvisGraph.AddGCTraceGraphPoint(gcTrace)
		case scvgTrace := <-scvgChan:
			gcvisGraph.AddScavengerGraphPoint(scvgTrace)
		}
	}
}
