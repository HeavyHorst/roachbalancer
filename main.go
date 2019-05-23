package main

import (
	"flag"

	"github.com/HeavyHorst/roachbalancer/balancer"
	_ "github.com/lib/pq"
)

func main() {
	user := flag.String("user", "root", "Database user name.")
	certs := flag.String("certs-dir", "cert", "Path to the directory containing SSL certificates and keys.")
	node := flag.String("node", "", "A database node to bootstrap the loadbalancer")
	port := flag.Int("port", 26257, "The port to listen on")

	flag.Parse()

	b := balancer.New(*user, *certs, true, *node)
	b.Listen(*port)
}
