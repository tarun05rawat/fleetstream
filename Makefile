# FactoryFlow - Makefile for development and deployment

.PHONY: help dev build push deploy clean test lint

# Variables
PROJECT_NAME := factoryflow
DOCKER_REGISTRY := factoryflow
VERSION := latest

# Docker Images
SENSOR_IMAGE := $(DOCKER_REGISTRY)/sensor-simulator:$(VERSION)
BACKEND_IMAGE := $(DOCKER_REGISTRY)/backend:$(VERSION)
FRONTEND_IMAGE := $(DOCKER_REGISTRY)/frontend:$(VERSION)

help: ## Show this help message
	@echo 'Usage: make [TARGET]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment with Docker Compose
	docker-compose up --build -d
	@echo "Development environment started!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend API: http://localhost:8080"
	@echo "Database: localhost:5432"

dev-logs: ## Show logs from development environment
	docker-compose logs -f

dev-stop: ## Stop development environment
	docker-compose down

dev-clean: ## Clean development environment (remove volumes)
	docker-compose down -v
	docker system prune -f

build: ## Build all Docker images
	@echo "Building sensor-simulator..."
	docker build -t $(SENSOR_IMAGE) ./sensor-simulator
	@echo "Building backend..."
	docker build -t $(BACKEND_IMAGE) ./backend
	@echo "Building frontend..."
	docker build -t $(FRONTEND_IMAGE) ./frontend
	@echo "All images built successfully!"

push: build ## Push Docker images to registry
	docker push $(SENSOR_IMAGE)
	docker push $(BACKEND_IMAGE)
	docker push $(FRONTEND_IMAGE)
	@echo "All images pushed successfully!"

# Kubernetes deployment
k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -k k8s/
	@echo "Deployed to Kubernetes!"
	@echo "Wait for pods to be ready with: make k8s-status"

k8s-status: ## Check Kubernetes deployment status
	kubectl get pods -n $(PROJECT_NAME)
	@echo ""
	kubectl get services -n $(PROJECT_NAME)
	@echo ""
	kubectl get ingress -n $(PROJECT_NAME)

k8s-logs: ## Show logs from Kubernetes pods
	kubectl logs -n $(PROJECT_NAME) -l app=backend --tail=50 -f

k8s-delete: ## Delete Kubernetes deployment
	kubectl delete -k k8s/
	@echo "Kubernetes deployment deleted!"

k8s-restart: ## Restart Kubernetes deployments
	kubectl rollout restart deployment -n $(PROJECT_NAME)

# Local development with minikube
minikube-start: ## Start minikube and enable ingress
	minikube start --memory=4096 --cpus=2
	minikube addons enable ingress
	@echo "Minikube started with ingress enabled!"

minikube-deploy: minikube-start build ## Deploy to minikube
	@echo "Loading images to minikube..."
	minikube image load $(SENSOR_IMAGE)
	minikube image load $(BACKEND_IMAGE)
	minikube image load $(FRONTEND_IMAGE)
	$(MAKE) k8s-deploy
	@echo "Getting minikube IP..."
	@minikube ip
	@echo "Add the following to your /etc/hosts file:"
	@echo "$$(minikube ip) factoryflow.local"
	@echo "$$(minikube ip) api.factoryflow.local"

minikube-tunnel: ## Create tunnel to access services (run in separate terminal)
	minikube tunnel

minikube-clean: ## Stop and delete minikube
	minikube stop
	minikube delete

# Testing
test: ## Run tests for all components
	@echo "Running backend tests..."
	cd backend && go test ./... -v
	@echo "Running frontend tests..."
	cd frontend && npm test -- --coverage --watchAll=false

lint: ## Run linters for all components
	@echo "Linting backend..."
	cd backend && go fmt ./... && go vet ./...
	@echo "Linting frontend..."
	cd frontend && npm run lint

# Database operations
db-migrate: ## Run database migrations
	@echo "Database schema is initialized via init.sql in PostgreSQL container"

db-seed: ## Seed database with test data
	@echo "Test data is seeded via init.sql in PostgreSQL container"

# Monitoring (optional)
monitoring-deploy: ## Deploy Prometheus and Grafana
	kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml
	@echo "Monitoring stack deployed!"

# Cleanup
clean: ## Clean up Docker resources
	docker system prune -af
	docker volume prune -f
	@echo "Docker cleanup completed!"

# Environment setup
setup: ## Initial setup for development
	@echo "Setting up development environment..."
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed. Aborting." >&2; exit 1; }
	@echo "Creating .env files..."
	cp sensor-simulator/.env.example sensor-simulator/.env
	cp backend/.env.example backend/.env
	cp frontend/.env.example frontend/.env
	@echo "Setup completed! Run 'make dev' to start development environment."

# Port forwarding for K8s development
k8s-port-forward: ## Port forward services for local access
	@echo "Port forwarding services..."
	kubectl port-forward -n $(PROJECT_NAME) svc/backend-service 8080:8080 &
	kubectl port-forward -n $(PROJECT_NAME) svc/frontend-service 3000:3000 &
	kubectl port-forward -n $(PROJECT_NAME) svc/postgres-service 5432:5432 &
	@echo "Services are now accessible locally:"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend: http://localhost:8080"
	@echo "Database: localhost:5432"