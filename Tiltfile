# Load environment variables
load('ext://dotenv', 'dotenv')
dotenv()

# Allow Tilt to use the local Docker daemon
allow_k8s_contexts('docker-desktop', 'minikube')

# API Server
docker_build(
    'api-server',
    './api',
    dockerfile='api/Dockerfile',
    live_update=[
        sync('./api', '/app'),
        run('go build -o api-server'),
        restart_container(),
    ]
)

k8s_yaml('k8s/api.yaml')
k8s_resource('api-server', port_forwards=8080)

# Worker
docker_build(
    'worker',
    './worker',
    dockerfile='worker/Dockerfile',
    live_update=[
        sync('./worker', '/app'),
        run('go build -o worker'),
        restart_container(),
    ]
)

k8s_yaml('k8s/worker.yaml')
k8s_resource('worker')

# Enable live updates for both services
k8s_resource('api-server', labels=['api'])
k8s_resource('worker', labels=['worker']) 