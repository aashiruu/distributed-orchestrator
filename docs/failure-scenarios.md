# Fault Detection & Production Failure Analysis

## 1. Worker Crash Mid-Processing
* **Condition Check:** A worker picks up a job payload, updates the database status indicator row to `RUNNING`, and crashes or suffers an immediate hardware failure.
* **Infrastructure Recovery:** Because `auto-ack` is explicitly disabled, RabbitMQ detects the drop in the worker's TCP connection and safely moves the unacknowledged message back to the active queue.
* **Fencing Resolution:** A surviving worker nodes pulls the message, queries the database, and sees it marked as `RUNNING`. It checks the `updated_at` time window. If the lease age exceeds 30 seconds, it triggers a `CRASH RECOVERY` hijack, locks out the dead worker, and safe-executes the job.

## 2. Shared Broker / Network Interventions
* **Condition Check:** What happens if the PostgreSQL state register goes down or hits connection pool limits?
* **Infrastructure Recovery:** The Ingestion API code uses strict transactional ordering: **DB write must succeed before AMQP publish can trigger**. If the DB engine drops or fails, requests are safely rejected with an HTTP `500 Internal Server Error`, preventing orphaned messages from floating around the queue backbone without an audit trail.
