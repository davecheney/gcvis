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
	"io"
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

	var pipeRead io.ReadCloser
	var subcommand SubCommand

	flag.Parse()
	if len(flag.Args()) < 1 {
		pipeRead = os.Stdin
	} else {
		subcommand := NewSubCommand(flag.Args())
		pipeRead = subcommand.PipeRead
		go subcommand.Run()
	}

	parser := NewParser(pipeRead)
	gcvisGraph := NewGraph(strings.Join(flag.Args(), " "), GCVIS_TMPL)
	server := NewHttpServer(*iface, *port, &gcvisGraph)

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

			break
		}
	}

	if subcommand.cmd != nil && subcommand.Err() != nil {
		fmt.Fprintf(os.Stderr, subcommand.Err().Error())
		os.Exit(1)
	}

	os.Exit(0)
}
