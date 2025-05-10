# Load the restart_process extension
load('ext://restart_process', 'docker_build_with_restart')

# Allow Tilt to use either Minikube or Docker Desktop ('minikube', 'docker-desktop')
allow_k8s_contexts('docker-desktop')

# Build and deploy API
docker_build_with_restart(
    'api',
    'api',
    dockerfile='api/Dockerfile',
    entrypoint=['./main'],
    live_update=[
        sync('api', '/app'),
        run('cd /app && go mod tidy'),
        run('cd /app && go build -o main .')
    ]
)
k8s_yaml('k8s/api.yaml')
k8s_resource('api')

# Build and deploy Worker
docker_build_with_restart(
    'worker',
    'worker',
    dockerfile='worker/Dockerfile',
    entrypoint=['./main'],
    live_update=[
        sync('worker', '/app'),
        run('cd /app && go mod tidy'),
        run('cd /app && go build -o main .')
    ]
)
k8s_yaml('k8s/worker.yaml')
k8s_resource('worker')

# Deploy observability stack
k8s_yaml('k8s/observability/prometheus.yaml')
k8s_yaml('k8s/observability/grafana.yaml')
k8s_resource('prometheus')
k8s_resource('grafana') 