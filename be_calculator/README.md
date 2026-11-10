# Calculator API

A Go REST API for calculator operations. `GET /api/health` returns `200 OK` and `{"status":"ok"}`. The endpoint confirms the server can accept and answer HTTP requests.

## Run locally

Install a current patch release of [Go 1.26 or newer](https://go.dev/dl/). From this folder, start the API with `go run .`. In another terminal, try:

```sh
curl -i http://localhost:3000/api/health
```

The frontend's Vite proxy forwards `/api` to port `3000`. To run the API by itself on a different port, use `PORT=3001 go run .`. The server listens on all interfaces.

## Add two numbers

```sh
curl -i http://localhost:3000/api/operations/addition \
  -H 'Content-Type: application/json' \
  -d '{"left":2.5,"right":-1}'
```

`POST /api/operations/addition` returns `200 OK` with `{"result":1.5}`. Both fields must be finite JSON numbers. Go uses `float64`, the same binary precision as JavaScript's `Number`.

| HTTP status | Error code | When it occurs |
| --- | --- | --- |
| 400 | `INVALID_JSON` | The body does not contain exactly one JSON object. |
| 400 | `INVALID_OPERANDS` | An operand is missing, null, not a number, or too large in magnitude for `float64`. |
| 413 | `REQUEST_TOO_LARGE` | The request body exceeds 4 KiB. |
| 422 | `RESULT_OUT_OF_RANGE` | Adding two valid operands produces an infinite result. |

Errors from the handler use `{"error":{"code":"INVALID_OPERANDS","message":"Both left and right must be finite JSON numbers."}}`.

### How the addition handler works

1. **Route the request.** `newRouter` sends `/api/` requests to a separate `ServeMux`, similar to mounting an Express router with `app.use('/api', router)`. Go preserves the full request path in the nested router, so `registerOperations` uses `POST /api/operations/addition` as the route pattern. Keeping the API router separate ensures that a wrong HTTP method gets a `405` even when the SPA fallback is enabled. [Go HTTP package](https://pkg.go.dev/net/http#ServeMux)
2. **Read and validate JSON.** `http.MaxBytesReader` caps the body before the handler parses it, like the `limit` option on Express's `express.json()`. `json.Unmarshal` checks that the whole body is valid JSON. Decoding object fields as `json.RawMessage` keeps invalid JSON separate from valid JSON that contains bad operands. The handler then decodes each operand to `*float64` so it can distinguish zero from a missing or null value. Go's built-in decoder checks JSON type and numeric range without a validation dependency. [Go JSON package](https://pkg.go.dev/encoding/json#Unmarshal)
3. **Return the result.** The handler adds the operands, checks the sum with `math.IsInf`, and writes the status and a typed JSON response. The equivalent in Express would use `Number.isFinite` to check the sum and `res.status(...).json(...)` to send the response. An explicit range check makes overflow return a clear error before JSON encoding. [Go math package](https://pkg.go.dev/math#IsInf)

## From Node.js to Go, step by step

1. **Declare the module.** `go.mod` records the module name and minimum Go version. It serves a similar role to the dependency and runtime fields in `package.json`. The Go installation supplies all the packages this project imports, so setup needs no equivalent of `npm install`. [Go module reference](https://go.dev/ref/mod)
2. **Start at the entry point.** `package main` and `func main()` tell Go where execution starts, similar to running `node server.js`. `go run .` compiles the current package and immediately runs it. The compiler catches type errors before startup, as a TypeScript build can in Node.js. [Go command reference](https://pkg.go.dev/cmd/go)
3. **Register the route.** `newRouter()` creates a standard-library `http.ServeMux`. `mux.HandleFunc("GET /api/health", health)` fills the same role as `app.get('/api/health', health)` in Express. For a small service, Go's built-in router handles path and method matching without a framework dependency. [Go HTTP package](https://pkg.go.dev/net/http#ServeMux)
4. **Write JSON.** The handler receives a response writer and a request, the same two pieces Express passes as `req` and `res`. In place of Express's `res.json({ status: 'ok' })`, the handler sets the content type and uses `encoding/json` to encode a typed struct. The JSON tag controls the field name on the wire, and Go makes the write error explicit through a return value. [Go JSON package](https://pkg.go.dev/encoding/json#Encoder.Encode)
5. **Listen and serve.** `os.Getenv("PORT")` is the counterpart to `process.env.PORT`. `http.Server.ListenAndServe()` fills the same role as `app.listen()`, with a five-second limit on reading request headers. Go serves connections concurrently using goroutines, so handlers can use ordinary synchronous-looking calls. Node commonly uses async I/O on the event loop. Both models handle concurrent clients, and in Go you must protect mutable data shared by handlers. [Go HTTP server](https://pkg.go.dev/net/http#Server.Serve) and [Node.js event loop](https://nodejs.org/learn/asynchronous-work/event-loop-timers-and-nexttick)
6. **Test and build.** `go test ./...` checks the router with `httptest`, similar to using a Node test runner with an HTTP testing helper. `go vet ./...` runs static checks. `go build -o bin/calculator-api .` creates a native executable for your OS and architecture. For this service, deployment can copy the executable without installing the Go toolchain on the target host. [Go HTTP test package](https://pkg.go.dev/net/http/httptest) and [Go build reference](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
