package main

import (
	"embed"
	"flag"
	"log"
	"net"
	"net/http"
)

//go:embed index.html landing.html
var embedded embed.FS

var indexHTML []byte
var landingHTML []byte

func init() {
	var err error
	indexHTML, err = embedded.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	landingHTML, err = embedded.ReadFile("landing.html")
	if err != nil {
		panic(err)
	}
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	store, err := LoadStore("data/markets.json", "data/agents.json")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Yarrow demo listening on %s", displayURL(*addr))
	if err := http.ListenAndServe(*addr, NewRouter(store)); err != nil {
		log.Fatal(err)
	}
}

func displayURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr
	}
	if host == "" || host == "::" || host == "0.0.0.0" || host == "[::]" {
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port)
}
