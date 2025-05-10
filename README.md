# Tilt.dev Kubernetes Example

This project demonstrates how to use Tilt.dev for local Kubernetes development with a simple Go application. The application consists of two components:

1. **API Server**: A simple Hello World HTTP server
2. **Worker**: A background worker that runs a task every 5 minutes

## Project Structure

```
.
├── api/                 # API server code
│   ├── main.go         # API server implementation
│   └── Dockerfile      # API server container definition
├── worker/             # Worker code
│   ├── main.go         # Worker implementation
│   └── Dockerfile      # Worker container definition
├── k8s/                # Kubernetes manifests
│   ├── api.yaml        # API server deployment
│   └── worker.yaml     # Worker deployment
├── Makefile           # Build and development commands
├── .gitignore         # Git ignore rules
├── Tiltfile           # Tilt.dev configuration
└── README.md          # Project documentation
```

## Prerequisites

- Go 1.16 or later
- Docker
- Kubernetes cluster (local or remote)
- Tilt.dev CLI

## Local Kubernetes Setup

You have two options for running Kubernetes locally:

### Option 1: Docker Desktop with Kubernetes

1. Install Docker Desktop from [https://www.docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop)
2. Open Docker Desktop
3. Go to Settings > Kubernetes
4. Enable Kubernetes
5. Click "Apply & Restart"
6. Wait for Kubernetes to start

### Option 2: Minikube

1. Install Minikube:
   ```bash
   # For Linux
   curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
   sudo install minikube-linux-amd64 /usr/local/bin/minikube

   # For macOS
   brew install minikube
   ```

2. Start Minikube:
   ```bash
   minikube start
   ```

3. Verify installation:
   ```bash
   kubectl cluster-info
   ```

## Lens - Kubernetes IDE

Lens is a powerful IDE for managing Kubernetes clusters. It provides a user-friendly interface for:
- Monitoring pods, deployments, and services
- Viewing logs
- Managing resources
- Debugging applications

### Installing Lens

1. Download Lens from [https://k8slens.dev/](https://k8slens.dev/)
2. Install and launch Lens
3. Add your cluster:
   - For Docker Desktop: It should be automatically detected
   - For Minikube: Add the kubeconfig from `~/.kube/config`

### Using Lens with this Project

1. Open Lens
2. Connect to your local cluster
3. Navigate to the "Workloads" section
4. You should see:
   - `api-server` deployment and service
   - `worker` deployment
5. Click on any resource to:
   - View logs
   - Monitor metrics
   - Check pod status
   - Access the terminal

## Getting Started

1. Install Tilt.dev:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
   ```

2. Start the development environment:
   ```bash
   make tilt-up
   ```

## How it Works

### API Server
The API server is a simple HTTP server that responds with "Hello, World!" on the root endpoint.

### Worker
The worker is a background process that runs a task every 5 minutes. It demonstrates how to use cron-like functionality in a Kubernetes pod.

### Tilt.dev Configuration
The `Tiltfile` configures how Tilt.dev manages the development environment:
- Watches for file changes
- Builds Docker images
- Deploys to Kubernetes
- Provides live updates

## Development

1. Make changes to the code
2. Tilt.dev will automatically:
   - Rebuild the affected containers
   - Deploy the changes to Kubernetes
   - Show logs and status in the Tilt.dev UI

## Accessing the Application

- API Server: http://localhost:8080
- Worker logs: Available in the Tilt.dev UI or Lens
- Kubernetes Dashboard: Available through Lens

## Author

**Renato Magalhães**

## Repository

This project is hosted at: [https://github.com/renatomagalhaes/tilt-go](https://github.com/renatomagalhaes/tilt-go)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 