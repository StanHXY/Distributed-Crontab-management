package worker

import (
	"math/rand"
	"os/exec"
	"test/src/github.com/Stan-HXY/distributed_crontab/common"
	"time"
)

type Executor struct {
}

var (
	G_executor *Executor
)

func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
		)
		// execution result
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}

		//------ initialize, obtain distributed lock ------
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)

		// start time
		result.StartTime = time.Now()

		// random sleep
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		// try to compete lock ---------
		err = jobLock.TryLock()
		// release lock ---------
		defer jobLock.Unlock()

		// failed to lock
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else { // succeed to lock
			// start time
			result.StartTime = time.Now()

			// exec Shell command
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)

			output, err = cmd.CombinedOutput()

			// end time
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		// after execution, return the result to scheduler,
		//and scheduler will delete record from execution table
		G_scheduler.PushJobResult(result)
	}()
}

// initialize executor
func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
