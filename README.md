# Sentinel Orchestra

## Prerequisites
Sentinel Orchestra is intended to run as a unit of a [systemd service](https://github.com/gakiwate/sentinel-deployment). To test it separately, ensure that NSQ (nsqadmin, nsqlookupd, nsqd) has been started on the scratch worker.

## Usage
Once NSQ is running, run the following commands:

Run Sentinel Orchestra.

	./sentinel-orchestra

The following are optional arguments:
| Flag  | Meaning | Default |
| ------------- | ------------- | ------------- | 
| `--nsq-host`  | IP address of machine running nslookupd | `localhost`
| `--nsq-topic`  | The NSQ topic to publish on | `zdns`
| `--db-dir` | Path to directory for Pebble database storage  | `/mnt/projects/zdns/sentinel/sentinelStats`

Run ZDNS in sentinel mode

	./zdns sentinel --iterative --nsq-mode --result-verbosity trace --timeout 10

