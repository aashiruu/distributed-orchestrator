# Operational Runbook & Telemetry Monitoring Guide

## Identifying Cluster Saturation Problems

### 1. High Ingestion Volumetrics vs. Low Processing Latency
If your API counter metrics increase dramatically while worker completion rates stall, your consumers are saturated:
* **Query KPI:** `orchestrator_jobs_ingested_total` increases exponentially while `rate(orchestrator_jobs_ingested_total{job_name="video.transcode"}[1m])` stays flat.
* **Mitigation Strategy:** Scale up processing capabilities smoothly inside your network mesh (`docker compose up -d --scale worker=3`).

### 2. High Lock Failure Metrics
If your Redis metric logs report high collision numbers, multiple workers are competing for the exact same message deliveries:
* **Query KPI:** `worker_redis_lock_failures_total` matches or exceeds ingestion values.
* **Mitigation Strategy:** Review your queue topology configuration parameters and adjust prefetch counts (`ch.Qos`) to distribute messages more evenly across your consumer fleet.
