# Calculator starter

This repository contains a React/Vite frontend in [fe_calculator](fe_calculator/README.md) and a Go REST API in [be_calculator](be_calculator/README.md).

## Run both with Docker

From the repository root, with Docker running:

```sh
docker build -t sezzle-calculator .
docker run --rm -p 3000:3000 sezzle-calculator
```

Open [the app](http://localhost:3000) or check the API from another terminal.

```sh
curl -i http://localhost:3000/api/health
```

The Dockerfile builds the frontend with npm and compiles the API with Go. The final image contains the Go executable and the built frontend. Go serves both on port 3000, so the browser can use same-origin `/api` requests. Browser navigation to client-side routes loads the React app. The API and missing Vite assets retain their own HTTP responses.

## Develop locally

Use Go 1.26+ and Node.js 22.16+ in the Node 22 line. Run these in separate terminals from the repository root.

| Service | Command |
| --- | --- |
| API | `cd be_calculator && go run .` |
| Frontend | `cd fe_calculator && npm ci && npm run dev` |

Open [the Vite dev server](http://localhost:5173). It proxies `/api` to Go on port 3000.

## Test coverage

Run the commands below from the repository root. They print coverage in the terminal and generate detailed HTML reports.

```sh
(cd fe_calculator && npm ci && npm run test:coverage)

sh be_calculator/coverage.sh
```

Open `fe_calculator/coverage/index.html` for React and `be_calculator/coverage/index.html` for Go. The generated reports are ignored by Git. CI prints coverage for both layers on pushes and pull requests.

### Coverage report

| Layer | Statements | Branches | Functions | Lines |
| --- | ---: | ---: | ---: | ---: |
| React and TypeScript (Vitest V8) | 100% | 100% | 100% | 100% |
| Go (unit and startup tests combined) | 100% | — | — | — |

The frontend suite runs 28 tests, including a check that mounts the production React entry point. The Go report combines the unit-test profile with coverage from an instrumented binary that verifies startup error handling. The Go unit-test summary appears first in the terminal, followed by the combined statement total used in the table and HTML report.
