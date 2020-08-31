# Celery Operator

[![github-action-build](https://github.com/RyanSiu1995/celery-operator/workflows/Build/badge.svg)](https://github.com/RyanSiu1995/celery-operator/actions)
[![codecov](https://codecov.io/gh/RyanSiu1995/celery-operator/branch/master/graph/badge.svg)](https://codecov.io/gh/RyanSiu1995/celery-operator)

**Project Status**: The basic CRD has been created. The metrics
extraction and testing is ongoing right now...

Celery Operator deploys a quick celery solution within
Kubernetes. You can manage multiple celery clusters
within the Kubernetes through this operator. This
operator is powered by operator framework. So, this is a more
native way to create the celery infrastructure.

The tetative features supported are listed here.

* Dependency Free - No external broker is needed.
  It will spin up a broker for you automatically
* Mutiple Worker Pools - You can configure different
  worker pools for different queue
* Built-in HPA supported - A simple autoscaler based on resource usage

## Progress updated

Here is the tracker for the **MVP** of this operators. After MVP,
alpha version will be implemented.

* [X] Pod Spawning
  * [X] Redis Deployment Broker
  * [X] Scheduler Deployment
  * [X] Worker
* [X] Pod Scaling
  * [X] Worker
  * [X] Scheduler
* [ ] Field update on the CRD
  * [X] Worker
  * [ ] Scheduler
  * [X] Broker
  * [ ] Celery Stack
* [ ] HPA
  * [ ] Implement scaling in broker
  * [ ] Implement scaling in worker
  * [ ] Implement scaling in scheduler
* [ ] Metric Export
  * [ ] Get the metric from Broker
  * [ ] Display the metric in Get and Describe API
  * [ ] Create Metric Endpoint
* [ ] Testing
  * [X] Basic Celery Object Creation
  * [X] Pod delete and respawning testing
  * [ ] Task Queuing

**Please note that this has not been production ready yet.**
