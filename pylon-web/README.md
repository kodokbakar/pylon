# Pylon Web

React frontend for Pylon.

## Stack

- Bun
- Vite
- React
- TypeScript
- Tailwind CSS v4
- ESLint
- Prettier

## Tailwind

This project uses Tailwind CSS v4 with the official Vite plugin.

Tailwind is wired through:

- `@tailwindcss/vite` in `vite.config.ts`
- `@import 'tailwindcss';` in `src/index.css`

There is no `postcss.config.js` or `tailwind.config.js` because this setup intentionally uses the Tailwind v4 Vite flow.

## Folder Structure

```text
src/
  api/
  components/
  hooks/
  pages/
  types/
  utils/
```

## Commands

```bash
bun install
bun run dev
bun run lint
bun run build
bun run format:check
```

## Environment

```fish
cp .env.example .env
```

```text
VITE_API_BASE_URL=http://localhost:8080
```

## Verification

Last verified with:

```text
bun: 1.3.14
node: v22.22.3
vite: 8.1.0
```

Checks:

```bash
bun install
bun run lint
bun run build
bun run format:check
bun run dev -- --host 127.0.0.1 --port 5173
```

Expected dev server:

```text
http://127.0.0.1:5173/
```
