package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"

	fcgipool "github.com/moolen/fcgi-pool"
)

func main() {
	listen := flag.String("l", "9000", "main process listens to this port")
	command := flag.String("c", "", "which command to invoke. by convention the command is invoke with one extra argument: the port number")
	startport := flag.Int("s", 9001, "starting port for the worker processes")
	worker := flag.Int("n", 10, "number of workers to spawn and maintain")
	reuse := flag.Int("r", 10, "respawn counter: after how many requests a worker respawns")
	flag.Parse()
	if *command == "" {
		log.Printf("empty command..")
		os.Exit(1)
	}
	l, err := net.Listen("tcp", "127.0.0.1:"+*listen)
	if err != nil {
		log.Printf("error creating listener: %v", err)
		os.Exit(1)
	}

	p := fcgipool.NewPool(*command, *startport, *worker, *reuse)

	fcgi.Serve(l, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		p.Dispatch(res, req)
	}))
}
