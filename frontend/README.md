# Swava frontend

Four-page site for the Swava hackathon build: Home, How It Works, Live Demo,
Findings. React + Vite + TypeScript + Tailwind v4, calling `cmd/server` (Go)
over HTTP — no backend logic reimplemented here. See the [root README](../README.md)
for the full system architecture.

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

## Design system

Bricolage Grotesque (display headlines), Inter (body/UI), and Roboto Mono
(data/labels) are self-hosted under `src/fonts/` and loaded via `@font-face`
in `src/index.css` — no runtime CDN request. Full token system (palette,
type scale, spacing, component shapes) documented in `design.md`.

## API endpoints used

All calls go through the single `BASE` constant in `src/lib/api.ts`
(`import.meta.env.VITE_API_BASE ?? "/api"`):

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/wallets` | GET | The 3 demo agent identities shown on Live Demo |
| `/api/decide` | POST | Run the agent decision loop for a task + budget |
| `/api/trust-gate` | POST | Score + policy check for a wallet + amount, no Prava call |
| `/api/purchase` | POST | Full flow: trust gate, then (if not blocked) a real Prava session |
| `/api/agent-profile` | GET | Role, trust score, procurement policy, and real tx history for one wallet |
