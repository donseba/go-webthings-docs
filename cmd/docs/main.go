package main

import (
	"log"
	"net/http"
	"os"

	"donseba/go-webthings-docs/internal/site"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("go-webthings docs listening on %s", addr)
	log.Printf("local docs:     http://docs.rocketweb.nl%s/go-partial", addr)
	log.Printf("local showcase: http://showcase.rocketweb.nl%s/go-partial", addr)

	if err := http.ListenAndServe(addr, site.NewRouter()); err != nil {
		log.Fatal(err)
	}
}
