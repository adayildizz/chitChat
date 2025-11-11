package main

func main() {

	c := &client{
		state:      "CLOSED",
		isn:        1,
		toServer:   nil,
		fromServer: nil,
	}
	s := &server{
		state:      "LISTEN",
		isn:        10,
		toClient:   nil,
		fromClient: nil,
	}

	// task 1
	handShake(c, s)

	// task 2
	//handShake2(c, s)

	// task 3

}
