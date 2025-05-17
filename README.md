# Tilt.dev Kubernetes Example com Golang

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

- Go 1.23.0 ou superior (toolchain 1.23.1)
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
- Logs: Disponíveis no terminal do Tilt

## Características da Aplicação

### API Server
- Servidor HTTP com logs estruturados usando Zap
- Endpoints de probe (`/livez`, `/readyz`, `/healthz`)
- Logs em formato JSON com timestamp em UTC-3

### Worker
- Processo em background com agendamento
- Executa tarefa a cada minuto (configurável)
- Logs estruturados com Zap
- Timestamp em UTC-3 (horário de Brasília)
- Endpoints de probe (`/livez`, `/readyz`, `/healthz`) em um pequeno servidor HTTP dedicado

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
    path: /livez    # Endpoint específico para liveness
    port: 8080
  initialDelaySeconds: 5    # Tempo de espera antes da primeira verificação
  periodSeconds: 10         # Intervalo entre verificações
  timeoutSeconds: 2         # Tempo máximo para a resposta
  failureThreshold: 3       # Número de falhas antes de reiniciar
  successThreshold: 1       # Número de sucessos para considerar saudável

readinessProbe:
  httpGet:
    path: /readyz   # Endpoint específico para readiness
    port: 8080
  initialDelaySeconds: 3    # Menor que liveness para começar a receber tráfego mais rápido
  periodSeconds: 5          # Verificações mais frequentes que liveness
  timeoutSeconds: 1         # Timeout menor que liveness
  failureThreshold: 2       # Menos tentativas que liveness
  successThreshold: 1       # Um sucesso é suficiente

startupProbe:
  httpGet:
    path: /healthz  # Endpoint específico para startup
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

A aplicação expõe endpoints de saúde seguindo as melhores práticas do Kubernetes:

- `/livez`: Endpoint para Liveness Probe
  - Verifica se o processo está vivo
  - Deve ser rápido e leve
  - Não verifica dependências externas
  - Usado pelo Kubernetes para decidir se deve reiniciar o pod

- `/readyz`: Endpoint para Readiness Probe
  - Verifica se a aplicação está pronta para receber tráfego
  - Pode verificar dependências (banco de dados, cache, etc.)
  - Usado pelo Kubernetes para balanceamento de carga

- `/healthz`: Endpoint para Startup Probe
  - Verifica se a aplicação iniciou corretamente
  - Similar ao liveness, mas com threshold mais alto
  - Usado pelo Kubernetes durante a inicialização do pod

## Graceful Shutdown

O graceful shutdown é uma prática importante que permite que a aplicação encerre suas operações de forma ordenada quando recebe um sinal de término (SIGTERM ou SIGINT). Isso é crucial para:

1. **Integridade dos Dados**: Garantir que operações em andamento sejam concluídas
2. **Conexões**: Fechar conexões com banco de dados e outros serviços adequadamente
3. **Logs**: Registrar informações importantes antes do encerramento
4. **Kubernetes**: Permitir que o Kubernetes gerencie o ciclo de vida dos pods corretamente

### Implementação

Tanto a API quanto o Worker implementam graceful shutdown usando:

```go
// Criação do canal para sinais
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// Aguarda o sinal
<-quit

// Executa o shutdown
logger.Info("service_shutting_down")
```

### Testando o Graceful Shutdown

Você pode testar o graceful shutdown de várias formas:

1. **Usando o Makefile**:
   ```bash
   # Testa o graceful shutdown da API
   make test-shutdown-api

   # Testa o graceful shutdown do Worker
   make test-shutdown-worker
   ```

2. **Manual**:
   ```bash
   # Encontra o PID do processo
   ps aux | grep api-server
   # ou
   ps aux | grep worker-server

   # Envia o sinal SIGTERM
   kill -TERM <PID>
   ```

3. **Via Kubernetes**:
   ```bash
   # Escala o deployment para 0
   kubectl scale deployment api-server --replicas=0
   # ou
   kubectl scale deployment worker-server --replicas=0
   ```

### O que Observar

Ao testar o graceful shutdown, observe:

1. **Logs**: Deve aparecer a mensagem "service_shutting_down"
2. **Tempo**: O processo deve encerrar em até 30 segundos (default do Kubernetes)
3. **Conexões**: Conexões ativas devem ser fechadas adequadamente
4. **Estado**: O estado da aplicação deve estar consistente

### Exemplo de Teste

```bash
# 1. Inicie a aplicação
make tilt-up

# 2. Em outro terminal, teste o graceful shutdown
make test-shutdown-api

# 3. Observe os logs no terminal do Tilt
# Você deve ver:
# - "service_shutting_down"
# - "service_stopped"
# - Conexões sendo fechadas adequadamente
```

## Desenvolvimento com Dependências

O projeto usa uma configuração híbrida para desenvolvimento:

1. **Ambiente Go**: Container Docker com Go instalado
2. **Dependências**: Serviços gerenciados via Docker Compose
3. **Hot Reload**: Código executado via `go run` dentro do container

### Serviços Disponíveis

- **MySQL**: `localhost:3306`
  - Usuário: root
  - Senha: root
  - Database: app

- **Memcached**: `localhost:11211`

- **RabbitMQ**: `localhost:5672`
  - Management UI: `localhost:15672`
  - Usuário: guest
  - Senha: guest

### Como Usar

1. Inicie o ambiente:
   ```bash
   make tilt-up
   ```

2. Acesse os serviços:
   - API: http://localhost:8080
   - Worker: http://localhost:8081
   - RabbitMQ Management: http://localhost:15672

3. Monitore os logs:
   - Logs da aplicação: Terminal do Tilt
   - Logs dos serviços: `docker-compose logs -f [serviço]`

### Estrutura de Desenvolvimento

```
.
├── api/
│   ├── Dockerfile.dev    # Dockerfile para desenvolvimento
│   └── main.go
├── worker/
│   ├── Dockerfile.dev    # Dockerfile para desenvolvimento
│   └── main.go
├── docker-compose.dev.yml # Serviços de dependência
├── k8s/
│   ├── api-config.yaml   # Configurações da API
│   └── worker-config.yaml # Configurações do Worker
└── Tiltfile             # Configuração do Tilt
```

### Comandos Úteis

```bash
# Iniciar ambiente
make tilt-up

# Parar ambiente
make tilt-down

# Ver logs dos serviços
make logs-services

# Reiniciar serviços
make restart-services

# Limpar ambiente
make clean
``` 

## Autor

**Renato Magalhães**

## Repositório

Este projeto está hospedado em: [https://github.com/renatomagalhaes/tilt-go](https://github.com/renatomagalhaes/tilt-go)

## Licença