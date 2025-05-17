# Tilt.dev Kubernetes Example

Este projeto demonstra como usar Tilt.dev para desenvolvimento local com Kubernetes usando Go. A aplicação consiste em dois componentes:

1. **API Server**: Servidor HTTP com logs estruturados usando Zap e endpoints de saúde.
2. **Worker**: Processo em background que executa uma tarefa agendada com logs estruturados e endpoint de saúde.

## Estrutura do Projeto

```
.
├── api/                # Código do servidor API
│   ├── main.go         # Implementação do servidor API
│   └── Dockerfile      # Definição do container da API
├── worker/             # Código do worker
│   ├── main.go         # Implementação do worker
│   └── Dockerfile      # Definição do container do worker
├── k8s/                # Manifests Kubernetes
│   ├── api.yaml        # Deployment da API
│   └── worker.yaml     # Deployment do worker
├── Makefile            # Comandos de build e desenvolvimento
├── Tiltfile            # Configuração do Tilt.dev
└── README.md           # Documentação do projeto
```

## Pré-requisitos

- Go 1.21 ou superior
- Docker
- Kubernetes (local ou remoto)
- Tilt.dev CLI

## Configuração do Kubernetes Local

Você tem várias opções para rodar Kubernetes localmente:

### 1. Docker Desktop com Kubernetes (Windows + WSL2)

1. Instale o Docker Desktop para Windows
2. Ative o WSL2 e instale o Ubuntu
3. No Docker Desktop:
   - Vá em Settings > Kubernetes
   - Ative o Kubernetes
   - Clique em "Apply & Restart"
4. Aguarde o cluster iniciar

### 2. MicroK8s (Ubuntu)

1. Instale o MicroK8s:
   ```bash
   # Instalação
   sudo snap install microk8s --classic

   # Adicione seu usuário ao grupo microk8s
   sudo usermod -a -G microk8s $USER
   sudo chown -f -R $USER ~/.kube

   # Reinicie sua sessão ou execute:
   newgrp microk8s
   ```

2. Inicie o MicroK8s:
   ```bash
   # Iniciar
   microk8s start

   # Habilitar addons necessários
   microk8s enable dns storage

   # Verificar status
   microk8s status
   ```

3. Configure o kubectl:
   ```bash
   # Criar alias para kubectl
   echo "alias kubectl='microk8s kubectl'" >> ~/.bashrc
   source ~/.bashrc
   ```

### 3. Minikube

#### Ubuntu
```bash
# Instalar Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Iniciar Minikube
minikube start

# Verificar status
minikube status
```

#### macOS
```bash
# Instalar Minikube via Homebrew
brew install minikube

# Iniciar Minikube
minikube start

# Verificar status
minikube status
```

## Probes do Kubernetes

Os probes são verificações de saúde que o Kubernetes usa para monitorar a aplicação. Existem três tipos:

- **Liveness Probe**: Verifica se a aplicação está viva. Se falhar, o Kubernetes reinicia o pod.
- **Readiness Probe**: Verifica se a aplicação está pronta para receber tráfego. Se falhar, o pod é removido do balanceamento de carga.
- **Startup Probe**: Verifica se a aplicação iniciou corretamente. Se falhar, o pod é reiniciado.

### Exemplo de Configuração

```yaml
# Exemplo para a API (k8s/api.yaml)
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5    # Tempo de espera antes da primeira verificação
  periodSeconds: 10         # Intervalo entre verificações
  timeoutSeconds: 2         # Tempo máximo para a resposta
  failureThreshold: 3       # Número de falhas antes de reiniciar
  successThreshold: 1       # Número de sucessos para considerar saudável

readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 3    # Menor que liveness para começar a receber tráfego mais rápido
  periodSeconds: 5          # Verificações mais frequentes que liveness
  timeoutSeconds: 1         # Timeout menor que liveness
  failureThreshold: 2       # Menos tentativas que liveness
  successThreshold: 1       # Um sucesso é suficiente

startupProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 0    # Começa imediatamente
  periodSeconds: 5          # Verifica a cada 5 segundos
  timeoutSeconds: 1         # Timeout de 1 segundo
  failureThreshold: 30      # Permite até 30 falhas (2.5 minutos) para iniciar
  successThreshold: 1       # Um sucesso é suficiente
```

### Explicação dos Parâmetros

- **initialDelaySeconds**: Tempo de espera antes da primeira verificação
  - Liveness: 5s - Dá tempo para a aplicação inicializar
  - Readiness: 3s - Começa a verificar mais cedo
  - Startup: 0s - Começa imediatamente

- **periodSeconds**: Intervalo entre verificações
  - Liveness: 10s - Verificações menos frequentes
  - Readiness: 5s - Verificações mais frequentes
  - Startup: 5s - Verificações moderadas

- **timeoutSeconds**: Tempo máximo para a resposta
  - Liveness: 2s - Mais tolerante
  - Readiness: 1s - Mais rigoroso
  - Startup: 1s - Mais rigoroso

- **failureThreshold**: Número de falhas antes de tomar ação
  - Liveness: 3 - Mais tolerante a falhas
  - Readiness: 2 - Menos tolerante
  - Startup: 30 - Muito tolerante para dar tempo de iniciar

- **successThreshold**: Número de sucessos para considerar saudável
  - Todos: 1 - Um sucesso é suficiente

### Endpoints de Saúde

A aplicação expõe dois endpoints de saúde:

- `/healthz`: Endpoint simples para probes do Kubernetes
  - Retorna 200 OK se a aplicação está saudável
  - Usado pelos probes do Kubernetes

- `/health`: Endpoint mais detalhado para monitoramento
  - Retorna informações adicionais sobre a saúde da aplicação
  - Pode ser usado para monitoramento externo

## Tilt.dev com Golang

O Tilt.dev é uma ferramenta que simplifica o desenvolvimento de aplicações em Kubernetes. Com Golang, ele oferece:

1. **Live Reload**: Atualiza automaticamente os containers quando há mudanças no código
2. **Build Otimizado**: Utiliza cache de camadas do Docker para builds mais rápidos
3. **Logs em Tempo Real**: Mostra logs dos containers em tempo real
4. **Deploy Automático**: Faz deploy das alterações no Kubernetes automaticamente

### Como Funciona

1. O Tiltfile configura:
   - Quais arquivos monitorar para mudanças
   - Como construir as imagens Docker
   - Como fazer deploy no Kubernetes
   - Como expor os serviços

2. Quando você faz uma alteração:
   - Tilt detecta a mudança
   - Reconstrói apenas o container afetado
   - Faz deploy da nova versão
   - Atualiza os logs em tempo real

## Iniciando o Projeto

1. Instale o Tilt:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash
   ```

2. Inicie o ambiente:
   ```bash
   make tilt-up
   ```

## Acessando a Aplicação

- API Server: http://localhost:8080
- Health Check API: http://localhost:8080/health
- Health Check Worker: http://localhost:8081/health
- Healthz Endpoints (usados pelos probes):
  - API: http://localhost:8080/healthz
  - Worker: http://localhost:8081/healthz
- Logs: Disponíveis no terminal do Tilt

## Características da Aplicação

### API Server
- Servidor HTTP com logs estruturados usando Zap
- Endpoint de health check (`/health`) e probe (`/healthz`)
- Logs em formato JSON com timestamp em UTC-3

### Worker
- Processo em background com agendamento
- Executa tarefa a cada minuto (configurável)
- Logs estruturados com Zap
- Timestamp em UTC-3 (horário de Brasília)
- Endpoint de health check (`/health`) e probe (`/healthz`) em um pequeno servidor HTTP dedicado

## Autor

**Renato Magalhães**

## Repositório

Este projeto está hospedado em: [https://github.com/renatomagalhaes/tilt-go](https://github.com/renatomagalhaes/tilt-go)

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para detalhes. 