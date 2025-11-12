package main

import (
	"context"
	ra "ricargAlgo/grpc"

	"log"
)

type RAServer struct {
	ra.UnimplementedRAServerServer
	node *Node
}

func (s *RAServer) RequestCS(ctx context.Context, req *ra.Request) (*ra.Empty, error) {

	n := s.node
	n.mu.Lock()

	n.onReceive(req.Timestamp)

	replyNow := false
	if    !n.isRequesting {
		replyNow = true
	} else if s.node.tupleLess(req.Timestamp, req.From, n.requestTS, n.node_id){
		replyNow = true
	} else {
		n.deferred[req.From] = true
	}

	// Capture values for logging while holding the lock
	L := n.lamport
	reqTs := n.requestTS
	requesting := n.isRequesting
	n.mu.Unlock()

	if replyNow {

		_ = n.SendReply(req.From)
		log.Printf("[L=%d][node=%s][EVENT=RequestRecv][from=%s][ts=%d][decision=REPLIED][reqTs=%d][requesting=%v]",
			L, n.node_id, req.From, req.Timestamp, reqTs, requesting)
	} else {
		log.Printf("[L=%d][node=%s][EVENT=RequestRecv][from=%s][ts=%d][decision=DEFERRED][reqTs=%d][requesting=%v]",
			L, n.node_id, req.From, req.Timestamp, reqTs, requesting)
	}

	return &ra.Empty{}, nil

}


func (s *RAServer) ReplyCS(ctx context.Context, rep *ra.Reply) (*ra.Empty, error) {

	n := s.node
	n.mu.Lock()
	n.onReceive(rep.Timestamp)
	n.remainingReplies--
	L := n.lamport
	out := n.remainingReplies
	n.cond.Broadcast()
	n.mu.Unlock()

	log.Printf("[L=%d][node=%s][EVENT=ReplyRecv][from=%s][outstanding=%d]", L, n.node_id, rep.From, out)
	return &ra.Empty{}, nil
}


func (n *Node) RequestCriticalSection(work func()) {
	// Start request
	n.mu.Lock()
	n.bump()
	n.isRequesting = true
	n.requestTS = n.lamport
	n.remainingReplies = len(n.peers) - 1 // excluding self
	reqTs := n.requestTS
	L := n.lamport
	log.Printf("[L=%d][node=%s][EVENT=RequestStart][reqTs=%d][need=%d]", L, n.node_id, reqTs, n.remainingReplies)
	n.mu.Unlock()

	// Broadcast REQUEST
	n.BroadcastRequest(reqTs)

	// Wait for all replies
	n.mu.Lock()
	for n.remainingReplies > 0 {
		n.cond.Wait()
	}
	n.mu.Unlock()

	// ENTER CS
	log.Printf("[L=%d][node=%s][EVENT=ENTER_CS]", L, n.node_id)
	work()
	log.Printf("[node=%s][EVENT=EXIT_CS]", n.node_id)

	// Exit: flush deferred
	var toFlush []string
	n.mu.Lock()
	n.isRequesting = false
	for pid := range n.deferred {
		toFlush = append(toFlush, pid)
	}
	n.deferred = make(map[string]bool)
	n.mu.Unlock()

	for _, pid := range toFlush {
		_ = n.SendReply(pid)
	}
	if len(toFlush) > 0 {
		log.Printf("[node=%s][EVENT=ReplyFlush][to=%v]", n.node_id, toFlush)
	}
}
