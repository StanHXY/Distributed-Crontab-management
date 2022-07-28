package worker

import (
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"test/src/github.com/Stan-HXY/distributed_crontab/common"
	"test/src/github.com/coreos/etcd/clientv3"
	"time"
)

// Job management tool
type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	G_jobMgr *JobMgr
)

// watch to update on job
func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)
	//1. get all tasks under /cron/jobs/ directory, and acknowledge cluster revision
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	// Get current jobs
	for _, kvPair = range getResp.Kvs {
		// Deserialize json, get job info
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			// TODO: sync job to scheduler(schedule goroutine)
			G_scheduler.PushJobEvent(jobEvent)
		}
	}

	//2. start watch from this revision
	go func() { // watch goroutine
		watchStartRevision = getResp.Header.Revision + 1
		// start watching /cron/jobs/ changes
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// process watch event
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // Save event
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// build update event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)

				case mvccpb.DELETE: // Delete event

					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					// build delete event
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				// push the change to scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}

	}()

	return
}

// terminate watcher notice
func (jobMgr *JobMgr) watchKiller() {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *common.JobEvent
		jobName    string
		job        *common.Job
	)

	// watch /cron/killer/ directory
	go func() { // watch goroutine
		// start watching /cron/killer/ changes
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		// process watch event
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // kill job event
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					// push the event to scheduler
					G_scheduler.PushJobEvent(jobEvent)

				case mvccpb.DELETE: // killer tag out-of-date, auto delete it

				}

			}
		}

	}()
}

func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// assign to instance
	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	// launch watcher
	G_jobMgr.watchJobs()

	// launch terminator
	G_jobMgr.watchKiller()

	return
}

// Create job execution lock
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return
}
