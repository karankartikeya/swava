# SwarmPay frontend

Four-page site for the SwarmPay hackathon build: Home, How It Works, Live Demo,
Findings. React + Vite + TypeScript + Tailwind v4, calling `cmd/server` (Go)
over HTTP — no backend logic reimplemented here.

## Develop

```bash
npm install
npm run dev
```

Vite's dev server proxies `/api/*` to `http://localhost:8081` (see
`vite.config.ts`) — start `cmd/server` (and `reputation-api`, which it depends
on) before running the Live Demo page:

```bash
# from the repo root, in separate terminals
cd ../../swarmpay-backend && go run ./cmd/reputation-api
cd .. && go run ./cmd/server
```

## Build

```bash
npm run build
```

Outputs a static `dist/` — deployable to any static host (Vercel, Netlify,
S3+CloudFront, etc.). Set `VITE_API_BASE` at build time (see `.env.example`)
if the API isn't reachable at the same origin under `/api`.

## Fonts

Fraunces (display), Inter (body), and JetBrains Mono (data/code) are
self-hosted under `src/fonts/` and loaded via `@font-face` in `src/index.css`
— no runtime CDN request.
