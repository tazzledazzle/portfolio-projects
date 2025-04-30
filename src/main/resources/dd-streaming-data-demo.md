DESIGN DOCUMENT: Streaming-Data Demo Project
Overview:
Implement an event-processing pipeline: ingest synthetic clickstream into Kafka, process with Python or Go, store output in DynamoDB or Elasticsearch, and visualize via Grafana.

Goals and Objectives:
• Show skill in real-time data ingestion and processing
• Demonstrate storage and visualization of streaming metrics

Scope:
• Kafka producer generating synthetic events
• Consumer application normalizing and aggregating data
• Storage in DynamoDB (NoSQL) or Elasticsearch (search/analytics)
• Grafana dashboards via Prometheus metrics exporter

Architecture and Components:
• Producer: Python script publishing to Kafka topic
• Consumer: Go service consuming, processing, and writing to DB
• Storage: DynamoDB or Elasticsearch cluster
• Metrics: Prometheus exporter in consumer

Technology Stack:
• Kafka 3.x (Confluent or AWS MSK)
• Python 3.10 or Go 1.20
• DynamoDB or Elasticsearch 8.x
• Prometheus, Grafana

Data Flow and Interactions:

Producer emits events at configurable rate

Consumer reads, applies business logic (e.g., sessionization)

Aggregates written to storage

Prometheus exporter publishes processing latency, throughput

Grafana dashboards render time-series

Non-Functional Requirements:
• Process average 10k events/sec
• End-to-end latency < 200 ms

Security Considerations:
• Encrypt Kafka topics with TLS
• Use IAM roles for DynamoDB access

Deployment Strategy:
• Docker Compose for local demo
• Helm chart for k8s deployment

Testing Strategy:
• Load testing with Kafka Testkit or kafkatool
• Consumer unit tests and integration tests

Timeline and Milestones:
Week 1: Kafka setup and producer
Week 2: Consumer logic and storage
Week 3: Metrics exporter and Grafana
Week 4: Load tests and docs

Maintenance & Monitoring:
• Alert on consumer lag via Grafana
• Monthly dependency updates