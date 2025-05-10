# Tilt Go Project

A Go application with API and Worker components, managed with Tilt.dev for local Kubernetes development.

## Prerequisites

- Go 1.21 or later
- Docker
- Tilt
- Kubernetes cluster (local or remote)

## Local Kubernetes Setup

### Option 1: Docker Desktop
1. Install Docker Desktop
2. Enable Kubernetes in Docker Desktop settings
3. Wait for the cluster to start
4. Verify with:
   ```bash
   kubectl cluster-info
   ```

### Option 2: Minikube
1. Install Minikube
2. Start the cluster:
   ```bash
   minikube start
   ```
3. Verify with:
   ```bash
   kubectl cluster-info
   ```

## Lens - Kubernetes IDE

Lens is a powerful IDE for managing Kubernetes clusters. It provides a user-friendly interface for:
- Viewing and managing pods, services, and deployments
- Real-time logs and metrics
- Terminal access to containers
- Resource monitoring

### Installation
1. Download Lens from [lens.dev](https://k8slens.dev/)
2. Install and launch Lens
3. Add your local cluster:
   - For Docker Desktop: Use the kubeconfig from `~/.kube/config`
   - For Minikube: Use `minikube kubeconfig`

## Observability Stack

The project includes a complete observability stack for monitoring and debugging:

### Components
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization and dashboards
- **Elasticsearch**: Log storage and search
- **Kibana**: Log visualization and analysis
- **Jaeger**: Distributed tracing

### Usage
1. Start the stack:
   ```bash
   make observability-up
   ```

2. Access the tools:
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)
   - Kibana: http://localhost:5601
   - Jaeger: http://localhost:16686

3. Stop the stack:
   ```bash
   make observability-down
   ```

## Development

1. Install dependencies:
   ```bash
   make install-tilt
   ```

2. Start the development environment:
   ```bash
   make tilt-up
   ```

3. Access the services:
   - API: http://localhost:8080
   - Worker: Running in the background

4. Stop the development environment:
   ```bash
   make tilt-down
   ```

## Project Structure

```
.
├── api/                    # API service
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── k8s/               # Kubernetes manifests
├── worker/                # Worker service
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── k8s/              # Kubernetes manifests
├── k8s/                   # Shared Kubernetes manifests
│   └── observability/    # Observability stack manifests
├── Tiltfile              # Tilt configuration
├── Makefile              # Build and development commands
└── README.md
```

## Author

Renato Magalhães

## Repository

https://github.com/renatomagalhaes/tilt-goseguindo 