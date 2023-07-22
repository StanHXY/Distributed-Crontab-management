# Distributed Crontab Management Studio

The project is developed by Golang with [etcd](https://etcd.io/) and MongoDB.


## Overview ##
div
The entire project structure is as the figure below:

![GO_project](https://user-images.githubusercontent.com/65502269/189682207-285434f2-1239-40c7-af7b-1344c42cd41d.jpeg)

Used etcd to sychronize all tasks from the schedule to all Worker nodes. Each worker schedules full tasks independently, without directly generating rpc with the Master nodes, and each Worker node uses distributed lock preemption to solve the problem of concurrent scheduling of the same task.
