package worker

import (
	"fmt"
	"test/src/github.com/Stan-HXY/distributed_crontab/common"
	"time"
)

// Scheduler Goroutine
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              // etcd events queue
	jobPlanTable      map[string]*common.JobSchedulePlan // schedule plan table
	jobExecutingTable map[string]*common.JobExecuteInfo  // job execution table
	jobResultChan     chan *common.JobExecuteResult      // execution result queue
}

var (
	G_scheduler *Scheduler
)

// process the event
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		err             error
		jobExisted      bool
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
	)

	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan

	case common.JOB_EVENT_DELETE:
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL:
		// cancel execution of command
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // trigger terminate sub process
		}
	}
}

// Try to execute job
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)

	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		//fmt.Println("hasn't exit, skip execution:", jobPlan.Job.Name)
		return
	}

	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)

	// save execution status
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// execute the task
	fmt.Println("Executing:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)

}

// Recalculate job scheduled status
func (scheduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}

	// current time
	now = time.Now()

	// 1. iterate all jobs
	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now) // update NextTime
		}

		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// time gap to next execution time
	schedulerAfter = (*nearTime).Sub(now)
	return

	// 2. job that is out-of-date will be executed immediately

	// 3. count time of recent task(about to out-of-date)
}

// process execution result
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)
	// delete the execution status
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

	// generate execution log
	if result.Err != common.ERR_LOCK_ALREADY_OCCUPIED {
		jobLog = &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
		// TODO: save in mongoDB
		G_logSink.Append(jobLog)
	}

	fmt.Println("Task Executed:", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}

// loop in goroutine
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *common.JobExecuteResult
	)

	scheduleAfter = scheduler.TrySchedule()

	// Delay Timer
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: // watch the update of job
			// CRUD on JobList
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // job at deadline
		case jobResult = <-scheduler.jobResultChan: // watch the execution result of job
			scheduler.handleJobResult(jobResult)
		}

		scheduleAfter = scheduler.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}

// push the changes in Event
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// Initialize scheduler
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}
	// launch goroutine
	go G_scheduler.scheduleLoop()
	return
}

// sent execution result back
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
