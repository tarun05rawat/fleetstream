# FactoryFlow Deployment Guide

This document provides comprehensive instructions for deploying the FactoryFlow distributed manufacturing orchestration platform.

## Prerequisites

### Development Environment
- Docker 20.10+
- Docker Compose 2.0+
- Node.js 18+
- Go 1.21+
- Make utility

### Production Environment (Kubernetes)
- Kubernetes cluster (1.25+)
- kubectl configured
- Ingress controller (NGINX recommended)
- Persistent storage provider

### Local Development with Minikube
- Minikube 1.30+
- At least 4GB RAM available for minikube

## Quick Start (Development)

1. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd "Distributed Systems Project"
   make setup
   ```

2. **Start Development Environment**
   ```bash
   make dev
   ```

3. **Access Applications**
   - Frontend Dashboard: http://localhost:3000
   - Backend API: http://localhost:8080
   - API Health: http://localhost:8080/health

4. **View Logs**
   ```bash
   make dev-logs
   ```

## Production Deployment (Kubernetes)

### 1. Build and Push Images

```bash
# Build all Docker images
make build

# Push to your registry (update DOCKER_REGISTRY in Makefile)
make push
```

### 2. Deploy to Kubernetes

```bash
# Deploy all components
make k8s-deploy

# Check deployment status
make k8s-status

# View logs
make k8s-logs
```

### 3. Access Applications

Add the following to your `/etc/hosts` file (or configure DNS):
```
<your-cluster-ip> factoryflow.local
<your-cluster-ip> api.factoryflow.local
```

- Frontend: http://factoryflow.local
- Backend API: http://api.factoryflow.local

## Local Kubernetes Development (Minikube)

### 1. Deploy to Minikube

```bash
# Start minikube, build images, and deploy
make minikube-deploy
```

### 2. Configure Host Access

The deployment script will show you the minikube IP. Add to `/etc/hosts`:
```bash
# Example (replace with actual minikube IP)
192.168.49.2 factoryflow.local
192.168.49.2 api.factoryflow.local
```

### 3. Access Services

In a separate terminal, create a tunnel:
```bash
make minikube-tunnel
```

Now access:
- Frontend: http://factoryflow.local
- Backend API: http://api.factoryflow.local

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌───────────────────┐
│   Frontend      │    │     Backend      │    │ Sensor Simulator  │
│  (React/TS)     │────│   (Go/Kafka)     │────│     (Go)          │
│  Port: 3000     │    │   Port: 8080     │    │   Multiple Pods   │
└─────────────────┘    └──────────────────┘    └───────────────────┘
         │                        │                        │
         │                        │                        │
         └────────────────────────┼────────────────────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │                           │
            ┌───────▼────────┐         ┌───────▼────────┐
            │   PostgreSQL   │         │     Kafka      │
            │   Database     │         │   Messaging    │
            │   Port: 5432   │         │   Port: 9092   │
            └────────────────┘         └────────────────┘
```

## Service Configuration

### Environment Variables

#### Backend Service
- `KAFKA_BROKERS`: Kafka broker addresses
- `DB_HOST`: PostgreSQL host
- `DB_PORT`: PostgreSQL port
- `DB_NAME`: Database name
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password

#### Sensor Simulator
- `KAFKA_BROKERS`: Kafka broker addresses
- `KAFKA_TOPIC`: Topic to publish events
- `MACHINE_ID`: Unique machine identifier
- `SENSOR_FREQUENCY`: Event generation frequency (ms)

#### Frontend
- `REACT_APP_API_URL`: Backend API URL
- `REACT_APP_WS_URL`: WebSocket URL

### Resource Requirements

#### Minimum Resources (Development)
- CPU: 2 cores
- RAM: 4GB
- Storage: 10GB

#### Production Resources (per component)
- **Frontend**: 100m CPU, 128Mi RAM
- **Backend**: 250m CPU, 256Mi RAM
- **Sensor Simulator**: 100m CPU, 64Mi RAM
- **PostgreSQL**: 500m CPU, 512Mi RAM
- **Kafka**: 500m CPU, 512Mi RAM
- **Zookeeper**: 250m CPU, 256Mi RAM

## Monitoring and Observability

### Health Checks
- Backend health: `GET /health`
- Frontend health: `GET /health`

### Logs
```bash
# Development
make dev-logs

# Kubernetes
make k8s-logs

# Specific component
kubectl logs -n factoryflow -l app=backend -f
```

### Metrics
The system provides WebSocket-based real-time metrics:
- Connected clients count
- Event processing rates
- System uptime percentage
- Fault detection alerts

## Troubleshooting

### Common Issues

1. **Services not starting**
   ```bash
   # Check pod status
   kubectl get pods -n factoryflow

   # Check logs
   kubectl logs -n factoryflow <pod-name>
   ```

2. **Database connection issues**
   ```bash
   # Port forward to database
   kubectl port-forward -n factoryflow svc/postgres-service 5432:5432

   # Test connection
   psql -h localhost -p 5432 -U factoryuser -d factoryflow
   ```

3. **Kafka connectivity issues**
   ```bash
   # Check Kafka logs
   kubectl logs -n factoryflow -l app=kafka

   # Port forward to Kafka
   kubectl port-forward -n factoryflow svc/kafka-service 9092:9092
   ```

4. **Frontend not loading**
   - Check ingress configuration
   - Verify DNS/hosts file entries
   - Check browser console for API connection errors

### Performance Optimization

1. **Scale replicas based on load**
   ```bash
   kubectl scale deployment backend -n factoryflow --replicas=3
   kubectl scale deployment frontend -n factoryflow --replicas=2
   ```

2. **Adjust resource limits**
   Edit the YAML files in `k8s/` directory and redeploy.

3. **Database tuning**
   - Monitor PostgreSQL performance
   - Adjust connection pooling
   - Consider read replicas for high load

### Cleanup

```bash
# Stop development environment
make dev-stop

# Clean development environment
make dev-clean

# Delete Kubernetes deployment
make k8s-delete

# Clean minikube
make minikube-clean

# Clean Docker resources
make clean
```

## Security Considerations

### Development
- Default passwords are used for convenience
- Services are accessible without authentication
- Data is not encrypted in transit

### Production Recommendations
- Change all default passwords
- Enable TLS/SSL for all services
- Implement authentication and authorization
- Use Kubernetes secrets for sensitive data
- Enable network policies
- Use pod security policies
- Regular security updates

## Backup and Recovery

### Database Backup
```bash
# Create backup
kubectl exec -n factoryflow <postgres-pod> -- pg_dump -U factoryuser factoryflow > backup.sql

# Restore backup
kubectl exec -i -n factoryflow <postgres-pod> -- psql -U factoryuser factoryflow < backup.sql
```

### Persistent Data
- PostgreSQL data is stored in PersistentVolumeClaims
- Ensure regular backups of PV data
- Consider using cloud provider backup solutions

## Scaling Guidelines

### Horizontal Scaling
- Backend: Can be scaled horizontally (stateless)
- Frontend: Can be scaled horizontally (static content)
- Sensor Simulators: Can be scaled to simulate more machines
- Database: Single instance (consider read replicas)
- Kafka: Can be scaled with proper configuration

### Vertical Scaling
- Adjust resource limits based on monitoring metrics
- Monitor CPU and memory usage patterns
- Consider auto-scaling based on metrics

## Support

For issues and questions:
1. Check logs using provided commands
2. Review this deployment guide
3. Check Kubernetes resources status
4. Verify configuration files