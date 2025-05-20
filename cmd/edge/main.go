package main

import (
	"flag"
	"log"

	"github.com/example/nebula-edge/internal/agent"
)

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	flag.Parse()
	a := agent.New(*port)
	if err := a.Start(); err != nil {
		log.Fatal(err)
	}
	select {} // block forever
}
