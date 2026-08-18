# FleetStream

FleetStream is a real-time telemetry platform for simulating, ingesting, processing, and visualizing industrial sensor data. It combines a Go backend, Kafka-based event streaming, PostgreSQL storage, a React/TypeScript dashboard, and a sensor simulator to model factory-style operations with live monitoring and anomaly detection.

## Overview

The system is designed to emulate a lightweight manufacturing monitoring pipeline:

- sensors generate machine telemetry and fault events
- Kafka streams events through the platform in real time
- the backend processes, stores, and analyzes incoming data
- alerts and system metrics are pushed to clients over WebSockets
- the frontend dashboard displays live operational activity

## Features

- Real-time sensor event simulation
- Kafka-based event ingestion and streaming
- Live dashboard updates via WebSockets
- Anomaly detection for faults and warning conditions
- PostgreSQL persistence for events, alerts, machines, and process parameters
- Configurable deployment for local development and Kubernetes

## Tech Stack

**Frontend**
- React
- TypeScript
- Vite
- TailwindCSS
- Recharts

**Backend**
- Go
- Gin
- Gorilla WebSocket
- Sarama
- PostgreSQL

**Infrastructure**
- Kafka
- Zookeeper
- Docker
- Docker Compose
- Kubernetes

## Architecture

```text
Sensor Simulator -> Kafka -> Go Backend -> PostgreSQL
                                 |
                                 -> WebSocket Updates -> React Dashboard
Repository Structure
fleetstream/
├── backend/              # Go API, Kafka consumer, WebSocket server, anomaly logic
├── frontend/             # React + TypeScript dashboard
├── sensor-simulator/     # Go-based telemetry and fault event generator
├── database/             # PostgreSQL schema and seed data
├── k8s/                  # Kubernetes manifests
├── factoryflow-cdk/      # Infrastructure/deployment code
├── docker-compose.yml    # Local multi-service development setup
├── Makefile              # Common development and deployment commands
└── DEPLOYMENT.md         # Detailed deployment guide
How It Works
The sensor simulator generates telemetry for machines such as conveyors, robot arms, and sensor hubs.
Events are published to Kafka topics.
The backend consumes those events, stores them in PostgreSQL, and runs anomaly checks.
Alerts and system statistics are broadcast to connected clients over WebSockets.
The frontend displays live events, alerts, and operational metrics.
Data Model
The database includes the following core tables:
events — raw sensor and machine telemetry
alerts — anomaly and fault notifications
machines — machine metadata and configuration
process_parameters — tunable system parameters
Local Development
Prerequisites
Docker
Docker Compose
Go
Node.js
Run with Docker Compose
docker compose up --build
Default Local Endpoints
Frontend: http://localhost:3000
Backend API: http://localhost:8080
Health Check: http://localhost:8080/health
PostgreSQL: localhost:5432
Kafka: localhost:9092
API Endpoints
Some of the main backend routes include:
GET /health
GET /api/events
GET /api/events/stats
GET /api/alerts
PUT /api/alerts/:id/acknowledge
GET /api/parameters
PUT /api/parameters
GET /api/machines
GET /api/system/health
GET /api/anomaly/thresholds
PUT /api/anomaly/thresholds
GET /ws
Deployment
The project includes support for:
local Docker Compose development
Kubernetes deployment
additional cloud deployment configuration
For full deployment instructions, see [DEPLOYMENT.md](./DEPLOYMENT.md).