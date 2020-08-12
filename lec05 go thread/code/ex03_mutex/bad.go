package ex03_mutex

import "time"

func BadCase() {
	counter := 0
	for i := 0; i < 1000; i++ {
		go func() {
			counter = counter + 1
		}()
	}

	time.Sleep(1 * time.Second)
	println(counter)
}
