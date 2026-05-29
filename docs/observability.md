# Infrastructure Observability & Telemetry Matrix

This document profiles the scrape layouts, data collection paths, and monitoring queries used to evaluate the platform's performance under load.

## Prometheus Core Scrape Topologies

Prometheus operates natively inside the container isolation zone, leveraging internal Docker DNS records to target system endpoints without exposing metrics to the public internet:

| Target Component | Internal DNS Hook | Endpoint | Scrape Rate |
| :--- | :--- | :--- | :--- |
| Ingestion API Gateway | `http://orchestrator:8080` | `/metrics` | `2s` |
| Worker Processing Pool | `http://worker:8081` | `/metrics` | `2s` |

##
Live Production Metric Catalog

### 1. Ingestion Gateway Performance
* **Metric Identifier:** `orchestrator_jobs_ingested_total`
* **Metric Type:** Counter
* **Label Dimension Rules:** `job_name` (e.g., `video.transcode`, `fail.me`)
* **Production Query Expression:** Calculates moving volume ingestion spikes per second:
    ```promql
    rate(orchestrator_jobs_ingested_total[1m])
    ```

### 2. Infrastructure Latency Distribution
* **Metric Identifier:** `prometheus_tsdb_head_samples_appended_total`
* **Metric Type:** Counter
* **Purpose:** Diagnostic engine verifying storage engine state synchronization rates across the scraping engine.

## Dashboard Performance Profile

During traffic stress testing loops, your live panels will verify system performance via two distinct indicators:
1.  **Workload Processing Peaks:** Successful processing tracks a steady execution wave matching your input velocities.
2.  **Chaos Failure Rates:** Simulated exception vectors register immediately without cross-contaminating stable processing memory pools.
