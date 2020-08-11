package ex01_goroutine

import "sync"

//闭包
func Closure() {
	var a string
	var wg sync.WaitGroup
	wg.Add(1)
	v := 1
	//这个函数是闭包函数,闭包函数可以引用所在方法体的变量
	//跑一个新的线程
	go func() {
		a = "hello world"
		println(v)
		wg.Done()
	}()
	wg.Wait()
	println(a)
}
