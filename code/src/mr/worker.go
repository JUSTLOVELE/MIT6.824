package mr

//lab01代码
import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"sync"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func (w *worker) Done() bool {
	//ret := false
	w.mu.Lock()
	defer w.mu.Unlock()
	// Your code here.
	return w.done
}


//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	w := worker{}
	w.mapf = mapf
	w.done = false
	w.reducef = reducef
	w.RunTask()
	//w.register()
	//w.run()
	// uncomment to send the Example RPC to the master.
	//CallExample()
}

type worker struct {
	id      int
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string
	done bool
	mu sync.Mutex

}

//执行map任务
func (w *worker) doMapTask(task *Task) {

	contents, err := ioutil.ReadFile(task.FileName)

	if err != nil {
		fmt.Println("map任务读取文件失败")
		return
	}
	//kva是这样的键值对{Project 1} {Gutenberg 1} {tm 1}
	kva := w.mapf(task.FileName, string(contents))
	reduces := make([][]KeyValue, task.NReduce)

	for _, kv := range kva {

		idx := ihash(kv.Key) % task.NReduce
		reduces[idx] = append(reduces[idx], kv)
	}

	for idx, keyValues := range reduces {
		//中间文件名
		intermediateFileName := fmt.Sprintf("mr-%d-%d", task.Id, idx)
		file, _ := os.Create(intermediateFileName)
		sort.Sort(ByKey(keyValues))
		enc := json.NewEncoder(file)

		for _, kv := range keyValues {
			enc.Encode(&kv)
			//fmt.Fprintf(file, "%v %v\n", kv.Key, kv.Value)
		}

		file.Close()
	}
	//到这里第一个任务就完成了,可以告诉master该任务已完成
	task.Status = TASK_STATUS_MAP_FINISH_WAIT_REDUCE
	reply := ExampleReply{}
	call("Master.FinishMapTask", &task, &reply)
}

//reduce任务
func (w *worker) doReduceTask(task *Task) {
	//把map输出的中间值合并起来得到的task.Id=1,那么就要合并mr-1-0到mr-1-8;9个文件的数据
	intermediate := []KeyValue{}

	for i := 0; i < task.NReduce; i++ {

		var fileName string

		if task.Id == 1 {
			fileName = "mr-1-0"
		}else{
			fileName = fmt.Sprintf("mr-%d-%d", task.Id-1, i)
		}

		file, err := os.Open(fileName)

		if err != nil {
			fmt.Println("doReduceTask打开文件失败")
			return
		}
		//官网代码
		dec := json.NewDecoder(file)
		kva := []KeyValue{}
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				fmt.Println(err)
				break
				//return
			}

			kva = append(kva, kv)
		}

		intermediate = append(intermediate, kva...)
		file.Close()
	}

	oname := fmt.Sprintf("mr-out-%d", task.Id-1)
	ofile, _ := os.Create(oname)

	//
	// call Reduce on each distinct key in intermediate[],
	// and print the result to mr-out-0.
	//
	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}

		output := w.reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	ofile.Close()
	task.Status = TASK_STATUS_REDUCE_FINSH
	reply := TaskReply{allFinish: false}
	call("Master.FinishReduceTask", &task, &reply)

	if reply.allFinish {
		w.done = true
	}
}

//是否需要根据nReduce数量控制进程?
func (w *worker) RunTask() {
	//请求任务
	for w.Done() == false {

		task := Task{
			Id: -1,
		}
		args := ExampleArgs{}
		//第一个参数args是传递过去的是不会变的,在master修改这个值没有用
		//第二个参数也就是task是给master修改的
		call("Master.GetTask", &args, &task)

		if task.Id == -1 {
			//fmt.Println("获取不到可执行的任务")
			continue
		}

		if task.Status == TASK_STATUS_DO_MAP {
			w.doMapTask(&task)
		}

		if task.Status == TASK_STATUS_DO_REDUCE {
			w.doReduceTask(&task)
		}
	}
}

func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)
	call("Master.TestSuccess", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}