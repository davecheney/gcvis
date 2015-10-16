// gcvis is a tool to assist you visualising the operation of
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
	"os"
	"strings"

	"github.com/pkg/browser"
)

var iface = flag.String("i", "127.0.0.1", "specify interface to use. defaults to 127.0.0.1.")
var port = flag.String("p", "0", "specify port to use.")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: command <args>...\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		return
	}

	subcommand := NewSubCommand(flag.Args())
	parser := NewParser(subcommand.PipeRead)
	gcvisGraph := NewGraph(strings.Join(flag.Args(), " "), GCVIS_TMPL)
	server := NewHttpServer(*iface, *port, &gcvisGraph)

	go subcommand.Run()
	go parser.Run()
	go server.Start()

	url := server.Url()
	log.Printf("opening browser window, if this fails, navigate to %s", url)
	browser.OpenURL(url)

	for {
		select {
		case gcTrace := <-parser.GcChan:
			gcvisGraph.AddGCTraceGraphPoint(gcTrace)
		case scvgTrace := <-parser.ScvgChan:
			gcvisGraph.AddScavengerGraphPoint(scvgTrace)
		case output := <-parser.NoMatchChan:
			fmt.Fprintln(os.Stderr, output)
		case <-parser.done:
			if parser.Err != nil {
				fmt.Fprintf(os.Stderr, parser.Err.Error())
				os.Exit(1)
			}

			if subcommand.Err() != nil {
				fmt.Fprintf(os.Stderr, subcommand.Err().Error())
				os.Exit(1)
			}

			os.Exit(0)
		}
	}
}
