# Load the restart_process extension
load('ext://restart_process', 'docker_build_with_restart')

# Allow Tilt to use either Minikube or Docker Desktop ('minikube', 'docker-desktop')
allow_k8s_contexts('docker-desktop')

# API Server
docker_build_with_restart(
    'api-server',
    './api',
    dockerfile='api/Dockerfile',
    entrypoint=['./api-server'],
    live_update=[
        sync('./api', '/app'),
        run('cd /app && go mod init github.com/yourusername/tilt-go/api || true'),
        run('cd /app && go mod tidy'),
        run('cd /app && CGO_ENABLED=0 GOOS=linux go build -o api-server'),
    ]
)

k8s_yaml('k8s/api.yaml')
k8s_resource('api-server', port_forwards=8080)

# Worker
docker_build_with_restart(
    'worker',
    './worker',
    dockerfile='worker/Dockerfile',
    entrypoint=['./worker'],
    live_update=[
        sync('./worker', '/app'),
        run('cd /app && go mod init github.com/yourusername/tilt-go/worker || true'),
        run('cd /app && go mod tidy'),
        run('cd /app && CGO_ENABLED=0 GOOS=linux go build -o worker'),
    ]
)

k8s_yaml('k8s/worker.yaml')
k8s_resource('worker')

# Enable live updates for both services
k8s_resource('api-server', labels=['api'])
k8s_resource('worker', labels=['worker']) 