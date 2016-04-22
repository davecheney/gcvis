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

	"golang.org/x/crypto/ssh/terminal"
)

var iface = flag.String("i", "127.0.0.1", "specify interface to use. defaults to 127.0.0.1.")
var port = flag.String("p", "0", "specify port to use.")
var openBrowser = flag.Bool("o", true, "automatically open browser")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: command <args>...\n", os.Args[0])
		flag.PrintDefaults()
	}

	var pipeRead io.ReadCloser
	var subcommand *SubCommand

	flag.Parse()
	if len(flag.Args()) < 1 {
		if terminal.IsTerminal(int(os.Stdin.Fd())) {
			flag.Usage()
			return
		} else {
			pipeRead = os.Stdin
		}
	} else {
		subcommand = NewSubCommand(flag.Args())
		pipeRead = subcommand.PipeRead
		go subcommand.Run()
	}

	parser := NewParser(pipeRead)

	title := strings.Join(flag.Args(), " ")
	if len(title) == 0 {
		title = fmt.Sprintf("%s:%s", *iface, *port)
	}

	gcvisGraph := NewGraph(title, GCVIS_TMPL)
	server := NewHttpServer(*iface, *port, &gcvisGraph)

	go parser.Run()
	go server.Start()

	url := server.Url()

	if *openBrowser {
		log.Printf("opening browser window, if this fails, navigate to %s", url)
		browser.OpenURL(url)
	} else {
		log.Printf("server started on %s", url)
	}

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

			os.Exit(0)
		}
	}

	if subcommand != nil && subcommand.Err() != nil {
		fmt.Fprintf(os.Stderr, subcommand.Err().Error())
		os.Exit(1)
	}

	os.Exit(0)
}
