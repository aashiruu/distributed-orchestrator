# Core Design Trade-offs and Engineering Decisions

## 1. At-Least-Once Delivery vs. Performance Overhead
We chose **At-Least-Once Delivery** over maximum high-throughput processing. Enforcing persistent message flags in RabbitMQ, enabling Append-Only Files (AOF) in Redis, and executing explicit PostgreSQL status index updates on every lifecycle transition introduces I/O disk latency. 
* *The Trade-off Decision:* This trade-off is necessary for production operations. Protecting against duplicate payments, data loss, or missed customer analytics tracking runs is significantly more valuable than saving a few milliseconds of runtime speed.

## 2. Distributed Fencing vs. Dedicated Orchestrator Reapers
Instead of building a separate "reaper" cron-service that checks for stuck jobs every few minutes, we built the fencing logic directly into the worker nodes' message consumption loops.
* *The Trade-off Decision:* This distributed approach ensures near-instant recovery. A stale job is checked for expiration the exact moment a worker pulls it from the queue, completely eliminating the lag time introduced by background polling schedules.
