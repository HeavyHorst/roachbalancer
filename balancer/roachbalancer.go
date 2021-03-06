package balancer

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Balancer is a simple cockroachdb load balancer with automatic node detection.
type Balancer struct {
	nodes    []string
	n        uint
	nodeLock sync.RWMutex

	addr    string
	certdir string
	user    string
	logging bool

	ok chan struct{}
}

// New creates a new Balancer
func New(user, certdir string, logging bool, initialNodes ...string) *Balancer {
	b := &Balancer{
		nodes:   initialNodes,
		certdir: certdir,
		user:    user,
		logging: logging,
		ok:      make(chan struct{}, 1),
	}

	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for ; true; <-ticker.C {
			b.refreshLiveNodes()
			if b.logging {
				log.Println("Refreshed active node list: ", b.GetLiveNodes())
			}
		}
	}()

	return b
}

// ChooseNode returns an active database node while doing simple round robin load balancing.
func (b *Balancer) ChooseNode() string {
	b.nodeLock.Lock()
	defer b.nodeLock.Unlock()

	idx := b.n % uint(len(b.nodes))
	b.n++

	return b.nodes[idx]
}

// GetLiveNodes returns the current live node list.
func (b *Balancer) GetLiveNodes() []string {
	b.nodeLock.RLock()
	defer b.nodeLock.RUnlock()

	nodes := make([]string, 0, len(b.nodes))
	for _, v := range b.nodes {
		nodes = append(nodes, v)
	}
	return nodes
}

// GetNodeCount returns the current live node count.
func (b *Balancer) GetNodeCount() int {
	b.nodeLock.RLock()
	defer b.nodeLock.RUnlock()

	return len(b.nodes)
}

func (b *Balancer) refreshLiveNodes() {
	c := b.GetNodeCount()

	for i := 0; i < c; i++ {
		db, err := sql.Open("postgres",
			fmt.Sprintf("postgresql://%s@%s/defaultdb?connect_timeout=5&ssl=true&sslmode=require&sslrootcert=%s/ca.crt&sslkey=%s/client.%s.key&sslcert=%s/client.%s.crt", b.user, b.ChooseNode(), b.certdir, b.certdir, b.user, b.certdir, b.user))
		if err != nil {
			log.Println("[ERROR]:", err)
			continue
		}
		defer db.Close()

		rows, err := db.Query(`select address from 
		(select address, 
			CASE WHEN split_part(expiration,',',1)::decimal > now()::decimal 
			THEN true 
			ELSE false 
			END AS is_available, ifnull(is_live, false) 
			FROM crdb_internal.gossip_liveness 
			LEFT JOIN crdb_internal.gossip_nodes USING (node_id)
		) as a WHERE a.is_available = true;`)
		if err != nil {
			log.Println("[ERROR]:", err)
			continue
		}
		defer rows.Close()

		var newNodes []string
		for rows.Next() {
			var addr string
			if err := rows.Scan(&addr); err != nil {
				continue
			}
			newNodes = append(newNodes, addr)
		}

		rows.Close()
		db.Close()

		if len(newNodes) > 0 {
			b.nodeLock.Lock()
			defer b.nodeLock.Unlock()
			b.nodes = newNodes
		}

		break
	}
}

func (b *Balancer) getConnection() (net.Conn, error) {
	return net.Dial("tcp", b.ChooseNode())
}

// GetAddr returns the listener address
func (b *Balancer) GetAddr() string {
	return b.addr
}

// WaitReady blocks until Listen is ready
func (b *Balancer) WaitReady() {
	<-b.ok
}

// Listen starts the loadbalancer
func (b *Balancer) Listen(port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to bind: %s", err)
	}

	b.addr = ln.Addr().String()

	b.ok <- struct{}{}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept: %s", err)
			continue
		}

		ds, err := b.getConnection()
		if err != nil {
			log.Println("[ERROR]:", err)
			continue
		}

		go handleConnection(conn, ds)
	}
}

func copy(wc io.WriteCloser, r io.Reader) {
	defer func() {
		wc.Close()
	}()
	io.Copy(wc, r)
}

func handleConnection(us, ds net.Conn) {
	go copy(us, ds)
	go copy(ds, us)
}
