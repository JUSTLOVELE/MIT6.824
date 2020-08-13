package ex05_channel

func UnbufferedDeadLock() {
	c := make(chan bool)
	c <- true
	<-c
}
