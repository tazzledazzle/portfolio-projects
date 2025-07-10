# DESIGN DOCUMENT: Cloud-Native Microservices Project

----

## Overview
Build a suite of cloud-native microservices in Kotlin (Spring Boot) or Go, backed by PostgreSQL, with event streaming via Kafka, deployed using Terraform to AWS. Provide observability via Prometheus and Grafana.

## Goals and Objectives
* Demonstrate end-to-end microservices lifecycle (dev → deploy → monitor)
* Showcase event-driven architecture
* Illustrate infrastructure-as-code best practices

## Scope
* Three microservices (User, Order, Notification)
* Terraform AWS modules for VPC, RDS, MSK (Kafka), ECS or EKS
* Prometheus exporters on each service, Grafana dashboards

----

## Architecture and Components
* API Gateway (ALB) routes to services on ECS Fargate
* Kafka cluster for events between services
* RDS PostgreSQL for stateful data
* Terraform modules in infra/ directory

## Technology Stack
* Kotlin + Spring Boot or Go + Gin
* AWS: VPC, ECS/EKS, RDS, MSK, IAM
* Terraform 1.x, AWS CLI
* Prometheus, Grafana

## Data Flow and Interactions
1. Client → API Gateway → User Service
2. User Service writes to PostgreSQL and emits “user.created” to Kafka
3. Order Service consumes Kafka events, processes orders, writes to PostgreSQL, emits “order.processed”
4. Notification Service consumes events and sends emails
5. Prometheus endpoints on each service scraped by Prometheus server

## Non-Functional Requirements
* Service availability ≥ 99.5%
* Average request latency < 200 ms
* Infrastructure provisioning idempotent

## Security Considerations
* Secure Kafka traffic with TLS
* Use AWS IAM roles for least-privilege access
* Enable encryption at rest for RDS and MSK

## Deployment Strategy
* GitOps: push Terraform code to infra repo → pipeline runs terraform apply
* Blue/green deploy services via ECS or k8s rollout

## Testing Strategy
* Unit and integration tests for each service
* Terraform plan validation in CI
* Chaos testing: randomly kill service instances to validate resilience

## Timeline and Milestones
Week 1: Service skeletons and local Docker Compose
Week 2: Terraform infra MVP
Week 3: Kafka integration and end-to-end flows
Week 4: Observability dashboards and docs

## Maintenance & Monitoring
* Alert on high error rates via Grafana
* Monthly Terraform drift detection