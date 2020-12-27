package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type KV struct {
	Key   string
	Value string
}

func main() {

	fmt.Println("hello world!")
	file, err := os.OpenFile("/Users/yangzuliang/Documents/develop/github/MIT6.824/code/src/main/mr-1-2", os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println("doReduceTask打开文件失败")
		return
	}
	//官网代码
	//dec := json.NewDecoder(file)
	//for {
	//	var kv KV
	//	if err := dec.Decode(&kv); err != nil {
	//		fmt.Println("编码失败:", file.Name())
	//		fmt.Println(err)
	//		break
	//	}
	//}

	buff := bufio.NewReader(file)
	for {
		line, err := buff.ReadString('\n')
		if err == io.EOF {
			break
		}
		//fmt.Print(line)
		var k KV
		json.Unmarshal([]byte(line), &k)
		fmt.Println(k.Key)
		time.Sleep(10*time.Millisecond)
	}



}
