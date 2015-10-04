// gzvis is a tool to assist you visualising the operation of
// the go runtime garbage collector.
//
// usage:
//
//     gcvis program [arguments]...
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/browser"
)

var gcvisGraph Graph

func indexHandler(w http.ResponseWriter, req *http.Request) {
	gcvisGraph.write(w)
}

var iface = flag.String("i", "127.0.0.1", "specify interface to use.  defaults to 127.0.0.1.")

func main() {

	var err error

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: command <args>...\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	ifaceAndPort := fmt.Sprintf("%v:0", *iface)
	listener, err := net.Listen("tcp4", ifaceAndPort)
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

	gcvisGraph = NewGraph(strings.Join(flag.Args(), " "), GCVIS_TMPL)

	go startSubprocess(pw, flag.Args())
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
