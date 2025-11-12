package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	ra "ricargAlgo/grpc"
	"sync"
	"time"

	"google.golang.org/grpc"
)




type Node struct {
	node_id      string
	isRequesting bool
	mu sync.Mutex
	deferred     map[string]bool
	lamport int64
	requestTS int64
	remainingReplies int
	peers map[string]string
	clients map[string]ra.RAServerClient
	cond *sync.Cond
	addr string 
}


func createNode(id, addr string, peers map[string]string) *Node {
	n := &Node{
		node_id:       id,
		addr:     addr,
		peers:    peers,
		clients:  make(map[string]ra.RAServerClient),
		deferred: make(map[string]bool),
	}
	n.cond = sync.NewCond(&n.mu)
	return n
}


func (n *Node) bump() int64{
	n.lamport++
	return n.lamport
}


func ( n *Node) onReceive(timestamp int64) {
	if timestamp > n.lamport {
		n.lamport = timestamp
	}
	n.lamport++
}

func (n *Node) tupleLess(tsJ int64, j string, tsI int64, i string) bool {
	if tsJ < tsI {
		return true
	}
	if tsJ > tsI {
		return false
	}
	// tie-break 
	return j < i
}

func startServer(n *Node, listen string) {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("listen %s: %v", listen, err)
	}
	s := grpc.NewServer()
	ra.RegisterRAServerServer(s, &RAServer {node: n})
	log.Printf("[node=%s] gRPC server listening on %s", n.node_id, listen)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func loadPeers(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}


func main() {
	var id, listen, peersFile, mode string
	var slowReplyMs int
	flag.StringVar(&id, "id", "", "node id (e.g., A)")
	flag.StringVar(&listen, "listen", "", "listen addr (e.g., :50051)")
	flag.StringVar(&peersFile, "peers", "peers.json", "peers json path")
	flag.StringVar(&mode, "mode", "manual", "demo or manual")
	flag.IntVar(&slowReplyMs, "slow-reply-ms", 0, "artificial delay for replies (testing)")
	flag.Parse()

	if id == "" || listen == "" {
		log.Fatalf("usage: --id A --listen :50051 [--peers peers.json] [--mode demo|manual]")
	}

	peers, err := loadPeers(peersFile)
	if err != nil {
		log.Fatalf("load peers: %v", err)
	}

	node := createNode(id, listen, peers)
	
	go startServer(node, listen)

	
	time.Sleep(200 * time.Millisecond)

	
	node.dialPeers()

	
	switch mode {
	case "demo":
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			node.RequestCriticalSection(func() {
				// emulate CS work
				time.Sleep(500 * time.Millisecond)
				log.Printf("[node=%s] did CS work", node.node_id)
			})
		}
	default: // manual: press Enter to request
		log.Printf("[node=%s] manual mode: press ENTER to request CS", node.node_id)
		for {
			_, _ = fmt.Scanln()
			node.RequestCriticalSection(func() {
				time.Sleep(500 * time.Millisecond)
				log.Printf("[node=%s] did CS work", node.node_id)
			})
		}
	}
}