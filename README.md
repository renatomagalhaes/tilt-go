# Ambiente de Desenvolvimento Local com Tilt.dev, Golang e Docker Compose

Este projeto demonstra um ambiente de desenvolvimento local eficiente utilizando Tilt.dev para orquestração, Golang para os serviços da aplicação (API e Worker) rodando em Kubernetes, e Docker Compose para gerenciar as dependências.

Nesta branch, o foco é no fluxo de desenvolvimento local.

## Componentes Principais

1. **API Server**: Servidor HTTP em Go, responsável por lidar com as requisições.
2. **Worker**: Processo em background em Go, que executa tarefas agendadas.
3. **Dependências**: Serviços como MySQL, Memcached e RabbitMQ gerenciados via Docker Compose.
4. **Tilt.dev**: Orquestra a construção, implantação e live update dos serviços Go no Kubernetes e gerencia o ciclo de vida das dependências via Docker Compose.

## Estrutura do Projeto

```
.
├── api/                # Código do servidor API
│   ├── main.go         # Implementação do servidor API
│   └── Dockerfile.dev    # Definição do container de desenvolvimento da API
├── worker/             # Código do worker
│   ├── main.go         # Implementação do worker
│   └── Dockerfile.dev    # Definição do container de desenvolvimento do worker
├── k8s/                # Manifests Kubernetes (Deployments, Services, ConfigMaps)
│   ├── api-config.yaml   # ConfigMap da API
│   ├── api.yaml        # Deployment e Service da API
│   ├── worker-config.yaml # ConfigMap do Worker
│   └── worker.yaml     # Deployment e Service do Worker
├── docker-compose.dev.yml # Definição dos serviços de dependência (MySQL, Memcached, RabbitMQ)
├── Makefile            # Comandos úteis para build e desenvolvimento
├── Tiltfile            # Configuração do Tilt.dev para orquestração
└── README.md           # Documentação do projeto
```

## Pré-requisitos

Para rodar este projeto localmente, você precisa ter instalado:

- Go 1.23.0 ou superior
- Docker
- Kubernetes (local ou remoto) - **Recomendado Docker Desktop com Kubernetes ativado**
- Tilt.dev CLI

## Configuração do Kubernetes Local

A forma mais simples de configurar um cluster Kubernetes local para desenvolvimento é usando **Docker Desktop com Kubernetes ativado**. Siga as instruções oficiais do Docker Desktop para ativar o Kubernetes nas configurações.

Alternativas como MicroK8s ou Minikube também funcionam, mas a configuração pode variar. Este README foca na integração com Docker Desktop + Tilt.

## Orquestração com Tilt.dev e Docker Compose

O Tilt.dev é a ferramenta central deste ambiente de desenvolvimento. Ele é configurado através do arquivo `Tiltfile` na raiz do projeto.

O `Tiltfile` define:

- Quais serviços contêinerizados compõem a aplicação (serviços Go no Kubernetes e dependências no Docker Compose).
- Como construir as imagens dos serviços Go para desenvolvimento (`Dockerfile.dev`).
- Como implantar os serviços Go no Kubernetes (usando os manifests em `k8s/`).
- Como gerenciar as dependências definidas no `docker-compose.dev.yml`.
- Como realizar o **Live Update** do código Go.

### Como Funciona o Live Update do Código Go

Para os serviços API e Worker, o Tilt utiliza uma estratégia de Live Update otimizada para Go:

1.  O `Tiltfile` monitora mudanças nos arquivos Go dentro dos diretórios `./api` e `./worker`.
2.  Quando uma mudança é detectada, o Tilt sincroniza (`sync`) os arquivos alterados para o diretório de trabalho (`/app`) dentro do contêiner em execução.
3.  Em seguida, um comando (`run`) é executado DENTRO do contêiner para:
    *   Atualizar as dependências Go (`go mod tidy`).
    *   **Recompilar o binário Go** (`go build -o /app/[nome_do_servico] .`).
4.  Finalmente, o `docker_build_with_restart` (configurado no Tiltfile) reinicia o processo principal do contêiner (que executa o binário recém-compilado `/app/[nome_do_servico]`).

Este fluxo permite atualizações de código muito rápidas sem a necessidade de reconstruir a imagem Docker completa a cada mudança.

### Gerenciamento de Dependências com Docker Compose

As dependências externas como banco de dados, cache e fila de mensagens são definidas no arquivo `docker-compose.dev.yml`. O Tilt é configurado para ler este arquivo e gerenciar o ciclo de vida desses serviços Docker Compose (`docker_compose('./docker-compose.dev.yml')`).

Você pode ver os serviços do Docker Compose (mysql, memcached, rabbitmq) no painel do Tilt, gerenciá-los (iniciar/parar) e visualizar seus logs.

## Iniciando o Ambiente de Desenvolvimento

1.  Certifique-se de que o Docker Desktop (com Kubernetes ativado) ou seu ambiente Kubernetes local escolhido esteja em execução.
2.  Abra um terminal na raiz do projeto (`tilt-go`).
3.  Execute o comando:
    ```bash
    make tilt-up
    ```

O Tilt irá ler o `Tiltfile`, iniciar os serviços Docker Compose, construir as imagens Go (apenas na primeira vez ou quando o Dockerfile.dev mudar), implantar os serviços Go no Kubernetes e abrir a UI do Tilt no seu navegador (geralmente em `http://localhost:10350/`).

## Acessando a Aplicação e Dependências

Com o Tilt rodando:

-   **API Server**: Acesível em `http://localhost:8080`.
-   **Worker**: Não possui interface web principal, mas seus logs são visíveis no Tilt.
-   **Logs da Aplicação (API e Worker)**: Visíveis diretamente no terminal onde você rodou `make tilt-up` ou na UI do Tilt.

### Endpoints de Saúde (Probes)

Os serviços API e Worker expõem endpoints de saúde utilizados pelo Kubernetes:

-   `/livez`: Liveness Probe - Verifica se o processo está rodando.
-   `/readyz`: Readiness Probe - Verifica se o serviço está pronto para receber tráfego (pode incluir checagens de dependências).
-   `/healthz`: Startup Probe - Usado durante a inicialização para indicar quando a aplicação está pronta para responder aos outros probes.

Você pode testar esses endpoints diretamente (ex: `http://localhost:8080/livez`).

### Acesso às Dependências

Os serviços definidos no `docker-compose.dev.yml` são acessíveis no seu host local pelas portas mapeadas:

-   **MySQL**: `localhost:3306`
    -   Usuário padrão: `root`
    -   Senha padrão: `root`
    -   Database padrão: `app`
-   **Memcached**: `localhost:11211`
-   **RabbitMQ**: `localhost:5672`
    -   Management UI: `http://localhost:15672`
    -   Usuário padrão: `guest`
    -   Senha padrão: `guest`

## Características da Aplicação

-   Logs estruturados em formato JSON com timestamp em UTC-3.
-   Graceful Shutdown implementado na API e Worker.
-   Scheduler simples no Worker para executar uma tarefa agendada.

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

## Comandos Úteis

-   `make tilt-up`: Inicia o ambiente de desenvolvimento completo com Tilt (serviços Go + dependências Docker Compose).
-   `make tilt-down`: Interrompe os serviços gerenciados pelo Tilt.
-   `make logs-services`: Exibe os logs dos serviços Docker Compose (MySQL, Memcached, RabbitMQ).
-   `make restart-services`: Reinicia os serviços Docker Compose.
-   `make clean`: Limpa o ambiente de build e containers.

## Autor

**Renato Magalhães**

## Repositório

Este projeto está hospedado em: [https://github.com/renatomagalhaes/tilt-go](https://github.com/renatomagalhaes/tilt-go)

## Licença