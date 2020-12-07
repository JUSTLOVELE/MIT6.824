package mr

//lab01代码
//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
	"time"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
//type RegisterArgs struct {
//
//}
//
//type RegisterReply struct {
//	WorkerId int
//}
//
//type TaskArgs struct {
//	WorkerId int
//}
//
//type TaskReply struct {
//	Task *Task
//}
//
//type ReportTaskArgs struct {
//	Done bool
//	Seq int
//	Phase TaskPhase
//	WorkerId int
//}
//
//type ReportTaskReply struct {
//}

type Task struct {
	Id     int       //任务id,从1开始;如果从0开始rpc传输过程会变为-1
	Status int       //任务状态
	FileName string
	Time   time.Time //启动时间
	NReduce int
}
//map任务数目
type MapTaskNumber struct {
	Number int
}
//reduce任务数目
type ReduceTaskNumber struct {
	Number int
}

type TaskReply struct {
	allFinish bool
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

const(

	TASK_STATUS_WAIT_MAP = 0//初始化完成,等待map任务请求
	TASK_STATUS_DO_MAP = 1 //执行map任务
	TASK_STATUS_MAP_FINISH_WAIT_REDUCE = 2
	TASK_STATUS_DO_REDUCE = 3
	TASK_STATUS_REDUCE_FINSH = 4
)


