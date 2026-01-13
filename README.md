# Jaiu

Sistema de gestao de academia (MVP) com Go + templates HTML + HTMX + Tailwind + chi.

## Stack

- Go 1.22
- `html/template`
- HTMX para interacoes dinamicas
- Tailwind CSS (CDN no MVP, com build opcional)
- chi para roteamento HTTP

## Estrutura

```
cmd/server           # entrada da aplicacao
internal/app         # bootstrap da aplicacao
internal/http        # rotas e handlers
internal/view        # renderer de templates
internal/domain      # entidades e enums de dominio
web/templates        # layouts e paginas
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

## Proximos passos sugeridos

- Persistencia (SQLite ou Postgres) + migrations
- CRUDs completos com validacoes
- Autenticacao (admin/operador) e controle de acesso
- Relatorios com filtros por periodo
