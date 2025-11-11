package main

import (
	"fmt"
	"sync"
	"time"
)

var minEating = 3

type fork struct {
	id int
	req chan request
	rel chan release

}

type phil struct {
	id int
	forkLow *fork
	forkHigh *fork
}

type request struct{
	reply chan bool
	madeBy int
}

type release struct {
	madeBy int
}

func forkProcess(f *fork, wg *sync.WaitGroup){
	defer wg.Done();
	// we can keep held by (who) in process since it is not shared
	isFree := true
	var heldBy int

	for {

		if isFree {
		// allow requests, assign them
		r := <- f.req
		heldBy = r.madeBy
		isFree = false
		r.reply <- true

	} else {
		// allow releases, assign them
		r := <- f.rel
		if r.madeBy == heldBy {
			isFree = true
		}
	}
	}
	
}

func requestFork(p *phil, f *fork){
		reply1 := make(chan bool)
		f.req <- request{reply: reply1, madeBy: p.id}
		<-reply1 // this is the core logic that blocks until reply comes 
}

func releaseFork(p *phil, f *fork){
	f.rel <- release{madeBy: p.id}

}

func philosopherProcess(p *phil, wg *sync.WaitGroup){
	defer wg.Done()

	// for 3 meals of each philosopher
	for i:=0; i<minEating; i++ {
		
		// we alwys request for the lower first
		requestFork(p, p.forkLow)
		requestFork(p, p.forkHigh)

		fmt.Printf("Philosopher %d is eating now.\n", p.id)
		time.Sleep(2 *time.Second)
		fmt.Printf("Philosopher %d finished eating.\n", p.id)

		// order does not matter 
		releaseFork(p, p.forkLow)
		releaseFork(p, p.forkHigh)
	
		fmt.Printf("Philosopher %d is thinking now.\n", p.id)
		time.Sleep(2 *time.Second)
		fmt.Printf("Philosopher %d finished thinking.\n", p.id)
	}
}


func main(){
	// number of forks and philosophers
	const N = 5
	var wg sync.WaitGroup

	// creating forks
	forks := make([]*fork, N)
	for i:=0; i<N; i++ {
		f := &fork{
			id : i,
			req: make(chan request),
			rel: make(chan release),
		}
		forks[i] = f
		
		go forkProcess(f, &wg)
	}

	//creating philosophers
	phils := make([]*phil, N)
	for i:=0; i<N-1; i++ {
		p := &phil{
			id: i,
			forkLow: forks[i],
			forkHigh: forks[i+1],
		}
		phils[i]=p
	}
	// the different solution for the last philosopher
	// this is the part that we prevent deadlock, since philosophers will always 
	// try to reach out the lowest numbered fork firstly.
	phils[N-1] = &phil{
		id: N-1,
		forkLow: forks[0],
		forkHigh: forks[N-1],
	}

	for _,p := range phils {
		wg.Add(1)
		go philosopherProcess(p, &wg)

	}

	// waiting for all processes
	wg.Wait()
	fmt.Printf("All philosophers ate %d times!", minEating)

}
