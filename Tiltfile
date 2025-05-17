# Carrega a extensão restart_process que permite reiniciar o processo dentro do container
# quando há mudanças no código, sem precisar reconstruir a imagem
load('ext://restart_process', 'docker_build_with_restart')

# Define quais contextos Kubernetes são permitidos
# Neste caso, apenas o Docker Desktop está habilitado
# Exemplos de outros contextos (descomente a linha que deseja usar):
# allow_k8s_contexts('docker-desktop')  # Para Windows com WSL2 + Docker Desktop
# allow_k8s_contexts('minikube')       # Para Minikube
# allow_k8s_contexts('microk8s')       # Para MicroK8s
allow_k8s_contexts('docker-desktop')

# Configuração do servidor API
docker_build_with_restart(
    'api-server',                    # Nome da imagem Docker
    './api',                         # Diretório do código fonte
    dockerfile='api/Dockerfile',     # Caminho do Dockerfile
    entrypoint=['./api-server'],     # Comando para iniciar a aplicação
    live_update=[                    # Configuração de atualização em tempo real
        # Sincroniza os arquivos do diretório ./api para /app no container
        sync('./api', '/app'),
        # Inicializa o módulo Go (se não existir)
        run('cd /app && go mod init github.com/yourusername/tilt-go/api || true'),
        # Atualiza as dependências
        run('cd /app && go mod tidy'),
        # Compila a aplicação
        run('cd /app && CGO_ENABLED=0 GOOS=linux go build -o api-server'),
    ]
)

# Carrega o manifesto Kubernetes da API
k8s_yaml('k8s/api.yaml')
# Configura o redirecionamento de porta 8080 para a API
k8s_resource('api-server', port_forwards=8080)

# Configuração do Worker
docker_build_with_restart(
    'worker-server',                 # Nome da imagem Docker
    './worker',                      # Diretório do código fonte
    dockerfile='worker/Dockerfile',  # Caminho do Dockerfile
    entrypoint=['./worker-server'],  # Comando para iniciar a aplicação
    live_update=[                    # Configuração de atualização em tempo real
        # Sincroniza os arquivos do diretório ./worker para /app no container
        sync('./worker', '/app'),
        # Inicializa o módulo Go (se não existir)
        run('cd /app && go mod init github.com/yourusername/tilt-go/worker || true'),
        # Atualiza as dependências
        run('cd /app && go mod tidy'),
        # Compila a aplicação
        run('cd /app && CGO_ENABLED=0 GOOS=linux go build -o worker-server'),
    ]
)

# Carrega o manifesto Kubernetes do Worker
k8s_yaml('k8s/worker.yaml')
k8s_resource('worker-server')

# Habilita atualizações em tempo real para ambos os serviços
# As labels são usadas para agrupar os recursos no UI do Tilt
k8s_resource('api-server', labels=['api'])
k8s_resource('worker-server', labels=['worker']) 