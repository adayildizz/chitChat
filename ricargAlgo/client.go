package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	ra "ricargAlgo/grpc"

	"google.golang.org/grpc"
)


func (n *Node) dialPeers() {
	for pid, addr := range n.peers {
		if pid == n.node_id {
			continue
		}
		// establish a single connection per peer
		conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
		if err != nil {
			log.Printf("[node=%s] dial to %s (%s) failed: %v (will try lazy dial later)", n.node_id, pid, addr, err)
			continue
		}
		n.clients[pid] = ra.NewRAServerClient(conn)
		log.Printf("[node=%s] connected to peer %s at %s", n.node_id, pid, addr)
	}
}

func (n *Node) ensureClient(pid string) (ra.RAServerClient, error) {
	if c, ok := n.clients[pid]; ok {
		return c, nil
	}
	// lazy dial
	addr := n.peers[pid]
	if addr == "" {
		return nil, fmt.Errorf("unknown peer %s", pid)
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	if err != nil {
		return nil, err
	}
	client := ra.NewRAServerClient(conn)
	n.clients[pid] = client
	return client, nil
}

func (n *Node) callWithRetry(call func(ctx context.Context) error, label string) error {
	var last error
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		err := call(ctx)
		cancel()
		if err == nil {
			return nil
		}
		last = err
		time.Sleep(120 * time.Millisecond)
	}
	return fmt.Errorf("%s failed after retries: %w", label, last)
}

func (n *Node) BroadcastRequest(reqTs int64) {
	for pid := range n.peers {
		if pid == n.node_id {
			continue
		}
		peer := pid
		go func() {
			client, err := n.ensureClient(peer)
			if err != nil {
				log.Printf("[node=%s] ensureClient %s error: %v", n.node_id, peer, err)
				return
			}
			msg := &ra.Request{Timestamp: reqTs, From: n.node_id}
			err = n.callWithRetry(func(ctx context.Context) error {
				_, e := client.RequestCS(ctx, msg)
				return e
			}, "RequestCS")
			if err != nil {
				log.Printf("[node=%s] RequestSent to=%s ERROR: %v", n.node_id, peer, err)
				return
			}
			n.mu.Lock()
			L := n.lamport 
			n.mu.Unlock()
			log.Printf("[L=%d][node=%s][EVENT=RequestSent][to=%s][reqTs=%d]", L, n.node_id, peer, reqTs)
		}()
	}
}

func (n *Node) SendReply(to string) error {
	client, err := n.ensureClient(to)
	if err != nil {
		return err
	}
	n.mu.Lock()
	ts := n.bump()
	L := n.lamport
	n.mu.Unlock()

	msg := &ra.Reply{Timestamp: ts, From: n.node_id}
	err = n.callWithRetry(func(ctx context.Context) error {
		_, e := client.ReplyCS(ctx, msg)
		return e
	}, "ReplyCS")
	if err != nil {
		log.Printf("[node=%s] ReplySent to=%s ERROR: %v", n.node_id, to, err)
		return err
	}
	log.Printf("[L=%d][node=%s][EVENT=ReplySent][to=%s]", L, n.node_id, to)
	return nil
}
