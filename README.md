# Sentinel Orchestra

## Prerequisites
[Ensure that NSQ is setup correctly.](https://github.com/gakiwate/sentinel-deployment)

## Usage
Run Sentinel Orchestra

	./sentinel-orchestra

Run ZDNS in sentinel mode

	./zdns sentinel --iterative --nsq-mode --result-verbosity trace --timeout 10

