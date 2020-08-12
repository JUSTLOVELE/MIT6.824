package ex02_time

import "time"

func Sleep() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodicCase()
	time.Sleep(5 * time.Second) // wait for a while so we can observe what ticker does
}

func periodicCase() {
	for {
		println("tick")
		time.Sleep(1 * time.Second)
	}
}
