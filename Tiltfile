# Extensões do Tilt
load('ext://docker_compose', 'docker_compose')

# Carrega as dependências via docker-compose
docker_compose('docker-compose.dev.yml')

# Configuração do Go para desenvolvimento
docker_build(
    'api-dev',
    'api',
    dockerfile='api/Dockerfile.dev',
    live_update=[
        # Monta o código fonte
        sync('api', '/app'),
        # Executa go run
        run('cd /app && go run main.go'),
    ]
)

docker_build(
    'worker-dev',
    'worker',
    dockerfile='worker/Dockerfile.dev',
    live_update=[
        sync('worker', '/app'),
        run('cd /app && go run main.go'),
    ]
)

# Configuração do Kubernetes
k8s_yaml('k8s/api.yaml')
k8s_yaml('k8s/worker.yaml')

# Configuração dos ConfigMaps
k8s_resource(
    'api-config',
    labels=['api', 'config'],
    objects=['api-config']
)

k8s_resource(
    'worker-config',
    labels=['worker', 'config'],
    objects=['worker-config']
)

# Labels para organização no Tilt
k8s_resource('api-server', labels=['api'])
k8s_resource('worker-server', labels=['worker']) 