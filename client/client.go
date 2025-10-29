package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	pb "chitChat/grpc/chitchatpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewChitChatClient(conn)

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter your client ID: ")
    clientID, _ := reader.ReadString('\n')
    clientID = clientID[:len(clientID)-1]

    ctx := context.Background()

    stream, err := client.Join(ctx, &pb.JoinRequest{ClientId: clientID})
    if err != nil {
        log.Fatalf("Join failed: %v", err)
    }

    
    go func() {
        for {
            msg, err := stream.Recv()
            if err != nil {
                log.Printf("[Client %s] Stream closed: %v", clientID, err)
                return
            }
            fmt.Printf("[%d] %s: %s\n", msg.LogicalTime, msg.SenderId, msg.Content)
        }
    }()

   
    for {
        fmt.Print("> ")
        text, _ := reader.ReadString('\n')
        text = text[:len(text)-1]

        if text == "exit" {
            _, _ = client.Leave(ctx, &pb.LeaveRequest{ClientId: clientID})
            log.Println("Left Chit Chat.")
            return
        }

        _, err := client.Publish(ctx, &pb.PublishRequest{
            SenderId: clientID,
            Content:  text,
        })
        if err != nil {
            log.Printf("Publish failed: %v", err)
        }
    }
}
