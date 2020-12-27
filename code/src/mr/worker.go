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
	"strings"
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
	w.register()
	w.run()
	// uncomment to send the Example RPC to the master.
	//CallExample()
}

type worker struct {
	id      int
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string
	done    bool
	mu      sync.Mutex
}

func (w *worker) register() {
	args := &ExampleArgs{}
	reply := &RegisterWork{}
	if ok := call("Master.RegWorker", args, reply); !ok {
		log.Fatal("reg fail")
	}
	w.id = reply.WorkerId
	DPrintf("注册成功")
}


func (w *worker) run() {
	// if reqTask conn fail, worker exit
	for {
		t := w.reqTask()
		if !t.Alive {
			DPrintf("任务不存活")
			return
		}
		w.doTask(t)
	}
}

func (w *worker) reqTask() Task {

	reply := TaskReply{}
	args := TaskArgs{}
	args.WorkerId = w.id
	//第一个参数args是传递过去的是不会变的,在master修改这个值没有用
	//第二个参数也就是task是给master修改的
	//call("Master.GetOneTask", &args, &task)
	if ok := call("Master.GetOneTask", &args, &reply); !ok {
		fmt.Println("worker get task fail,exit")
		os.Exit(1)
	}

	return *reply.Task
}

func (w *worker) doTask(t Task) {

	switch t.Phase {
	case TASK_STATUS_DO_MAP:
		w.doMapTask(t)
	case TASK_STATUS_DO_REDUCE:
		w.doReduceTask(t)
	default:
		panic(fmt.Sprintf("task phase err: %v", t.Phase))
	}

}

func (w *worker) reportTask(t Task, done bool, err error) {
	if err != nil {
		log.Printf("%v", err)
	}
	args := ReportTaskArgs{}
	args.Done = done
	args.FileIndex = t.FileIndex
	args.Phase = t.Phase
	args.WorkerId = w.id
	reply := ExampleReply{}
	if ok := call("Master.ReportTask", &args, &reply); !ok {
		DPrintf("report task fail:%+v", args)
	}
}

//执行map任务
func (w *worker) doMapTask(t Task) {

	contents, err := ioutil.ReadFile(t.FileName)

	if err != nil {
		w.reportTask(t, false, err)
		return
	}
	//kva是这样的键值对{Project 1} {Gutenberg 1} {tm 1}
	kva := w.mapf(t.FileName, string(contents))
	reduces := make([][]KeyValue, t.NReduce)

	for _, kv := range kva {
		idx := ihash(kv.Key) % t.NReduce
		reduces[idx] = append(reduces[idx], kv)
	}

	for idx, keyValues := range reduces {
		//中间文件名
		intermediateFileName := fmt.Sprintf("mr-%d-%d", t.FileIndex, idx)
		file, err := os.Create(intermediateFileName)

		if err != nil {
			w.reportTask(t, false, err)
			return
		}
		//sort.Sort(ByKey(keyValues))
		enc := json.NewEncoder(file)

		for _, kv := range keyValues {
			//enc.Encode(&kv)'
			if err := enc.Encode(&kv); err != nil {
				w.reportTask(t, false, err)
			}
		}

		if err := file.Close(); err != nil {
			w.reportTask(t, false, err)
		}
	}
	//到这里任务就完成了,可以告诉master该任务已完成
	w.reportTask(t, true, nil)
}

//reduce任务
func (w *worker) doReduceTask(t Task) {

	maps := make(map[string][]string)
	for idx := 0; idx < t.NMaps; idx++ {
		fileName := reduceName(idx, t.FileIndex)
		file, err := os.Open(fileName)
		if err != nil {
			w.reportTask(t, false, err)
			return
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			if _, ok := maps[kv.Key]; !ok {
				maps[kv.Key] = make([]string, 0, 100)
			}
			maps[kv.Key] = append(maps[kv.Key], kv.Value)
		}
	}

	res := make([]string, 0, 100)
	for k, v := range maps {
		res = append(res, fmt.Sprintf("%v %v\n", k, w.reducef(k, v)))
	}

	if err := ioutil.WriteFile(mergeName(t.FileIndex), []byte(strings.Join(res, "")), 0600); err != nil {
		w.reportTask(t, false, err)
	}

	w.reportTask(t, true, nil)
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
