# Carrega a extensão restart_process que permite reiniciar o processo dentro do container
load('ext://restart_process', 'docker_build_with_restart')

# Define quais contextos Kubernetes são permitidos
# Neste caso, apenas o Docker Desktop está habilitado
allow_k8s_contexts('docker-desktop')

# Configuração do servidor API
docker_build_with_restart(
    'api-server',                    # Nome da imagem Docker
    './api',                         # Diretório do código fonte
    dockerfile='api/Dockerfile.dev', # Caminho do Dockerfile
    entrypoint=["/app/api-server"],
    live_update=[
        sync('./api', '/app'),
        run('cd /app && go mod tidy && go build -o /app/api-server .'),
    ]
)

# Carrega os manifests Kubernetes da API
k8s_yaml('k8s/api-config.yaml')     # Carrega o ConfigMap primeiro
k8s_yaml('k8s/api.yaml')            # Depois carrega o deployment
k8s_resource('api-server', port_forwards=8080, labels=['api'])

# Configuração do Worker
docker_build_with_restart(
    'worker-server',                 # Nome da imagem Docker
    './worker',                      # Diretório do código fonte
    dockerfile='worker/Dockerfile.dev',  # Caminho do Dockerfile
    entrypoint=["/app/worker-server"],
    live_update=[
        sync('./worker', '/app'),
        run('cd /app && go mod tidy && go build -o /app/worker-server .'),
    ]
)

# Carrega os manifests Kubernetes do Worker
k8s_yaml('k8s/worker-config.yaml')  # Carrega o ConfigMap primeiro
k8s_yaml('k8s/worker.yaml')         # Depois carrega o deployment
k8s_resource('worker-server', port_forwards=8081, labels=['worker'])

# Cria recursos Tilt para os ConfigMaps e aplica labels
k8s_resource(
    objects=['api-config:configmap'],
    new_name='api-config',
    labels=['api']
)

k8s_resource(
    objects=['worker-config:configmap'],
    new_name='worker-config',
    labels=['worker']
) 