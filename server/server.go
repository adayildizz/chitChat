package main

import (
	pb "chitChat/grpc/chitchatpb"
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)


type clientStream struct {
	id string
	stream pb.ChitChat_JoinServer
}

type Server struct {
	pb.UnimplementedChitChatServer
	mu sync.Mutex
	clients map[string]clientStream
	logicalClock int64
}

func (s *Server) incrementClock() int64 {
	s.logicalClock++
	return s.logicalClock
}


func (s *Server) Join(req *pb.JoinRequest, stream pb.ChitChat_JoinServer) error {
	s.mu.Lock()
	s.clients[req.ClientId] = clientStream{id: req.ClientId, stream: stream}
	timestamp := s.incrementClock()
	s.mu.Unlock()

	message := &pb.ChatMessage{
		SenderId: "Server",
		Content: fmt.Sprintf("Client %s joined Chat at logical time %d", req.ClientId, timestamp),
		LogicalTime: timestamp,
	}
	s.broadcast(message)

	log.Printf("[Server] [ClientJoin] %s at logical time %d", req.ClientId, timestamp)

	<-stream.Context().Done()
	return nil
}


func (s *Server) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.Ack, error){
	s.mu.Lock()
	timestamp := s.incrementClock()
	s.mu.Unlock()

	message := &pb.ChatMessage{
		SenderId: req.SenderId,
		Content: req.Content,
		LogicalTime: timestamp,
	}

	s.broadcast(message)
	log.Printf("[Server] [MessageBroadcast] From=%s Time=%d", req.SenderId, timestamp)
    return &pb.Ack{Status: "Delivered"}, nil

}

func (s *Server) Leave(ctx context.Context, req *pb.LeaveRequest) (*pb.Ack, error){
	s.mu.Lock()
	delete(s.clients, req.ClientId)
	timestamp := s.incrementClock()
	s.mu.Unlock()

	message := &pb.ChatMessage{
		SenderId: "Server",
		Content: fmt.Sprintf("Participant %s left Chit Chat at logical time %d", req.ClientId, timestamp),
		LogicalTime: timestamp,
	}

	s.broadcast(message)
	log.Printf("[Server] [ClientLeave] %s at logical time %d", req.ClientId, timestamp)
    return &pb.Ack{Status: "Left"}, nil

}

func (s *Server) broadcast(message *pb.ChatMessage){
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, client := range s.clients {
		go func(id string, c pb.ChitChat_JoinServer){
			if err := c.Send(message); err != nil{
				log.Printf("[Server] [BroadcastError] to %s: %v", id, err)
			}
		}(id, client.stream)
	}
}

func main() {
	lis, err := net.Listen("tcp", "5051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatServer := &Server{
		clients: make(map[string]clientStream),
	}

	pb.RegisterChitChatServer(grpcServer, chatServer)
	log.Println("[Server] Started on port 50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }

}

