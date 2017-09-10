package main

import (
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
)

func main() {
	port := "9001"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	l, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Printf("error creating listener: %v", err)
		os.Exit(1)
	}
	log.Printf("listening on port %s", port)
	fcgi.Serve(l, http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		if req.Method == http.MethodGet {
			res.Write([]byte("Hello from FCGI child"))
			return
		}
		if req.Method == http.MethodPost {
			res.Write([]byte("Hello from FCGI child via POST"))
			return
		}
	}))
}
