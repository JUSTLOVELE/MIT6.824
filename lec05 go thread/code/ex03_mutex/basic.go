package ex03_mutex

import "sync"
import "time"

func Basic() {
	counter := 0
	var mu sync.Mutex
	for i := 0; i < 1000; i++ {
		go func() {
			//要执行counter就要拿到锁也就是mu.Lock()
			mu.Lock()
			//这段代码defer,相当于把mu.Mulock放到counter+1的后面
			defer mu.Unlock()
			counter = counter + 1
		}()
	}

	time.Sleep(1 * time.Second)
	mu.Lock()
	println(counter)
	mu.Unlock()
}
