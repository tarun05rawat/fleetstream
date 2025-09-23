Overview

Build FactoryFlow, a distributed manufacturing orchestration platform that simulates a Tesla-style production line.
The goal is to demonstrate Tesla-relevant skills: Go microservices, Kubernetes, Kafka, distributed computing, TCP/IP networking, front-end React dashboarding, and clean, testable code.

1️⃣ High-Level Architecture
┌────────────────────────────────────────┐
│ React + TypeScript Dashboard │
│ (WebSocket + REST API) │
└────────────────────────────────────────┘
▲
│
┌───────────────────────────┐
│ FactoryFlow Backend (Go) │
│ - REST & WebSocket APIs │
│ - Stream Processor (Kafka)│
│ - Fault Detection │
│ - Data Persistence (PostgreSQL/MongoDB)│
└───────────────────────────┘
▲
│
┌───────────────────────────┐ ┌──────────────────────────────┐
│ Sensor Simulator Service │ ... │ Additional Machine Simulators │
│ (Go or Python) │ │ (optional) │
│ - Emits JSON over Kafka │ │ - Robot arms, conveyors etc. │
│ - Random + fault events │ │ │
└───────────────────────────┘ └──────────────────────────────┘

Deployment: Docker containers orchestrated by Kubernetes (minikube or k3d for local dev).

2️⃣ Tech Stack
Layer Tech Purpose
Sensor simulation Go or Python Generates machine-like sensor events over TCP/IP or Kafka
Messaging Kafka Reliable, fault-tolerant event streaming
Backend Go REST & WebSocket API, distributed processing
Data store PostgreSQL or MongoDB Store events and historical logs
Frontend React + TypeScript Real-time dashboard with live graphs and controls
Deployment Docker + Kubernetes Microservices orchestration and scaling
Optional Grafana/Prometheus Metrics and monitoring
3️⃣ Functional Requirements

Real-time Sensor Streams

Conveyor speed, temperature, robot arm angle, fault events (e.g. jam/overheat).

Adjustable frequency (e.g. every 100 ms).

Fault Detection & Alerts

Detect out-of-bound sensor values and push alerts to dashboard.

Dynamic Process Flow Control

Ability to change process parameters from the front-end, persisted in DB.

Historical Data Replay

Reconstruct production flow for debugging and optimization.

4️⃣ Implementation Plan (Stepwise)
Phase A – Sensor Simulator

Microservice that emits JSON messages like:

{ "timestamp": "2025-09-23T10:00:00Z",
"conveyor_speed": 1.5,
"temperature": 72.4,
"status": "ok" }

Publish to Kafka topics (e.g., line1.sensor).

Phase B – Backend Go Service

Connect to Kafka, consume events, validate and enrich.

Expose REST and WebSocket APIs:

GET /api/events – recent events

WS /ws/live – real-time push to dashboard

Implement anomaly detection (simple thresholds or sliding window analysis).

Store events in PostgreSQL/MongoDB.

Phase C – React + TypeScript Dashboard

Display live charts using libraries like Recharts or Chart.js.

Show machine status indicators and alerts.

Provide controls to adjust simulator parameters (frequency, thresholds).

Phase D – Deployment

Containerize each service (sensor, backend, frontend, Kafka).

Define docker-compose for local dev and Helm charts for Kubernetes.

Optional: Add Prometheus/Grafana monitoring.

Please act as a senior full-stack engineer.
Use this context as the long-term project brief.
Whenever I ask for code (e.g., “write the Go Kafka producer”), generate complete, production-grade code files with comments, modular design, and test coverage.
Keep your explanations concise and code-heavy.
# distributed-systems
