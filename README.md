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

## Database Migrations

O MySQL é inicializado usando scripts SQL localizados no diretório `migrations/`. O `docker-compose.dev.yml` monta este diretório no contêiner MySQL, e o `docker-entrypoint-initdb.d` do MySQL executa automaticamente os arquivos `.sql` encontrados em ordem alfabética.

Recomenda-se um padrão de nomeação sequencial (ex: `001_create_table_x.sql`, `002_insert_initial_data_y.sql`) para garantir a ordem correta de execução das migrations.

## Iniciando o Ambiente

1.  Confirme que seu ambiente Kubernetes local está rodando.
2.  Na raiz do projeto, execute:
    ```bash
    make tilt-up
    ```

O Tilt iniciará os serviços Docker Compose (executando as migrations do MySQL), construirá e implantará os serviços Go no Kubernetes e abrirá a UI do Tilt (`http://localhost:10351/`).

## Acesso e Monitoramento

Com o ambiente ativo via Tilt:

*   **API Server**: Acessível em `http://localhost:8080`.
*   **Worker**: Logs monitorados via Tilt.
*   **Logs da Aplicação (API/Worker)**: Consolidados na UI do Tilt ou no terminal.
*   **Logs das Dependências (DC)**: Use `docker logs [nome_do_container]` ou `make logs-services`.

### Endpoints de Saúde (Kubernetes Probes)

Os serviços API e Worker expõem endpoints específicos para probes de saúde do Kubernetes:

*   `/livez`: Liveness Probe (verifica processo ativo).
*   `/readyz`: Readiness Probe (verifica prontidão para tráfego e **conexão com o banco de dados**).
*   `/healthz`: Startup Probe (verifica inicialização completa).

Os endpoints `/readyz` e `/healthz` agora incluem uma verificação da conexão com o banco de dados MySQL, garantindo que o serviço só reporte prontidão ou startup completo se a comunicação com o DB estiver funcional. Estes endpoints são projetados para serem leves e rápidos, seguindo as práticas do Kubernetes.

### Rota de Frases Aleatórias

A API agora possui uma nova rota:

*   `GET /quotes/random`: Retorna uma frase aleatória do banco de dados.

### Logging Padronizado (API - /quotes/random)

A rota `/quotes/random` na API utiliza logging estruturado (JSON) via `go.uber.org/zap` para observabilidade aprimorada. As logs incluem:

*   Timestamp, nível, serviço, endpoint e método.
*   Endereço remoto da requisição.
*   Duração da chamada ao banco de dados.
*   Em caso de sucesso, o ID da frase (`quote_id`) e o status code (200).
*   Em caso de erro, detalhes do erro e status code (500).

Este padrão de logging facilita a análise e monitoramento do comportamento da API.

### Controle de Nível de Logging

Tanto a API quanto o Worker utilizam logging estruturado com níveis controlados por uma variável de ambiente:

*   Para habilitar logs de nível `DEBUG` (mais verbosos, úteis para análise detalhada e desenvolvimento), defina a variável de ambiente `DEBUG_LOGGING=true`.
*   Por padrão, sem esta variável definida, os serviços logarão no nível `INFO` (resumos de operações importantes e erros), ideal para ambientes de produção ou com menos verbosidade necessária.

Logs de nível `ERROR` sempre serão exibidos, independentemente da configuração de `DEBUG_LOGGING`.

### Acesso Direto às Dependências (Host)

Os serviços Docker Compose são mapeados para `localhost`:

*   **MySQL**: `localhost:3306` (user: `root`, pass: `root`, db: `app`)
*   **Memcached**: `localhost:11211`
*   **RabbitMQ**: `localhost:5672` (Management UI: `http://localhost:15672`, user: `guest`, pass: `guest`)

## Características Implementadas

*   Logs estruturados em JSON.
*   Graceful Shutdown para API e Worker.
*   Scheduler de exemplo no Worker.
*   Database Migrations via scripts SQL.
*   Rota de frases aleatórias (`/quotes/random`).
*   Health checks com verificação de conexão ao DB.
*   Logging padronizado para rota `/quotes/random`.

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