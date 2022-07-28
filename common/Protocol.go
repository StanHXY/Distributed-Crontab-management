package common

import (
	"context"
	"encoding/json"
	"strings"
	"test/src/github.com/gorhill/cronexpr"
	"time"
)

// Timed task
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

// job execute status
type JobExecuteInfo struct {
	Job        *Job
	PlanTime   time.Time
	RealTime   time.Time
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

// schedule plan
type JobSchedulePlan struct {
	Job      *Job // job that needs to be scheduled
	Expr     *cronexpr.Expression
	NextTime time.Time // next time to be executed
}

// Http response
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

// Event change: update(save), delete
type JobEvent struct {
	EventType int // SAVE, DELETE
	Job       *Job
}

// execution result
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output      []byte
	Err         error
	StartTime   time.Time
	EndTime     time.Time
}

// task execution log
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`
	Command      string `json:"command" bson:"command"`
	Err          string `json:"err" bson:"err"`
	Output       string `json:"output" bson:"output"`
	PlanTime     int64  `json:"planTime" bson:"planTime"`
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"`
	StartTime    int64  `json:"startTime" bson:"startTime"`
	EndTime      int64  `json:"endTime" bson:"endTime"`
}

// log batch
type LogBatch struct {
	Logs []interface{}
}

// log filter
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// log sort rules
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"`
}

// Response method
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)

	response.Errno = errno
	response.Msg = msg
	response.Data = data

	resp, err = json.Marshal(response)

	return
}

// Deserialize job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// extract job name from etcd Key
func ExtractJobName(jobKey string) (jobName string) {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

// extract job name from etcd killer Key
func ExtractKillerName(killerKey string) (jobName string) {
	return strings.TrimPrefix(killerKey, JOB_KILLER_DIR)
}

// Event change: update(save), delete
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// build job execute plan
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	// Generate schedule plan
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

// build job execution information
func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

// extract worker ip
func ExtractWorkerIP(regKey string) (ip string) {
	return strings.TrimPrefix(regKey, JOB_WORKER_DIR)
}
