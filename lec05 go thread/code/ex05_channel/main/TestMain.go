package main

import (
	"fmt"
	"lec05/ex05_channel"
)

func main() {

	//channelStudyGoodCase()
	//ex05_channel.UnBuffered()
	//ex05_channel.UnbufferedDeadLock()
	ex05_channel.Wait()
	//ex05_channel.ProducerConsumer()
	//ex05_channel.Buffered()
}
/***
为什么把go f1(out)往前提了一行就不报错了?
 */
func channelStudyBadCase() {

	out := make(chan int)
	out <- 2
	go f1(out)
	fmt.Println("end")
}

func channelStudyGoodCase() {

	out := make(chan int)
	go f1(out)
	out <- 2
	fmt.Println("end")
}

func f1(in chan int) {
	number := <-in
	fmt.Println(number)
}
