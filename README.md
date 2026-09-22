<p align="center">
  <img src="docs/assets/gor-logo.png" width="520" alt="GoR logo">
</p>

<h1 align="center">Go Nature in REST</h1>

<p align="center">
  <strong>Write REST naturally. Keep business handlers pure.</strong><br>
  GoR handles the HTTP boundary without changing how application code feels.
</p>

<p align="center">
  <img src="docs/assets/gor-hero-v5.png" width="900" alt="Two GoR engineers guide typed application data through a vibrant natural system of water, roots, and growing plants">
</p>

GoR is a zero-dependency framework built with the Go standard library. “Natural” describes the developer experience: handlers use familiar Go signatures, native `context.Context`, and application-owned request and response types. GoR does not introduce a custom context or force HTTP types into business code. Application code remains ordinary Go while GoR provides routing, middleware, decoding, encoding, authentication, error mapping, and server lifecycle at the boundary.

The engine includes production-oriented lifecycle primitives for Kubernetes workloads: startup, liveness, and readiness probes; `SIGTERM` and interrupt handling; graceful HTTP shutdown; and deadline-aware hooks for closing application-owned resources. Together, these let a service stop accepting traffic, drain in-flight requests, and clean up resources within the pod's termination grace period.

```go
func CreateOrder(ctx context.Context, command CreateOrderCommand) (Order, error) {
	// Business logic only.
}
```

GoR owns the transport plumbing around that function: routing, middleware, request decoding, error mapping, response encoding, and writing to the network.

The key feature is the typed `RouterFunc` boundary:

```go
type RouterFunc[Req any, Resp any] func(context.Context, Req) (Resp, error)
```

Your function does not import GoR or implement a framework interface. Go infers its request and response types when it is registered, giving the router compile-time type information without leaking transport concerns into the application. Unlike conventional router APIs centered on `http.Handler`, this preserves the experience of writing and testing a normal Go function.

> **Important:** GoR is in early development. Expect API changes before the first stable release.

## Table of contents

- [Idea](#idea)
- [Key Features](#key-features)
- [Why GoR?](#why-gor)
  - [The RouterFunc difference](#the-routerfunc-difference)
  - [Pure Go handlers](#pure-go-handlers)
- [Simple routing](#simple-routing)
  - [Engine-managed routers](#engine-managed-routers)
  - [Native net/http interoperability](#native-nethttp-interoperability)
- [Quick start](#quick-start)
- [Decode query parameters](#decode-query-parameters)
- [Customize the boundaries](#customize-the-boundaries)
- [Authentication](#authentication)
  - [Basic authentication](#basic-authentication)
  - [Bearer/JWT authentication](#bearerjwt-authentication)
  - [Middleware order](#middleware-order)
  - [Security context and principal](#security-context-and-principal)
  - [Add endpoint permissions without changing the handler](#add-endpoint-permissions-without-changing-the-handler)
- [Production readiness](#production-readiness)
  - [Kubernetes probes](#kubernetes-probes)
  - [Container CPU sizing](#container-cpu-sizing)
  - [Server shutdown](#server-shutdown)
  - [Application resource management](#application-resource-management)
  - [OpenTelemetry integration](#opentelemetry-integration)
  - [HTTP resilience](#http-resilience)
  - [Production security](#production-security)
- [Project layout](#project-layout)
- [Example project](#example-project)
- [Development](#development)

## Idea

- **Better remote collaboration:** clear typed boundaries let teammates work independently on transport adapters, middleware, and business handlers.
- **Patterns are communication:** consistent coding patterns form a shared language that communicates intent across locations and time zones.
- **Less repetitive review work:** removing repeated transport plumbing lets reviewers spend more time on business behavior, design decisions, and correctness.

The framework's approach to repetition is informed by the ideas in O'Reilly's archived article [Don't Repeat Yourself](https://web.archive.org/web/20131204221336/http://programmer.97things.oreilly.com/wiki/index.php/Don't_Repeat_Yourself).

## Key Features

- **Pure typed endpoints:** `RouterFunc` infers request and response types through Go generics while handlers remain normal Go functions.
- **Native Go context:** handlers use standard `context.Context`, application request and response types, and `error`.
- **No reflection or third-party dependencies:** the framework stays explicit, compile-time checked, and built on the Go standard library.
- **Explicit HTTP boundaries:** request decoding, response encoding, status codes, and error mapping stay outside business logic and can be replaced.
- **Error management:** centralized error handling, HTTP error mapping, and logging keep failure behavior consistent across endpoints.
- **Standard middleware:** middleware composes through `func(http.Handler) http.Handler`, so standard Go and third-party HTTP middleware work directly.
- **Routing and lifecycle:** routers support path prefixes, per-router and per-endpoint middleware, dynamic mounting, managed server startup and graceful shutdown, OS signal handling, completion notification, and application-resource cleanup hooks.
- **Authentication and authorization:** basic and bearer credential extraction, pluggable authenticators, security contexts, and application-owned principals and roles.
- **Native interoperability:** mount any `http.Handler` directly for streaming, files, protocol upgrades, or specialized HTTP behavior.

## Why GoR?

In many services, handlers become tightly coupled to HTTP. They read path values, decode bodies, select status codes, serialize responses, and write headers alongside business decisions. This repetition makes handlers harder to read, test, and reuse.

With GoR, the public shape of an endpoint is a normal typed Go function:

```go
func(context.Context, RequestModel) (ResponseModel, error)
```

The function can be tested directly, called without an HTTP server, and understood as a business operation. Transport-specific work stays in replaceable adapters at the edge of the application.

### The RouterFunc difference

`RouterFunc` connects a pure application function to HTTP without changing that function’s signature. The router creates the request model, invokes the function, maps its result or error, and writes the HTTP response. Your business code sees none of those steps:

```go
func AddUser(ctx context.Context, req AddUserRequest) (AddUserResponse, error) {
	// Pure Go application code.
}

router.Post("/users", AddUser)
```

This gives you framework capabilities at runtime while preserving a natural Go experience in the handler itself.

### Pure Go handlers

GoR encourages developers to keep business handlers vendor-agnostic by using only standard Go and application-owned types in their signatures—without GoR, HTTP, database-driver, or other vendor-specific types. This keeps the business layer portable and independent of infrastructure choices.

The following method is the reference handler shape:

```go
func (h *H) GetUser(ctx context.Context, userID GetUserID) (GetUserResponse, error) {
	// Pure Go business logic.
}
```

`context.Context` and `error` come from Go, while `GetUserID`, `GetUserResponse`, and `H` belong to the application. GoR does not wrap or replace `context.Context` with a framework-specific context type. The handler does not know which router decoded the request, which protocol delivered it, or which component will encode its response. This is pure Go code and can be called directly like any other method.

GoR turns that flow into a small, explicit pipeline:

```text
HTTP request
    → Router
    → Middleware
    → RequestHandler
    → typed RouterFunc
    → ResponseHandler / ErrorHandler
    → ResponseWriter
```

You keep control of each boundary and can replace its behavior when the defaults do not fit.

## Simple routing

### Engine-managed routers

Create and mount a router through the engine when the engine owns the complete server:

```go
engine := gor.NewEngine()

engine.NewRouter("/api/v1/").
	Get("/users/{id}", getUser).
	Post("/users", addUser)
```

`Engine.NewRouter` creates the router and mounts it automatically.

Alternatively, configure a router independently and mount it explicitly:

```go
router := gor.NewRouter("/api/v1/").
	Get("/users/{id}", getUser).
	Post("/users", addUser)

engine := gor.NewEngine().Route(router)
```

Use the first form for concise application setup. Use the second when routers are created in separate packages, tested independently, or shared with a standard `http.Server`.

### Native `net/http` interoperability

Typed routes and native HTTP handlers can coexist on the same router. Use `HandleHTTP` for streaming, files, protocol upgrades, or existing Go HTTP libraries:

```go
router.HandleHTTPFunc("GET /events", streamEvents)
```

Because `Router` implements `http.Handler`, it can also be used directly with Go's standard HTTP server without the GoR engine:

```go
router := gor.NewRouter("/api/v1/").
	Get("/users/{id}", getUser).
	Post("/users", addUser)

server := &http.Server{
	Addr:              ":8080",
	Handler:           router,
	ReadHeaderTimeout: 5 * time.Second,
}

log.Fatal(server.ListenAndServe())
```

## Quick start

```go
package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	gor "gor/pkg/server"
)

type empty struct{}

func helloWorld(_ context.Context, _ empty) (string, error) {
	return "Hello, World!", nil
}

type GetUserID int

func (id *GetUserID) DecodeFromHTTPRequest(req *http.Request) error {
	value, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		return err
	}
	*id = GetUserID(value)
	return nil
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getUser(_ context.Context, userID GetUserID) (User, error) {
	return User{ID: int(userID), Name: "Ada"}, nil
}

func main() {
	router := gor.NewRouter("/api/")
	router.Use(
		gor.RecoveryMiddleware,
		gor.ResponseWriterStatusCodeMiddleware,
	)

	router.
		Get("/hello", helloWorld).
		Get("/users/{id}", getUser)

	engine := gor.NewEngine()
	engine.Route(router)

	if err := engine.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
```

`Listen` starts the server and blocks until it stops. See [Server shutdown](#server-shutdown) for signal handling and shutdown deadlines.

`helloWorld` is a complete endpoint with no HTTP-specific code. `getUser` demonstrates the same pure function shape with an application-owned path value and response model. `GetUserID.DecodeFromHTTPRequest` is a transport adapter; it can be replaced globally through `UseRequestHandler` when business models should contain no HTTP-aware methods at all.

Run the complete example from the repository root:

```bash
cd example
go run ./cmd
```

Then request `http://localhost:8080/api/users/42` after adapting the example's configured authentication.

## Decode query parameters

A request can be a small application-owned type that knows how to decode itself from HTTP. For example, this type reads and validates the required `name` query parameter:

```go
var ErrNameRequired = errors.New("name query parameter is required")

type NameQuery string

func (q *NameQuery) DecodeFromHTTPRequest(req *http.Request) error {
	value := req.URL.Query().Get("name")
	if value == "" {
		return ErrNameRequired
	}

	*q = NameQuery(value)
	return nil
}
```

`DecodeFromHTTPRequest` can validate input as it decodes it. When it returns an error, GoR skips the handler and passes that error to the router's `ErrorHandler`:

```go
router.UseErrorHandler(func(ctx context.Context, err error) gor.HTTPResponse {
	if errors.Is(err, ErrNameRequired) {
		return gor.NewJSONResponse(
			http.StatusBadRequest,
			`{"error":"name is required"}`,
		)
	}
	return gor.DefaultErrorHandler(ctx, err)
})
```

The handler receives the decoded value as an ordinary Go type:

```go
func hello(_ context.Context, name NameQuery) (string, error) {
	return "Hello world, " + string(name), nil
}

router.Get("/hello", hello)
```

Call the endpoint with:

```bash
curl -vvv -u "admin:admin" \
  "http://localhost:8080/api/v1/hello?name=Alex"
```

## Customize the boundaries

Each router exposes focused extension points:

- `UseRequestHandler` decodes incoming HTTP requests.
- `UseResponseHandler` converts application values into HTTP responses.
- `UseErrorHandler` maps application errors into HTTP responses.
- `UseResponseWriter` controls the final write to `http.ResponseWriter`.

### Response status codes and route options

- A successful response uses `200 OK` by default; typed values are JSON-encoded.
- An error returns `500 Internal Server Error`.
- An `HTTPResponse` keeps its own status code.

Use `Get`, `Post`, `Put`, and `Delete` for ordinary typed endpoints. These methods accept a path without an HTTP method prefix:

```go
router.
	Get("/users/{id}", h.GetUser).
	Post("/users", h.AddUser).
	Put("/users/{id}", h.EditUser).
	Delete("/users/{id}", h.DeleteUser)
```

`Post` maps a successful response to `201 Created`, while `Delete` maps one to `204 No Content`. Use `HandleFunc` with `WithStatusCode` when an endpoint needs a different success status:

```go
router.HandleFunc(
	"GET /reports/{id}",
	h.GetReport,
	gor.WithStatusCode(http.StatusAccepted),
)
```

The option applies only to that route. Other endpoints keep the router's default response mapping. The status code stays in the routing layer, so `h.AddUser` remains a pure Go business function with no HTTP-specific return type.

Use `UseResponseHandler` to customize successful responses and `UseErrorHandler` to map decoding or handler errors, as shown in [Decode query parameters](#decode-query-parameters).

### Middleware scope: `Use` vs `With`

| Method | Scope |
| --- | --- |
| `router.Use(middleware...)` | Global: applies to every endpoint registered on that router afterward. |
| `router.With(middleware)` | Specific: returns a router copy used only to register selected endpoints. The original router is unchanged. |

```go
// Common middleware for all endpoints registered afterward.
router.Use(
	gor.RecoveryMiddleware,
	gor.AuthenticationMiddleware(authenticator),
)

// Extra middleware for selected endpoints only.
router.With(Authorize("admin")).
	Delete("/users/{id}", h.DeleteUser)

router.Get("/users/{id}", h.GetUser)
```

## Authentication

Credential extraction and authentication are separate by design. GoR provides middleware that reads credentials from the `Authorization` header, but the application owns the policy for validating those credentials and constructing its principal.

### Basic authentication

`BasicMiddleware` extracts the encoded username and password. The application's `Authenticator` decodes and validates them, then returns either `gor.Authenticated` or `gor.Rejected`:

```go
type Principal struct {
	Username string
	Roles    []string
}

func BasicAuthenticator(expectedUser, expectedPassword string) gor.Authenticator {
	return func(s gor.SecurityContext) (gor.SecurityContext, error) {
		raw, ok := s.Identity().([]byte)
		if !ok {
			return gor.Rejected("invalid credentials"), nil
		}

		username, password, ok := gor.DecodeBasic(raw)
		if !ok || string(username) != expectedUser || string(password) != expectedPassword {
			return gor.Rejected("invalid credentials"), nil
		}

		principal := Principal{
			Username: string(username),
			Roles:    []string{"admin"},
		}
		return gor.Authenticated(principal.Username, principal), nil
	}
}

router.Use(
	gor.BasicMiddleware,
	gor.AuthenticationMiddleware(BasicAuthenticator("admin", "secret")),
)
```

This example uses fixed credentials only to show the contract. A real application can validate against its database, identity provider, secret store, or another application-owned service.

### Bearer/JWT authentication

`BearerMiddleware` extracts the token without imposing a token format. Implement JWT parsing, signature and claim validation, key selection, and principal construction in the application, then pass that authenticator to GoR:

```go
router.Use(
	gor.BearerMiddleware,
	gor.AuthenticationMiddleware(JwtAuthenticator),
)
```

The demo JWT authenticator is in [`example/internal/jwt.go`](example/internal/jwt.go). Keeping the `Authenticator` on the application side lets each service choose its credential source, JWT library, claims, key rotation strategy, and identity model without putting those policies into the framework.

### Middleware order

Authentication must be last in the credential-processing part of the middleware chain:

```go
router.Use(
	gor.RecoveryMiddleware,
	gor.ResponseWriterStatusCodeMiddleware,
	gor.SimpleLoggingMiddleware,
	gor.BearerMiddleware, // 1. Extract the raw credential.
	gor.AuthenticationMiddleware(jwtAuthenticator), // 2. Validate it.
)
```

GoR executes middleware in declaration order. `AuthenticationMiddleware` therefore has to come after `BasicMiddleware`, `BearerMiddleware`, or another credential-extraction middleware; otherwise there is no `SecurityContext` for it to authenticate and the request receives `401 Unauthorized`. Authorization middleware must run after authentication so it sees the validated principal.

Use the Basic and Bearer pipelines separately unless the application authenticator is deliberately designed to accept both credential types.

### Security context and principal

The request context carries a `gor.SecurityContext`, which exposes the authentication state, a stable identity string, and the application-defined identity. Retrieve an authenticated principal with its concrete application type:

```go
principal, ok := gor.PrincipalFromContext[Principal](ctx)
if !ok {
	return ErrUnauthenticated
}
```

`PrincipalFromContext` succeeds only for an authenticated context whose identity has the requested type. Use `SecurityFromContext` when code needs the complete security state. The concrete principal is implementation-specific: it can be a username, user record, claims object, or a struct such as `Principal` above containing roles and other authorization data. `IdentityString()` provides a stable textual identity for logging or display without requiring consumers to understand that concrete type.

Application-specific authorization remains ordinary middleware, keeping business rules explicit and testable.

### Add endpoint permissions without changing the handler

For example, an app-owned `Authorize` middleware can read roles from the authenticated principal:

```go
func Authorize(requiredRole string) gor.HTTPMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := gor.PrincipalFromContext[Principal](r.Context())
			if !ok || !slices.Contains(principal.Roles, requiredRole) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

Use `router.With` to add authorization only to selected endpoints. Calls can be chained because each registration returns the router:

```go
router.With(Authorize("admin")).
	Get("/user/{id}", h.GetUser).
	Delete("/user/{id}", h.DeleteUser)
```

`Authorize("admin")` runs only for endpoints registered with that router copy. The handlers need no permission-related parameters:

```go
func (h *H) GetUser(ctx context.Context, userID GetUserID) (GetUserResponse, error) {
	// Business logic remains unchanged.
	return GetUserResponse{}, nil
}
```

The demo implementation is in [`example/internal/mdw.go`](example/internal/mdw.go).

## Production readiness

GoR provides the lifecycle and observability building blocks needed to run an HTTP service in a Kubernetes cluster while staying close to `net/http`, `log/slog`, and standard middleware contracts.

| Concern | Support |
| --- | --- |
| Kubernetes | Built-in startup, liveness, readiness, and compatibility health endpoints |
| Lifecycle | Configurable OS signals, graceful HTTP shutdown, shutdown timeout, and completion notification |
| Resources | Context-aware cleanup hooks for databases, telemetry providers, consumers, and other resources |
| Observability | OpenTelemetry trace propagation, trace-correlated structured logs, HTTP metrics, and custom metrics |
| HTTP resilience | Configurable server timeouts and header limits, panic recovery, response status recording, replayable request bodies |
| Runtime | Container-aware Go scheduling and a minimal non-root `scratch` image |
| Security | Basic and bearer extraction, application-owned authentication, route-scoped authorization |

### Kubernetes probes

Enable lightweight HTTP probes on the engine:

```go
engine := gor.NewEngine().EnableProbes()
engine.Route(router)
```

The default successful probe paths are `/startupz`, `/livez`, `/readyz`, `/healthz`, and `/health`. Configure Kubernetes with the specific lifecycle endpoints:

```yaml
startupProbe:
  httpGet: {path: /startupz, port: 8080}
livenessProbe:
  httpGet: {path: /livez, port: 8080}
readinessProbe:
  httpGet: {path: /readyz, port: 8080}
```

Probe responses use `application/health+json`. The built-in endpoints are shallow process checks; they confirm that the process can answer HTTP, not that every dependency is healthy. When readiness depends on a database, queue, or another resource, register an application-owned readiness handler instead of the default one.

### Container CPU sizing

Kubernetes CPU requests and limits serve different purposes. The scheduler uses `requests.cpu` to place Pods, and the request determines the container's relative CPU weight during contention. A CPU limit is a hard CPU-time ceiling enforced by Linux cgroups; exceeding it throttles the container instead of terminating it. For example, `500m` represents half of one logical CPU's processing time. See [Kubernetes resource management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

Go 1.25 and later automatically derive the default `GOMAXPROCS` from the smaller of the logical CPU count, CPU affinity, and the container's cgroup CPU limit. Go observes the CPU **limit**, not the request, rounds fractional limits up, and normally keeps `GOMAXPROCS` at least `2`. It also periodically detects limit changes. Setting the `GOMAXPROCS` environment variable or calling `runtime.GOMAXPROCS` disables this automatic behavior. GoR targets Go 1.27, so an additional `automaxprocs` dependency is not needed. See [container-aware `GOMAXPROCS`](https://go.dev/blog/container-aware-gomaxprocs) and the current [`runtime` documentation](https://pkg.go.dev/runtime#GOMAXPROCS).

For latency-sensitive services, begin with a measured request and consider omitting the CPU limit so the service can use idle node capacity without cgroup throttling:

```yaml
resources:
  requests:
    cpu: 500m
    memory: 128Mi
  limits:
    memory: 256Mi
```

Add a CPU limit when strict workload isolation is more important than burst capacity, then load-test that exact value. Avoid setting `GOMAXPROCS` manually unless measurements show that Go's default is unsuitable—especially for sub-CPU limits, where Go's minimum and rounding behavior matter. Monitor CPU usage, `container_cpu_cfs_throttled_periods_total`, scheduler latency, and HTTP p95/p99 latency. Size CPU requests carefully because percentage-based HPA CPU utilization is calculated relative to the request.

The example uses a two-stage build with a static Go binary and a non-root `scratch` runtime containing only the binary, CA roots, and static files. Build it from the repository root because the example module replaces `gor` with its parent directory:

```bash
docker build --pull -f example/Dockerfile -t gor-example:latest .
```

### Server shutdown

`Listen` blocks for the server lifecycle. Enable managed signals and configure the shutdown timeout before constructing the engine:

```go
gor.DefaultGracefulTimeout = 25 * time.Second

engine := gor.NewEngine().
	EnableSignals(os.Interrupt, syscall.SIGTERM).
	Route(router)

if err := engine.Listen(":8080"); err != nil {
	log.Fatal(err)
}
```

Calling `EnableSignals()` without arguments uses interrupt and `SIGTERM`. When a configured signal arrives, the engine calls `http.Server.Shutdown` with `DefaultGracefulTimeout`; this stops new connections and gives active requests time to finish. `StopGracefully(ctx)` is available when the application owns signal handling or initiates shutdown programmatically, and `Done()` reports when serving has ended.

In Kubernetes, set `terminationGracePeriodSeconds` longer than the engine timeout so the kubelet does not send `SIGKILL` before shutdown and cleanup finish:

```yaml
spec:
  terminationGracePeriodSeconds: 45
  containers:
    - name: app
      ports:
        - name: http
          containerPort: 8080
      startupProbe:
        httpGet: {path: /startupz, port: http}
      livenessProbe:
        httpGet: {path: /livez, port: http}
      readinessProbe:
        httpGet: {path: /readyz, port: http}
```

### Application resource management

Register application resources on the engine so shutdown waits for cleanup within the same deadline:

```go
engine.
	OnShutdownWithContext(func(ctx context.Context) {
		if err := meterProvider.Shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "metrics shutdown failed", "error", err)
		}
	}).
	OnShutdownWithContext(func(ctx context.Context) {
		if err := consumer.Shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "consumer shutdown failed", "error", err)
		}
	}).
	OnShutdown(func() {
		_ = database.Close()
	})
```

Hooks run in registration order. Each context-aware hook should return when `ctx.Done()` is closed; a context communicates cancellation but cannot forcibly stop a function. Use `OnShutdown(func())` only for short cleanup operations that do not accept a context.

### OpenTelemetry integration

GoR uses standard HTTP middleware and `context.Context`, so OpenTelemetry propagation works without a framework-specific adapter. Install the global providers first, then place `otelhttp` before logging middleware:

```go
otel.SetTracerProvider(tracerProvider)
otel.SetMeterProvider(meterProvider)
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
))

router.Use(
	otelhttp.NewMiddleware("users-service"),
	gor.ResponseWriterStatusCodeMiddleware,
	gor.SimpleLoggingMiddleware,
)
```

The middleware extracts incoming trace headers, places the span in the request context, and records standard HTTP server metrics. Use context-aware `slog` calls to correlate application logs:

```go
slog.InfoContext(ctx, "user loaded", "user_id", userID)
```

[`example/internal/telemetry.go`](example/internal/telemetry.go) contains a `TraceLogHandler` that adds `trace_id`, plus a Prometheus-backed meter provider with a dedicated registry and application label. Expose its returned handler like any other native HTTP handler:

```go
meterProvider, metricsHandler, err := SetupMeter()
if err != nil {
	return err
}

engine.NewRouter("").HandleHTTP("GET /metrics", metricsHandler)
```

Custom metrics can be added with a typed route decorator. The instrument is created once during route setup, and the returned function records each call before returning the original result:

```go
func UserCounter[Req, Resp any](
	next gor.RouterFunc[Req, Resp],
) gor.RouterFunc[Req, Resp] {
	counter, err := otel.Meter("users").Int64Counter("users_total")
	if err != nil {
		panic("create users counter: " + err.Error())
	}

	return func(ctx context.Context, req Req) (Resp, error) {
		resp, err := next(ctx, req)
		status := "success"
		if err != nil {
			status = "failure"
		}

		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "get_user"),
			attribute.String("status", status),
		))
		return resp, err
	}
}

router.Get("/users/{id}", UserCounter(h.GetUser))
```

Keep metric attributes low-cardinality: do not use user IDs, request IDs, email addresses, or raw URLs as labels. Register tracer and meter provider shutdown through `OnShutdownWithContext` so buffered telemetry is flushed during application shutdown.

### HTTP resilience

The provided middleware can recover panics, record response status, log requests, and make buffered request bodies available again through `req.GetBody()`:

```go
router.Use(
	gor.RecoveryMiddleware,
	gor.ResponseWriterStatusCodeMiddleware,
	gor.SimpleLoggingMiddleware,
	gor.ReplayBodyMiddleware,
)
```

`ReplayBodyMiddleware` buffers the complete body in memory. Apply an application-appropriate request-size limit before it for untrusted or potentially large bodies. The engine accepts standard server timeout and header-limit options, while `Router` still implements `http.Handler` for applications that need direct control of `http.Server`.

Configure the engine's underlying `http.Server` with chainable `ServerOpts`:

```go
serverOpts := gor.NewServerOpts().
	WithReadHeaderTimeout(5 * time.Second).
	WithReadTimeout(15 * time.Second).
	WithWriteTimeout(30 * time.Second).
	WithIdleTimeout(60 * time.Second).
	WithMaxHeaderBytes(1 << 20)

engine := gor.NewEngine(serverOpts)
```

TLS remains application- or ingress-managed until engine TLS configuration is implemented.

### Production security

Keep credential extraction, authentication, and authorization in that order:

```go
router.Use(
	gor.BearerMiddleware,
	gor.AuthenticationMiddleware(authenticator),
)

router.With(Authorize("admin")).
	Delete("/users/{id}", h.DeleteUser)
```

Use either the Basic or bearer extraction pipeline unless the authenticator intentionally supports both. The `authenticator` and `Authorize` functions above are application-owned; GoR does not implement JWT verification, authorization policy, TLS, or rate limiting. Use Basic authentication only over TLS, terminate TLS at the application or ingress boundary, avoid logging credentials, and apply route-scoped authorization with `With`. Do not put secrets in URLs because the default request logger includes the request URI. See [Authentication](#authentication) for the complete flow.

Environment-based engine configuration and engine-managed TLS are tracked in [`TODO.md`](TODO.md). Until they are implemented, configure them in the application or at the Kubernetes ingress/proxy boundary.

## Project layout

```text
gor/
├── pkg/
│   ├── server/                 # Routing, middleware, and server lifecycle
│   │   ├── adapter.go          # Engine/router constructors and typed handlers
│   │   ├── engine.go           # Server lifecycle
│   │   ├── option.go           # HTTP server configuration
│   │   ├── interface.go        # Public framework contracts
│   │   ├── middleware.go       # Logging, recovery, and authentication
│   │   ├── router.go           # Routes and customization points
│   │   └── *_test.go           # Server package tests
│   ├── client/                 # HTTP client package
│   └── repository/             # Repository package
├── example/                    # Complete demo service (separate Go module)
│   ├── cmd/main.go             # Program entry point
│   ├── internal/
│   │   ├── app.go              # Application composition
│   │   ├── handler.go          # Typed business handlers
│   │   ├── mdw.go              # Application middleware
│   │   └── telemetry.go        # Tracing, logging, and metrics setup
│   ├── static/                 # Static-file example
│   ├── Dockerfile              # Two-stage scratch image
│   └── go.mod
├── docs/assets/                # Project branding
├── go.mod
├── README.md
└── TODO.md
```

## Example project

The [`example`](example/) folder contains a complete demo project showing how to use GoR effectively when writing services. It demonstrates typed business handlers, request models, routing, middleware, authentication, authorization, error mapping, JSON responses, logging, and static file serving in one small application.

Use it as a practical starting point for organizing a GoR-based service and for seeing how transport concerns remain separate from handler business logic.

## Development

```bash
go test ./...

cd example
go test ./...
```

Contributions and practical feedback from real service codebases are welcome.
