package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ricoberger/gotemplate/pkg/version"

	"github.com/julienschmidt/httprouter"
)

var (
	showVersion   = flag.Bool("version", false, "Show version information.")
	listenAddress = flag.String("listen-address", ":8080", "Address to listen on for web interface.")
)

// IndexHandler is the main handler and always returns "Hello World".
func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Hello World\n")
}

// HelloHandler handles all requests with a parameter and returns "Hello ${parameter}".
func HelloHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "Hello %s\n", ps.ByName("name"))
}

func main() {
	// Parse command-line flags.
	flag.Parse()

	// Show version information if the "-version" flag is present.
	if *showVersion {
		v, err := version.Print("gotemplate")
		if err != nil {
			log.Fatalf("Failed to print version information: %#v\n", err)
		}

		fmt.Fprintln(os.Stdout, v)
		os.Exit(0)
	}

	fmt.Printf("Starting server %s\n", version.Info())
	fmt.Printf("Build context %s\n", version.BuildContext())
	fmt.Printf("gotemplate listening on %s\n", *listenAddress)

	// Create the router and start the HTTP server.
	// The default listen address ":8080" can be overwritten with the "-listen-address" flag.
	router := httprouter.New()
	router.GET("/", IndexHandler)
	router.GET("/:name", HelloHandler)

	err := http.ListenAndServe(*listenAddress, router)
	if err != nil {
		log.Fatalf("Fatal error: %#v\n", err)
	}
}
