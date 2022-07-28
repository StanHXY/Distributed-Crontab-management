package worker

import (
	"context"
	"test/src/github.com/Stan-HXY/distributed_crontab/common"
	"test/src/github.com/coreos/etcd/clientv3"
)

// Distributed transaction
type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	cancelFunc context.CancelFunc // for terminate auto-renew
	leaseID    clientv3.LeaseID
	isLocked   bool
}

// initialize a lock object
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

// Try to lock
func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseID        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)

	// 1. create lease 5s
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	leaseID = leaseGrantResp.ID

	// 2. auto renew lease
	if keepRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseID); err != nil {
		goto FAIL
	}

	// goroutine for process lease renewal
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)

		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	// 3. create txn
	txn = jobLock.kv.Txn(context.TODO())
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName

	// 4. txn compete for lock
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockKey))

	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	// 5. return for success, rollback for fail, release lease
	//-- failed, lock is occupied
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_OCCUPIED
		goto FAIL
	}

	//-- succeeded, commit transaction
	jobLock.leaseID = leaseID
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return

FAIL:
	cancelFunc()                                  // cancel auto-renew
	jobLock.lease.Revoke(context.TODO(), leaseID) // release lease
	return
}

// release lock(close lease)
func (jobLock *JobLock) Unlock() {
	if jobLock.isLocked {
		jobLock.cancelFunc()                                  // cancel goroutine
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseID) // the key will be deleted
	}
}
