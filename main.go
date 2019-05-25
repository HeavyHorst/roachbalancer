package main

import (
	"flag"
	"fmt"

	"github.com/HeavyHorst/roachbalancer/balancer"
	_ "github.com/lib/pq"
)

type nodes []string

func (n *nodes) String() string {
	return fmt.Sprint([]string(*n))
}

func (n *nodes) Set(value string) error {
	*n = append(*n, value)
	return nil
}

func main() {
	var n nodes

	user := flag.String("user", "root", "Database user name.")
	certs := flag.String("certs-dir", "cert", "Path to the directory containing SSL certificates and keys.")
	flag.Var(&n, "node", "A database node to bootstrap the loadbalancer")
	port := flag.Int("port", 26257, "The port to listen on")

	flag.Parse()

	b := balancer.New(*user, *certs, true, n...)
	b.Listen(*port)
}
