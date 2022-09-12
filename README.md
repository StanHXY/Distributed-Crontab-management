# Distributed Crontab Management Studio

The project is developed by Golang with [etcd](https://etcd.io/) and MongoDB.

The entire project structure is as the figure below:

![GO_project.jpeg](Distributed%20Crontab%20Management%20Studio%206e57c4281dbd45e0874b21efcab683ef/GO_project.jpeg)

Used etcd to sychronize all tasks from the schedule to all Worker nodes. Each worker schedules full tasks independently, without directly generating rpc with the Master nodes, and each Worker node uses distributed lock preemption to solve the problem of concurrent scheduling of the same task.
