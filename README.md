# Celery Operator

**Project Status**: It is now under development. The tetative first release will be on April, 2020.

Celery Operator deploys a quick celery solution within Kubernetes. You can manage multiple celery clusters
within the Kubernetes through this operator. This operator is powered by operator framework.

The tetative features supported are listed here.
* Dependency Free - No external broker is needed. It will spin up a broker for you automatically
* Mutiple Worker Pools - You can configure different worker pools for different queue
* Built-in HPA supported - A simple autoscaler based on resource usage

## Progress updated
Here is the tracker for the **MVP** of this operators. After MVP, alpha version will be implemented.

- [ ] Deployment Spawning
    - [X] Redis Deployment Broker
    - [X] Scheduler Deployment
    - [ ] Worker
- [ ] E2E Test
    - [ ] Basic Celery Object Creation
    - [ ] Task Queuing

**Please note that this has not been production ready yet.**
