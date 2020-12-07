package main

import (
	"container/list"
	"fmt"
)

func main() {
	//闭包
	//ex01_goroutine.Closure()
	//ex01_goroutine.Loop()
	//ex01_goroutine.BadCase()

	s1 := Student{
		Name: "ysa",
		Age:  12,
	}

	s2 := Student{
		Name: "dsd",
		Age:  1,
	}

	l := list.New()
	l.PushBack(&s1)
	l.PushBack(&s2)

	s := (l.Front().Value).(*Student)
	fmt.Println(s.Name)

	x := 10
	var p *int = &x //获取x的地址，然后保存到指针类型的变量p中
	*p += 20        //通过指针类型来操作变量x
	fmt.Println(p, *p)
}

type Student struct {
	Name string
	Age  int
}
