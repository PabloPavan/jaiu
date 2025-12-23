# jaiu_frontend

Base do MVP usando Next.js, TypeScript e TailwindCSS. O backend será desenvolvido
em um repositório separado, portanto este projeto contém apenas o necessário
para o frontend.

## Como rodar

```bash
npm install
npm run dev
```

## Variáveis de ambiente

Copie o `.env.example` para `.env` e ajuste o backend:

```bash
cp .env.example .env
```

`NEXT_PUBLIC_API_BASE_URL` define a URL base do backend REST (mantido em outro
repositório).
