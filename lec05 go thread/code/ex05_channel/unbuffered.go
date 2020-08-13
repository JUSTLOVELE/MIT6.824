package ex05_channel

import "time"
import "fmt"

func UnBuffered() {
	c := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		<-c
	}()
	start := time.Now()
	c <- true // blocks until other goroutine receives
	fmt.Printf("send took %v\n", time.Since(start))
}
