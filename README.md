# Jaiu

Sistema de gestao de academia (MVP) com Go + templ + HTMX + Tailwind + chi.

## Stack

- Go 1.22
- templ (componentes de UI)
- HTMX para interacoes dinamicas
- Tailwind CSS (CDN no MVP, com build opcional)
- chi para roteamento HTTP
- Postgres
- sqlc + migrations

## Estrutura

```
cmd/server           # entrada da aplicacao
internal/app         # bootstrap da aplicacao
internal/adapter     # adaptadores (db, etc)
internal/http        # rotas e handlers
internal/ports       # portas (interfaces)
internal/service     # casos de uso
internal/view        # componentes (templ)
internal/domain      # entidades e enums de dominio
db/migrations        # migrations SQL
db/queries           # queries sqlc
db/schema.sql        # schema para sqlc
sqlc.yaml            # config do sqlc
web/static           # css/js estatico
web/tailwind         # config do Tailwind (build opcional)
```

## Modulos cobertos

- Alunos
- Planos
- Assinaturas
- Pagamentos
- Relatorios
- Usuarios e autenticacao (tela inicial)

## Rodando localmente

```bash
export DATABASE_URL="postgres://jaiu:secret@localhost:5432/jaiu?sslmode=disable"
go run ./cmd/server
```

Acesse: `http://localhost:8080`

## Docker

```bash
docker compose up --build
```

## Tailwind

No MVP esta usando o CDN do Tailwind direto no template base. Quando quiser compilar:

```bash
./scripts/tailwind.sh
```

Isso gera `web/static/css/app.css`. Depois, remova o script do CDN no template base.

## templ

Gere os componentes a partir dos arquivos `.templ`:

```bash
templ generate
```

Se o binario `templ` nao estiver instalado:

```bash
go install github.com/a-h/templ/cmd/templ@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

Atalho:

```bash
./scripts/templ.sh
```

Se preferir rodar via Docker:

```bash
docker run --rm -v "$PWD":/workspace -w /workspace ghcr.io/a-h/templ:latest generate
```

## Migrations

```bash
docker compose --profile migrate run --rm migrate
```

## sqlc

```bash
./scripts/sqlc.sh
```

Se o binario `sqlc` nao estiver instalado:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

Ou rode via Docker:

```bash
docker run --rm -v "$PWD":/workspace -w /workspace sqlc/sqlc:1.26.0 generate
```

## Seed de usuario

```bash
export DATABASE_URL="postgres://jaiu:secret@localhost:5432/jaiu?sslmode=disable"
go run ./cmd/seed-user -email admin@academia.com -password 123456 -name "Admin" -role admin
```

## Proximos passos sugeridos

- CRUDs completos com validacoes
- Autenticacao (admin/operador) e controle de acesso
- Relatorios com filtros por periodo
