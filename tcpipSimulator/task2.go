package main

import (
	"fmt"
	"sync"
)


func handShake2(c *client, s *server) {
	c2s := make(chan tcpPacket)
	s2c := make(chan tcpPacket)
	c.toServer, c.fromServer = c2s, s2c
	s.toClient, s.fromClient = s2c, c2s

	var wg sync.WaitGroup
	wg.Add(2)

	go serverProcess(s, &wg)
	go clientProcess(c, &wg)

	wg.Wait()

}

func serverProcess2(s *server, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		packet := <- s.fromClient;
		fmt.Printf("SERVER ⇐ CLIENT : %-8s seq=%d ack=%d\n", packet.messageType, packet.seq, packet.ack)

		if s.state == Listen && packet.messageType == KindSYNC{
			s.state = SyncReceived 
			response := tcpPacket{
				messageType: KindSYNCACK,
				seq: s.isn,
				ack: packet.seq+1,
			}

			s.toClient <- response
			fmt.Printf("SERVER ⇒ CLIENT : %-8s seq=%d ack=%d\n", response.messageType, response.seq, response.ack)
			continue 


			
		}

		if packet.messageType == KindACK && s.state == SyncReceived {

				s.state = Established
				fmt.Println("SERVER: connection ESTABLISHED")
				return
			}

	}

}

func clientProcess2(c *client, wg *sync.WaitGroup) {
	defer wg.Done()

	sync := tcpPacket{
		messageType: KindSYNC,
		seq: c.isn,
	}

	c.state = SyncSent
	c.toServer <- sync
	fmt.Printf("CLIENT ⇒ SERVER : %-8s seq=%d ack=%d\n", sync.messageType, sync.seq, sync.ack)

	response := <- c.fromServer
	fmt.Printf("CLIENT ⇐ SERVER : %-8s seq=%d ack=%d\n", response.messageType, response.seq, response.ack)

	if response.messageType == KindSYNCACK && response.ack == c.isn+1 {
		ack := tcpPacket{
			messageType: KindACK,
			seq: c.isn+1,
			ack: response.seq+1,
		}
		fmt.Printf("CLIENT ⇒ SERVER : %-8s seq=%d ack=%d\n", ack.messageType, ack.seq, ack.ack)
		c.toServer <- ack
		c.state = Established
		fmt.Println("CLIENT: connection ESTABLISHED")
		return
	}

}
