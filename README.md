# Ambiente de Desenvolvimento Local: Go (Kubernetes) + Dependências (Docker Compose) com Tilt.dev

Este projeto configura um ambiente de desenvolvimento local para serviços Go orquestrados via Tilt.dev, rodando no Kubernetes, com dependências externas (banco de dados, cache, fila) gerenciadas por Docker Compose.

O foco desta branch é otimizar o ciclo de desenvolvimento local com hot-reload eficiente para código Go e gerenciamento integrado das dependências.

## Componentes

*   **API Server**: Serviço Go HTTP.
*   **Worker**: Serviço Go em background com scheduler.
*   **Dependências**: MySQL, Memcached, RabbitMQ via Docker Compose.
*   **Tilt.dev**: Orquestração do ambiente híbrido e ciclo de vida dos serviços.

## Pré-requisitos

*   Go (>= 1.23.0)
*   Docker
*   Kubernetes (local ou remoto) - **Docker Desktop com Kubernetes ativado é o cenário padrão aqui**.
*   Tilt.dev CLI

## Configuração do Kubernetes Local

Para um ambiente local integrado com Docker Compose, a configuração mais comum é o **Docker Desktop com o feature de Kubernetes ativado**. Garanta que seu Docker Desktop esteja configurado para expor a API do Kubernetes.

Outras distribuições Kubernetes locais (MicroK8s, Minikube) também podem ser usadas, ajustando a configuração do seu cliente `kubectl`.

## Orquestração do Ambiente com Tilt.dev

O `Tiltfile` na raiz do projeto centraliza a configuração do ambiente de desenvolvimento.

Ele define a interação entre:

*   **Serviços Go (API/Worker)**: Construção via `Dockerfile.dev` e deploy no Kubernetes (`k8s/*.yaml`).
*   **Dependências**: Ciclo de vida gerenciado através do `docker-compose.dev.yml`.

### Live Update Otimizado para Go

O `Tiltfile` configura um pipeline de Live Update para os serviços Go (API e Worker) que garante feedback rápido durante o desenvolvimento:

1.  Monitora mudanças nos diretórios `./api` e `./worker`.
2.  Sincroniza (`sync`) arquivos modificados para `/app` no contêiner em execução.
3.  Executa um comando (`run`) dentro do contêiner para recompilar o binário Go (`go mod tidy && go build -o /app/[nome_do_servico] .`).
4.  O `docker_build_with_restart` reinicia o processo do contêiner, executando o binário recém-compilado (`/app/[nome_do_servico]`).

Este processo foca em atualizações incrementais no contêiner sem builds de imagem completos, reduzindo significativamente o tempo de feedback.

### Gerenciamento de Dependências

As dependências essenciais para o desenvolvimento local (MySQL, Memcached, RabbitMQ) são definidas no `docker-compose.dev.yml`. O `Tiltfile` integra esse arquivo (`docker_compose('./docker-compose.dev.yml')`), permitindo que o Tilt gerencie o startup e shutdown desses serviços juntamente com os serviços da aplicação. Labels são aplicadas via `dc_resource` para organização na UI do Tilt.

## Iniciando o Ambiente

1.  Confirme que seu ambiente Kubernetes local está rodando.
2.  Na raiz do projeto, execute:
    ```bash
    make tilt-up
    ```

O Tilt iniciará os serviços Docker Compose, construirá e implantará os serviços Go no Kubernetes e abrirá a UI do Tilt (`http://localhost:10350/`).

## Acesso e Monitoramento

Com o ambiente ativo via Tilt:

*   **API Server**: Acessível em `http://localhost:8080`.
*   **Worker**: Logs monitorados via Tilt.
*   **Logs da Aplicação (API/Worker)**: Consolidados na UI do Tilt ou no terminal.
*   **Logs das Dependências (DC)**: Use `docker logs [nome_do_container]` ou `make logs-services`.

### Endpoints de Saúde (Kubernetes Probes)

Os serviços API e Worker expõem endpoints específicos para probes de saúde do Kubernetes:

*   `/livez`: Liveness Probe (verifica processo ativo).
*   `/readyz`: Readiness Probe (verifica prontidão para tráfego).
*   `/healthz`: Startup Probe (verifica inicialização completa).

Estes endpoints são projetados para serem leves e rápidos, seguindo as práticas do Kubernetes.

### Acesso Direto às Dependências (Host)

Os serviços Docker Compose são mapeados para `localhost`:

*   **MySQL**: `localhost:3306` (user: `root`, pass: `root`, db: `app`)
*   **Memcached**: `localhost:11211`
*   **RabbitMQ**: `localhost:5672` (Management UI: `http://localhost:15672`, user: `guest`, pass: `guest`)

## Características Implementadas

*   Logs estruturados em JSON.
*   Graceful Shutdown para API e Worker.
*   Scheduler de exemplo no Worker.

## Detalhes Técnicos Adicionais

As seções a seguir aprofundam em aspectos técnicos da configuração, relevantes para a implantação em ambientes orquestrados.

### Probes do Kubernetes

(Manter a seção existente sobre Liveness, Readiness, Startup Probes, Exemplo de Configuração e Explicação dos Parâmetros. É detalhada e técnica, adequada para Devs Plenos).

### Graceful Shutdown

(Manter a seção existente sobre a implementação e testes de Graceful Shutdown).

## Comandos Úteis

*   `make tilt-up`: Inicia o ambiente Tilt.
*   `make tilt-down`: Interrompe o ambiente Tilt.
*   `make logs-services`: Exibe logs agregados dos serviços Docker Compose.
*   `make restart-services`: Reinicia serviços Docker Compose.
*   `make clean`: Limpa artefatos de build e containers.

## Autor

**Renato Magalhães**

## Repositório

[https://github.com/renatomagalhaes/tilt-go](https://github.com/renatomagalhaes/tilt-go)

## Licença