package mr

//lab01代码
//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const ISTEST = false

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
	FileIndex     int       //任务id,从1开始;如果从0开始rpc传输过程会变为-1
	Status int       //任务状态
	Phase int
	FileName string
	Time   time.Time //启动时间
	NReduce int
	Alive bool
	NMaps int
}

type FileTaskStat struct {
	Id int
	Status int
	StartTime time.Time
}

type TaskReply struct {
	Task *Task
}

type TaskArgs struct {
	WorkerId int
}

type RegisterWork struct {
	WorkerId int
}

type ReportTaskArgs struct {
	Done     bool
	FileIndex      int
	Phase    int
	WorkerId int
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

func DPrintf(format string, v ...interface{}) {
	if ISTEST {
		log.Printf(format+"\n", v...)
	}
}

const(

	TASK_STATUS_DO_MAP = 1 //map阶段
	TASK_STATUS_DO_REDUCE = 2 //reduce阶段
)

func reduceName(mapIdx, reduceIdx int) string {
	return fmt.Sprintf("mr-%d-%d", mapIdx, reduceIdx)
}

func mergeName(reduceIdx int) string {
	return fmt.Sprintf("mr-out-%d", reduceIdx)
}

