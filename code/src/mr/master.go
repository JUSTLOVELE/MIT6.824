package mr

// lab01代码
import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Master struct {
	// Your definitions here.
	nReduce int
	files []string
	mu sync.Mutex
	done bool
	taskCh chan Task //任务,有多少个服务器
	fileTaskStats []FileTaskStat //文件队列
	phase int //阶段
	workerId int
}
// Your code here -- RPC handlers for the worker to call.
//
// create a Master.
// main/mrmaster.go calls this function.
//files[] string : pg-* 代表的八个文件
// nReduce is the number of reduce tasks to use(输入是10,也就是说每份文件分为10份处理).
//有nreduce个map任务也有nReduce个reduce任务
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
	m.files = files
	m.nReduce = nReduce
	m.mu = sync.Mutex{}
	m.phase = TASK_STATUS_DO_MAP
	m.fileTaskStats = make([]FileTaskStat, len(files))

	if nReduce > len(files) {
		m.taskCh = make(chan Task, nReduce)
	}else{
		m.taskCh = make(chan Task, len(files))
	}

	go m.tickSchedule()
	m.server()
	return &m
}

func (m *Master) tickSchedule() {
	// 按说应该是每个 task 一个 timer，此处简单处理
	for !m.Done() {
		go m.schedule()
		time.Sleep(time.Millisecond * 500)
	}
}

const (
	File_Task_Status_Ready   = 0
	File_Task_Status_Queue   = 1
	File_Task_Status_Running = 2
	File_Task_Status_Finish  = 3
	File_Task_Status_Err     = 4
)

//文件切换是要在master这里完成的,map任务只负责获取中间文件
func (m *Master) schedule() {

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.done {
		return
	}
	allFinish := true
	for index, t := range m.fileTaskStats {
		switch t.Status {
		case File_Task_Status_Ready:
			allFinish = false
			m.taskCh <- m.getTask(index)
			m.fileTaskStats[index].Status = File_Task_Status_Queue
		case File_Task_Status_Queue:
			allFinish = false
		case File_Task_Status_Running:
			allFinish = false
			//超时丢回队列重新执行
			if time.Now().Sub(t.StartTime) > time.Second * 5 {
				m.fileTaskStats[index].Status = File_Task_Status_Queue
				m.taskCh <- m.getTask(index)
			}
		case File_Task_Status_Finish:
		case File_Task_Status_Err:
			//报错放回队列重新执行
			allFinish = false
			m.fileTaskStats[index].Status = File_Task_Status_Queue
			m.taskCh <- m.getTask(index)
		default:
			panic("t.status err")
		}
	}

	if allFinish {
		if m.phase == TASK_STATUS_DO_MAP {
			m.initReduceTask()
		} else {
			m.done = true
		}
	}
}

func (m *Master) initReduceTask() {

	m.phase = TASK_STATUS_DO_REDUCE
	m.fileTaskStats = make([]FileTaskStat, m.nReduce)
}

func (m *Master) getTask(index int) Task {
	task := Task{
		FileName: "",
		NReduce:  m.nReduce,
		NMaps:    len(m.files),
		FileIndex:      index,
		Phase: m.phase,
		Alive:    true,
	}

	if task.Phase == TASK_STATUS_DO_MAP {
		task.FileName = m.files[index]
	}

	return task
}
//1、work先向master注意个work
func (m *Master) RegWorker(args *ExampleArgs, reply *RegisterWork) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workerId += 1
	reply.WorkerId = m.workerId
	return nil
}
//2、获得一个任务
func (m *Master) GetOneTask(args *TaskArgs, taskReply *TaskReply) error {

	t := <-m.taskCh
	taskReply.Task = &t

	if t.Alive {
		m.regTask(args, &t)
	}

	return nil
}

func (m *Master) regTask(args *TaskArgs, task *Task) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task.Phase != m.phase {
		panic("req Task phase neq")
	}

	m.fileTaskStats[task.FileIndex].Status = File_Task_Status_Running
	m.fileTaskStats[task.FileIndex].Id = args.WorkerId
	m.fileTaskStats[task.FileIndex].StartTime = time.Now()
}

//3、收到worker的报告
func (m *Master) ReportTask(args *ReportTaskArgs, reply *ExampleReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	DPrintf("得到ReportTask: %+v, 任务阶段: %+v", args, m.phase)

	if m.phase != args.Phase || args.WorkerId != m.fileTaskStats[args.FileIndex].Id {
		return nil
	}

	if args.Done {
		m.fileTaskStats[args.FileIndex].Status = File_Task_Status_Finish
	} else {
		m.fileTaskStats[args.FileIndex].Status = File_Task_Status_Err
	}

	go m.schedule()
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
