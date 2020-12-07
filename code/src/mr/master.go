package mr

// lab01代码
import (
	"container/list"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Your code here -- RPC handlers for the worker to call.
//
// create a Master.
// main/mrmaster.go calls this function.
//files[] string : pg-* 代表的八个文件
// nReduce is the number of reduce tasks to use(输入是10,也就是说每份文件分为10份处理).
//分出n个nReduce任务等待worker调用,每个任务都有它的id
//不管worker过来多少个请求,如果这10个任务都被掉用了,那么多余的都不会得到任务
//每个任务分为map、reduce、done三个状态
//如果任务长时间没有被完成任务就会被收回?
//如果任务失败了怎么办?
//解决:弄一个定时器,每过10秒检测一次哪些任务没有完成,如果没有完成就通知worker来完成
//
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}
	// Your code here.
	//如果文件的数量是多于nReduce的数量,那么nReduce全部都分配出去后worker要执行完这些任务再继续获取
	m.done = false
	m.mapFileTaskExList = list.New()
	m.mapFileTask = make([]Task, len(files))
	m.nReduce = nReduce
	m.mu = sync.Mutex{}
	m.isMapPhaseFinish = false
	m.finishTaskCount = 0
	//创建任务
	for i := 0; i < len(files); i++ {

		task := Task{
			Id: i+1,//从1开始
			Status: TASK_STATUS_WAIT_MAP,
			Time: time.Now(),
			FileName: files[i],
			NReduce: nReduce,
		}

		m.mapFileTaskExList.PushBack(&task)
		m.mapFileTask[i] = task
	}

	m.server()
	return &m
}

type Master struct {
	// Your definitions here.
	nReduce int
	mu sync.Mutex
	done bool
	mapFileTaskExList *list.List //map任务执行列表
	mapFileTask []Task
	isMapPhaseFinish bool //map阶段是否结束
	finishTaskCount int
}
//RPC通信:获取map任务
func(m *Master) GetTask(args *ExampleArgs, task *Task) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.done{
		//任务已完成直接退出
		fmt.Println("m.done=true")
		return nil
	}

	if m.mapFileTaskExList.Len() == 0 {
		//没有任务要检测是否全部完成
		return nil
	}

	e := m.mapFileTaskExList.Front()
	t := (e.Value).(*Task)

	if t.Status == TASK_STATUS_WAIT_MAP {
		t.Status = TASK_STATUS_DO_MAP
	}

	if t.Status == TASK_STATUS_MAP_FINISH_WAIT_REDUCE {
		t.Status = TASK_STATUS_DO_REDUCE
	}

	task.Id = t.Id
	task.Status = t.Status
	task.FileName = t.FileName
	task.Time = t.Time
	task.NReduce = t.NReduce
	m.mapFileTaskExList.Remove(e)
	//if m.mapIndex == m.taskNumber && m.reduceIndex == m.reduceIndex {
	//	//分配任务结束
	//	return nil
	//}

	return nil
}
//完成map任务
func(m *Master) FinishMapTask(task *Task, reply *ExampleReply) error {

	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.mapFileTask[task.Id-1]
	t.Status = task.Status
	fmt.Printf("map任务-%d完成\n", t.Id)
	m.mapFileTaskExList.PushBack(&t)
	//完成map任务后,要么这里直接通知worker可以继续reduce任务了,要么往reduce任务里面加数据,然后等着rpc来调用
	return nil
}
//完成reduce任务
func(m *Master) FinishReduceTask(task *Task, reply *TaskReply) error {

	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.mapFileTask[task.Id-1]
	t.Status = task.Status
	fmt.Printf("reduce任务-%d完成\n", t.Id)
	m.finishTaskCount++

	if m.finishTaskCount == len(m.mapFileTask) {
		reply.allFinish = true
		m.done = true
	}

	return nil
}

func (m *Master) TestSuccess(args *ExampleArgs, reply *ExampleReply) error {
	m.done = true
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//先利用rpc注册了master实例的服务,然后监听服务,开启http.Server进程
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	//ret := false
	m.mu.Lock()
	defer m.mu.Unlock()
	// Your code here.
	return m.done
}
