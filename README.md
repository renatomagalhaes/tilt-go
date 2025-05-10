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
├── Tiltfile            # Tilt.dev configuration
└── README.md           # This file
```

## Prerequisites

- Go 1.16 or later
- Docker
- Kubernetes cluster (local or remote)
- Tilt.dev CLI

## Getting Started

1. Install Tilt.dev:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
   ```

2. Start the development environment:
   ```bash
   tilt up
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
- Worker logs: Available in the Tilt.dev UI 