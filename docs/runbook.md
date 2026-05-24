# Operational Runbook & Telemetry Monitoring Guide

## Identifying Cluster Saturation Problems

### 1. High Ingestion Volumetrics vs. Low Processing Latency
If your API counter metrics increase dramatically while worker completion rates stall, your consumers are saturated:
* **Query KPI:** `orchestrator_jobs_ingested_total` increases exponentially while `worker_job_execution_duration_seconds_count` stays flat.
* **Mitigation Strategy:** Scale out your computing resource pool by spinning up additional isolated worker node binary processes (`cmd/worker/main.go`).

### 2. High Lock Failure Metrics
If your Redis metric logs report high collision numbers, multiple workers are competing for the exact same message deliveries:
* **Query KPI:** `worker_redis_lock_failures_total` matches or exceeds ingestion values.
* **Mitigation Strategy:** Review your queue topology configuration parameters and adjust prefetch counts (`ch.Qos`) to distribute messages more evenly across your consumer fleet.
