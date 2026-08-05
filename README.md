# oracle-winthor-mocked-environment

Ambiente de banco Oracle voltado para simular parte do schema do Winthor localmente, com migrations em SQL puro executadas por um runner em Go.

O objetivo deste repositório e permitir que tabelas, relacionamentos, índices e seeds mínimas do contexto Winthor sejam recriados rapidamente em um Oracle local, sem depender do banco real do ERP.

A ideia é ter um ambiente simulado ao do ERP para testes, estudo etc.

## Contexto do projeto

Este projeto existe para montar um ambiente Oracle controlado, previsível e versionável para desenvolvimento local, testes de integração e experimentação com estruturas inspiradas no Winthor.

Na prática, ele centraliza:

- definição de tabelas e objetos SQL em arquivos versionados;
- bootstrap inicial do schema `WINTHOR`;
- execução sequencial das migrations;
- conexão com Oracle via Go usando `go-ora`;
- configuração do ambiente por `.env`.

## API mock do Winthor

O serviço HTTP mock fica disponível na porta `1522`; o Oracle continua usando a porta `1521`.

Para iniciar a API:

```bash
docker compose up --build -d api
```

Nesta primeira versão, os endpoints não exigem autenticação e não acessam o Oracle.

### Login

`POST /winthor/autenticacao/v1/login`

Gera um token aleatório a cada requisição:

```bash
curl -X POST http://localhost:1522/winthor/autenticacao/v1/login
```

Resposta:

```json
{"accessToken":"token-aleatorio"}
```

### Importar venda

`POST /winthor/venda/v0/importar-venda?ignoraProcessamento=false`

O corpo precisa ser um JSON válido. O parâmetro `ignoraProcessamento` é aceito, mas ignorado pelo mock.

```bash
curl -X POST \
  'http://localhost:1522/winthor/venda/v0/importar-venda?ignoraProcessamento=false' \
  -H 'Content-Type: application/json' \
  -d '{
    "numPedRca": 7001,
    "codCli": 1,
    "codUsur": 1,
    "items": []
  }'
```

Resposta:

```json
{
  "message": "Pedidos gerados com sucesso!",
  "data": {
    "numPedRca": 7001,
    "codCli": 1,
    "codUsur": 1,
    "items": []
  }
}
```

JSON inválido retorna `400`; métodos diferentes de `POST` retornam `405`.

Hoje, o fluxo foi simplificado para um modelo pragmático:

- o comando `up` lê os arquivos `.sql` em ordem de nome;
- o comando `seed` executa apenas os arquivos da pasta `seeds/`;
- o SQL de cada arquivo é executado diretamente no Oracle;
- a execução acontece sempre com o usuário admin configurado no `.env`;
- o usuário `WINTHOR` é tratado como schema alvo dos objetos, não como usuário de execução das migrations.

Essa decisão foi tomada porque, neste ambiente, o usuário `WINTHOR` não possui privilégios suficientes para executar o fluxo completo de migrations com consistência.

## Como o runner funciona hoje

O runner de migrations está em `internal/migration/runner.go`.

Comportamento atual do `up`:

- carrega todos os arquivos de `migrations/`;
- separa migrations de bootstrap das demais;
- abre uma conexão Oracle com o usuário admin;
- executa primeiro os arquivos de bootstrap;
- executa depois as migrations de schema;
- executa por fim as seeds de `seeds/`.

Comportamento atual do `down`:

- ainda depende de uma tabela de controle `WINTHOR.MIGRATIONS` para identificar a última migration aplicada;
- abre a conexão com o usuário admin;
- executa o `down.sql` correspondente à última versão registrada.

Comportamento atual do `seed`:

- carrega apenas os arquivos da pasta `seeds/`;
- abre a conexão Oracle com o usuário admin;
- executa as seeds em ordem lexical.

Importante:

- o `up` atualmente não registra versões em `WINTHOR.MIGRATIONS`;
- isso significa que o `down` só funciona corretamente se essa tabela estiver sendo alimentada por algum fluxo complementar ou se você adaptar esse comportamento depois;
- no estado atual, o foco principal do projeto está em subir estrutura via `up`.

## Estrutura do repositório

```text
.
├── cmd/
│   ├── api/
│   ├── migrate/
│   └── testconn/
├── internal/
│   ├── config/
│   ├── db/
│   └── migration/
├── migrations/
├── seeds/
├── assets/
├── .env.example
├── go.mod
└── README.md
```

## O que cada pasta faz

### `cmd/migrate`

CLI principal para trabalhar com migrations.

Comandos disponíveis:

```bash
docker compose run --rm go run ./cmd/migrate create <nome>
docker compose run --rm go run ./cmd/migrate create-seed <nome>
docker compose run --rm go run ./cmd/migrate up
docker compose run --rm go run ./cmd/migrate seed
docker compose run --rm go run ./cmd/migrate down
```

### `cmd/testconn`

Utilitário simples para testar conectividade com Oracle. Não é o fluxo principal, mas é útil quando você quer validar credenciais e reachability antes de subir migrations.

### `internal/config`

Carrega variáveis do `.env` e entrega a configuração usada pelo restante da aplicação.

### `internal/db`

Centraliza a abertura das conexões Oracle.

Embora ainda exista suporte a `OpenApp`, o fluxo operacional atual das migrations usa o admin.

### `internal/migration`

Contém:

- criação de arquivos de migration e seed;
- parser para SQL Oracle e blocos PL/SQL;
- execução das migrations;
- rollback baseado na tabela `WINTHOR.MIGRATIONS`.

### `migrations`

Arquivos SQL versionados que definem bootstrap, tabelas, foreign keys e índices.

Os nomes seguem o padrão:

```text
YYYYMMDDHHMMSS_nome_da_migration.up.sql
YYYYMMDDHHMMSS_nome_da_migration.down.sql
```

### `seeds`

Arquivos SQL de carga inicial ou complementar.

### `assets`

Material de apoio e documentação externa relacionada ao contexto Winthor. Não faz parte do runtime principal do runner.

## Requisitos

Antes de subir o ambiente, garanta:

- `Docker` e `Docker Compose` plugin instalados;
- porta `1521` livre na máquina;
- memória suficiente para rodar Oracle Free localmente;
- usuário com permissão para executar containers Docker.

Versões recomendadas:

- Docker Engine recente;
- imagem Oracle Free compatível com `FREEPDB1`.

## Variáveis de ambiente

Crie um `.env` a partir do exemplo:

```bash
cp .env.example .env
```

Exemplo de preenchimento:

```env
APP_ENV=local

ORACLE_ADMIN_USER=system
ORACLE_ADMIN_PASSWORD=oracle
ORACLE_HOST=oracle
ORACLE_PORT=1521
ORACLE_SERVICE=FREEPDB1

ORACLE_APP_USER=WINTHOR
ORACLE_APP_PASSWORD=Winthor123
```

Se você for executar o Go fora do `docker compose`, ajuste `ORACLE_HOST` para `localhost`.

### Significado das variáveis

- `ORACLE_ADMIN_USER`: usuário que executa as migrations.
- `ORACLE_ADMIN_PASSWORD`: senha do usuário admin.
- `ORACLE_HOST`: host do banco Oracle.
- `ORACLE_PORT`: porta do listener Oracle.
- `ORACLE_SERVICE`: service name, normalmente `FREEPDB1`.
- `ORACLE_APP_USER`: schema lógico da aplicação, hoje usado principalmente como referência ao schema `WINTHOR`.
- `ORACLE_APP_PASSWORD`: senha do usuário `WINTHOR`.

## Passo a passo para subir o ambiente com Docker Compose

1. Clonar o repositório

```bash
git clone https://github.com/isaias-systock/oracle-winthor-mocked-environment.git
cd oracle-winthor-mocked-environment
```

2. Criar e ajustar o `.env`

```bash
cp .env.example .env
```

Por padrão, o exemplo já vem pronto para o fluxo via `docker compose`, usando `ORACLE_HOST=oracle`.

Se você preferir rodar o binário Go direto na máquina, troque esse valor para `localhost`.

3. Subir o Oracle Free e o runner Go

O repositório agora inclui:

- `Dockerfile`: imagem de desenvolvimento com o toolchain Go;
- `docker-compose.yml`: stack com `oracle` e `go`.

Suba o banco:

```bash
docker compose up -d oracle
```

Se quiser forçar o build da imagem do runner Go antes do primeiro uso:

```bash
docker compose build go
```

Observações:

- a imagem é distribuída pela Oracle Container Registry;
- você pode precisar autenticar antes com `docker login container-registry.oracle.com`;
- o primeiro startup pode demorar alguns minutos;
- o PDB esperado neste projeto é `FREEPDB1`.

Para acompanhar a inicialização:

```bash
docker compose logs -f oracle
```

Espere até o banco aceitar conexões antes de seguir.

### 4. Validar acesso ao banco

Confirme que o listener está ativo e que o usuário admin do `.env` bate com o container.

Neste projeto, o caso esperado é:

- usuário admin: `system`
- senha admin: a mesma definida em `ORACLE_PWD`
- service: `FREEPDB1`

### 5. Baixar dependências Go no container

```bash
docker compose run --rm go mod download
```

Isso baixa e reaproveita dependências usando os volumes `go-mod-cache` e `go-build-cache`, sem exigir Go instalado na máquina.

### 6. Testar a conectividade

Você pode validar a conexão com o banco executando:

```bash
docker compose run --rm go run ./cmd/testconn
```

Se o banco estiver acessível, o runner vai abrir a conexão admin e começar a aplicar as migrations.

Se preferir validar com um utilitário separado, revise `cmd/testconn`, mas o fluxo principal hoje está concentrado no `cmd/migrate`.

### 7. Aplicar as migrations

```bash
docker compose run --rm go run ./cmd/migrate up
```

O runner executa os arquivos em ordem lexical, que neste projeto coincide com a ordem cronológica do timestamp.

Fluxo esperado:

1. bootstrap do schema `WINTHOR`;
2. criação das tabelas principais;
3. criação de chaves estrangeiras;
4. criação de índices;
5. execução das seeds, se existirem.

## Como criar novas migrations

Para gerar uma nova migration:

```bash
docker compose run --rm go run ./cmd/migrate create nome_da_migration
```

Isso cria:

```text
migrations/<timestamp>_nome_da_migration.up.sql
migrations/<timestamp>_nome_da_migration.down.sql
```

Para criar uma seed:

```bash
docker compose run --rm go run ./cmd/migrate create-seed dados_tabela_xpto
```

Para executar apenas as seeds:

```bash
docker compose run --rm go run ./cmd/migrate seed
```

## Convenções das migrations

Algumas regras implícitas do projeto:

- arquivos são executados em ordem alfabética;
- o prefixo com timestamp define a ordem;
- migrations de bootstrap são detectadas pelo nome, normalmente contendo `bootstrap`;
- o SQL é executado como Oracle SQL puro;

## Decisões importantes de arquitetura

### 1. Migrations executadas com usuário admin

Mesmo existindo `ORACLE_APP_USER`, o fluxo atual usa o usuário admin para executar o `up`.

Motivo:

- no ambiente atual, `WINTHOR` não tem privilégios suficientes para sustentar o fluxo de migration com confiabilidade.

### 2. `up` em modo stateless

O `up` foi simplificado para executar apenas o SQL puro dos arquivos.

Isso reduz atrito com privilégios do Oracle, principalmente em ambientes locais onde grants e ownership podem variar.

Consequência:

- reaplicar migrations pode falhar se o SQL não for idempotente;
- em alguns casos, o runner ignora objetos já existentes quando o Oracle retorna `ORA-00955`.

### 3. `down` ainda depende de tabela de controle

O rollback ainda foi mantido sobre `WINTHOR.MIGRATIONS`.

Se você quiser um fluxo totalmente coerente com o `up` stateless, o próximo ajuste natural é refatorar o `down` para não depender mais dessa tabela.

## Problemas comuns

### `ORA-01031: insufficient privileges`

Causa mais comum:

- tentativa de executar DDL com um usuário sem privilégio suficiente.

Abordagem adotada no projeto:

- usar sempre o usuário admin nas migrations.

### `ORA-01045: user does not have CREATE SESSION privilege`

Isso afeta o usuário `WINTHOR` quando ele tenta se conectar diretamente.

No fluxo atual, isso não deve bloquear o `up`, porque as migrations usam o admin.

### `dial tcp 127.0.0.1:1521: socket: operation not permitted`

Isso normalmente indica:

- banco não iniciado;
- porta não exposta;
- restrição do ambiente onde o comando foi executado;
- firewall ou permissão do Docker.

### `dial tcp 127.0.0.1:1521: connect: connection refused` ao usar `docker compose run`

Se isso acontecer no fluxo containerizado, quase sempre é sinal de `.env` com `ORACLE_HOST=localhost`.

Dentro do serviço `go`, `localhost` aponta para o próprio container, não para o Oracle.

Para o fluxo via Compose, use:

- `ORACLE_HOST=oracle`

### Objetos já existentes

Se você reaplicar o `up` em um banco parcialmente montado, algumas migrations podem falhar por objeto já existente.

O runner hoje tolera `ORA-00955` em alguns pontos do fluxo stateless, mas não substitui um desenho realmente idempotente de todas as migrations.

## Comandos úteis

Subir migrations:

```bash
docker compose run --rm go run ./cmd/migrate up
```

Gerar migration:

```bash
docker compose run --rm go run ./cmd/migrate create adicionar_tabela_x
```

Gerar seed:

```bash
docker compose run --rm go run ./cmd/migrate create-seed carga_inicial
```

Testar conectividade:

```bash
docker compose run --rm go run ./cmd/testconn
```

Rollback da última migration registrada:

```bash
docker compose run --rm go run ./cmd/migrate down
```

## Próximos ajustes recomendados

Se este repositório continuar evoluindo, os próximos passos mais úteis são:

- remover código morto do fluxo antigo de versionamento;
- alinhar `down` com o novo modelo stateless;
- documentar um SQL inicial opcional para criar `WINTHOR.MIGRATIONS`, se o rollback continuar dependendo dela;
- tornar migrations de bootstrap totalmente idempotentes.
- adicionar os demais indices, constraints etc faltantes.
- adicionar as demais tabelas, colunas e schemas faltantes.
