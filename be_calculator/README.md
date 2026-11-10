# Calculator API

A Go REST API starter with one endpoint. `GET /api/health` returns `200 OK` and `{"status":"ok"}`. The endpoint confirms the server can accept and answer HTTP requests.

## Run locally

Install a current patch release of [Go 1.26 or newer](https://go.dev/dl/). From this folder, start the API with `go run .`. In another terminal, try:

```sh
curl -i http://localhost:3000/api/health
```

The frontend's Vite proxy forwards `/api` to port `3000`. To run the API by itself on a different port, use `PORT=3001 go run .`. The server listens on all interfaces.

## From Node.js to Go, step by step

1. **Declare the module.** `go.mod` records the module name and minimum Go version. It serves a similar role to the dependency and runtime fields in `package.json`. The Go installation supplies all the packages this project imports, so setup needs no equivalent of `npm install`. [Go module reference](https://go.dev/ref/mod)
2. **Start at the entry point.** `package main` and `func main()` tell Go where execution starts, similar to running `node server.js`. `go run .` compiles the current package and immediately runs it. The compiler catches type errors before startup, as a TypeScript build can in Node.js. [Go command reference](https://pkg.go.dev/cmd/go)
3. **Register the route.** `newRouter()` creates a standard-library `http.ServeMux`. `mux.HandleFunc("GET /api/health", health)` fills the same role as `app.get('/api/health', health)` in Express. For a small service, Go's built-in router handles path and method matching without a framework dependency. [Go HTTP package](https://pkg.go.dev/net/http#ServeMux)
4. **Write JSON.** The handler receives a response writer and a request, the same two pieces Express passes as `req` and `res`. In place of Express's `res.json({ status: 'ok' })`, the handler sets the content type and uses `encoding/json` to encode a typed struct. The JSON tag controls the field name on the wire, and Go makes the write error explicit through a return value. [Go JSON package](https://pkg.go.dev/encoding/json#Encoder.Encode)
5. **Listen and serve.** `os.Getenv("PORT")` is the counterpart to `process.env.PORT`. `http.Server.ListenAndServe()` fills the same role as `app.listen()`, with a five-second limit on reading request headers. Go serves connections concurrently using goroutines, so handlers can use ordinary synchronous-looking calls. Node commonly uses async I/O on the event loop. Both models handle concurrent clients, and in Go you must protect mutable data shared by handlers. [Go HTTP server](https://pkg.go.dev/net/http#Server.Serve) and [Node.js event loop](https://nodejs.org/learn/asynchronous-work/event-loop-timers-and-nexttick)
6. **Test and build.** `go test ./...` checks the router with `httptest`, similar to using a Node test runner with an HTTP testing helper. `go vet ./...` runs static checks. `go build -o bin/calculator-api .` creates a native executable for your OS and architecture. For this service, deployment can copy the executable without installing the Go toolchain on the target host. [Go HTTP test package](https://pkg.go.dev/net/http/httptest) and [Go build reference](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
