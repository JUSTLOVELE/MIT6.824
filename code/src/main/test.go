package main

import (
	"fmt"
	"time"
)

type TaskStat struct {
	Status    int
	WorkerId  int
	StartTime time.Time
}

func main() {

	fmt.Println("hello world!")
	vt := make([]TaskStat, 4)

	for i := 0; i < len(vt); i++ {
		fmt.Println(vt[i].WorkerId)
	}
}
