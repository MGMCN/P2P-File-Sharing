package main

import (
	"flag"
)

type config struct {
	RendezvousString string
	listenHost       string
	listenPort       int
	sharedDirectory  string
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "default", "Unique string to identify group of nodes. Share this with your friends to let them connect with you\n")
	flag.StringVar(&c.listenHost, "host", "0.0.0.0", "The bootstrap node host listen address\n")
	flag.IntVar(&c.listenPort, "port", 6666, "node listen port\n")
	flag.StringVar(&c.sharedDirectory, "src", "./", "Path to shared folder\n")
	flag.Parse()
	return c
}
