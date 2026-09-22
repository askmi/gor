module example

go 1.27

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/coder/websocket v1.8.15
	github.com/prometheus/client_golang v1.24.1
	github.com/prometheus/otlptranslator v1.0.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.70.0
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/exporters/prometheus v0.67.0
	go.opentelemetry.io/otel/sdk v1.45.0
	go.opentelemetry.io/otel/sdk/metric v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
	gor v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace gor => ..
