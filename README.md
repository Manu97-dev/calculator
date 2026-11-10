# Full-stack calculator

A responsive React and TypeScript calculator backed by a Go REST API. It supports addition, subtraction, multiplication, division, and chained expressions with standard operator precedence.

**Try it:** Enter `2 + 3 × 4` and press **=** to get **14**. Then press **− 3 =** to continue from that result and get **11**.

## Run with Docker

With Docker running, execute these commands from the repository root:

```sh
docker build -t sezzle-calculator .
docker run --rm -p 3000:3000 sezzle-calculator
```

Open [the calculator on port 3000](http://localhost:3000). The container serves the built React app and the Go API from the same origin.

## Run locally

Use **Go 1.26 or newer** and **Node.js 22.16 or newer in the 22.x line**, with npm. Start each service in a separate terminal from the repository root.

| Service | Command |
| --- | --- |
| Go API on port 3000 | `cd be_calculator && go run .` |
| Vite on port 5173 | `cd fe_calculator && npm ci && npm run dev` |

Open [the calculator on port 5173](http://localhost:5173). Vite proxies `/api` requests to Go.

## API examples

All examples target port 3000 and work with either setup. The UI sends the full expression to `POST /api/operations/evaluate`.

```sh
curl -sS http://localhost:3000/api/operations/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"2 + 3 × 4"}'
# {"result":14}
```

The four individual routes under `/api/operations/` are `addition`, `subtraction`, `multiplication`, and `division`. Each accepts numeric `left` and `right` fields.

```sh
curl -sS http://localhost:3000/api/operations/division \
  -H 'Content-Type: application/json' \
  -d '{"left":5,"right":2}'
# {"result":2.5}
```

An invalid expression returns `400 INVALID_EXPRESSION`. Dividing by zero returns HTTP `422` and the following JSON:

```json
{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero."}}
```

See the [backend README](be_calculator/README.md) for the full error contract and more examples.

## Tests and coverage

After installing frontend dependencies, run these commands from the repository root:

```sh
(cd fe_calculator && npm test)
(cd be_calculator && go test ./...)
```

To reproduce the coverage summary and generate detailed HTML reports:

```sh
(cd fe_calculator && npm run test:coverage)
sh be_calculator/coverage.sh
```

| Layer | Statements | Branches | Functions | Lines |
| --- | ---: | ---: | ---: | ---: |
| React and TypeScript (Vitest V8) | 100% | 100% | 100% | 100% |
| Go (unit and startup tests combined) | 100% | — | — | — |

Open `fe_calculator/coverage/index.html` or `be_calculator/coverage/index.html` after running coverage. The Go script prints the unit-test summary first, then the combined total from unit tests and startup checks against the instrumented binary. GitHub Actions runs coverage for both layers.

## Design decisions

- **Go owns arithmetic and validation.** React manages keypad input and uses one TanStack Query mutation to submit an expression, show the result, or display a server error while keeping the input editable.
- **One request evaluates the entire expression.** A small recursive-descent parser evaluates multiplication and division before addition and subtraction. Operators at the same precedence run left to right. The individual endpoints let clients submit two structured operands directly and share the same arithmetic functions as the parser.
- **Numbers use `float64`.** This supports fractional division and follows ordinary binary floating-point rounding. Expressions are limited to 1,024 bytes and 32 levels of parentheses. Parentheses are accepted by the API.
- **Scope covers the four required operations.** Exponentiation, square root, and percentage are outside the current scope.

See the [frontend README](fe_calculator/README.md) for the Vite proxy and deployment configuration.
