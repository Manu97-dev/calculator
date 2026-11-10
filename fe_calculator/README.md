# Calculator frontend

A React and TypeScript starter with Vite, React Router, TanStack Query and Sass Modules. It includes a home screen, a fallback route and a shared HTTP client.

## Local setup

Use Node.js 22 LTS (22.16 or newer in the 22.x line) and npm.

```sh
npm ci
npm run dev
```

Open [http://localhost:5173](http://localhost:5173). Vite requires that port to be free. It forwards requests under `/api` to the Go backend at `http://localhost:3000` with the prefix intact. To check the proxy, start the Go API and run `curl -i http://localhost:5173/api/health`.

The HTTP client defaults to `/api`. To override the base, copy `.env.example` to `.env.local` and set `VITE_API_BASE_URL`. Vite embeds `VITE_*` values in the public bundle. Keep secrets on the server.

## Commands

| Command | Purpose |
| --- | --- |
| `npm run dev` | Start Vite with hot reload |
| `npm run typecheck` | Check TypeScript projects |
| `npm run lint` | Run Oxlint |
| `npm test` | Run Vitest and Testing Library tests |
| `npm run build` | Type-check and build into `dist/` |
| `npm run preview` | Preview the built frontend |

TanStack Query treats successful query data as fresh for 30 seconds. Queries and mutations disable automatic retries by default. A feature can override the defaults for individual operations. Validate the HTTP client’s `unknown` response against a feature’s contract before using it as typed data.

## Deployment

Publish the contents of `dist/` and configure the host to serve `index.html` for client-side routes. The Vite proxy runs only in development. In production, configure a same-origin reverse proxy for `/api` or set `VITE_API_BASE_URL` when building and configure the backend to allow the frontend origin through CORS.
