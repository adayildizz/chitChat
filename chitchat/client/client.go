package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	pb "chitChat/grpc/chitchatpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    var printMu sync.Mutex
    conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewChitChatClient(conn)

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter your client ID: ")
    clientID, _ := reader.ReadString('\n')
    clientID = strings.TrimSpace(clientID)


    ctx := context.Background()

    stream, err := client.Join(ctx, &pb.JoinRequest{ClientId: clientID})
    if err != nil {
        log.Fatalf("Join failed: %v", err)
    }

    
    go func() {
    for {
        msg, err := stream.Recv()
        if err != nil {
            printMu.Lock()
            log.Printf("[Client %s] Stream closed: %v", clientID, err)
            printMu.Unlock()
            return
        }

        printMu.Lock()
        fmt.Printf("\r[%d] %s: %s\n> ", msg.LogicalTime, msg.SenderId, msg.Content)
        printMu.Unlock()
    }
    }()

    printMu.Lock()
    fmt.Print("> ")
    printMu.Unlock()

   
    for {

    text, _ := reader.ReadString('\n')
    text = strings.TrimSpace(text)

    if text == "exit" {
        _, _ = client.Leave(ctx, &pb.LeaveRequest{ClientId: clientID})
        printMu.Lock()
        log.Println("Left Chit Chat.")
        printMu.Unlock()
        return
    }

    _, err := client.Publish(ctx, &pb.PublishRequest{
        SenderId: clientID,
        Content:  text,
    })
    if err != nil {
        printMu.Lock()
        log.Printf("Publish failed: %v", err)
        printMu.Unlock()
    }
}

}
