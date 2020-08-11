package ex01_goroutine

import "sync"

//这个案例中i的作用域被修改了所以显示不是按顺序
func BadCase() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			sendRPC_bad(i)
			wg.Done()
		}()
	}
	wg.Wait()
}

func sendRPC_bad(i int) {
	println(i)
}
