# CoreBore

Geological **core sample custody relay**: field crews post signed core-interval and gamma-scan events; the service journals them and forwards to a lab HTTP intake with retry, circuit breaker, rate limit, DLQ, and replay.

Operator UI: http://127.0.0.1:8080/

Not a hospital booking system, CRM, or warehouse inventory app.

## Run

```text
set GOTOOLCHAIN=local
set CGO_ENABLED=0
go test ./...
go run ./cmd/corebore
```

Listens on `:8080`. Default ingest secret: `dev-ingest-secret`. A loopback lab intake is seeded so a test event can succeed without an external URL.

## Event types

- `core.interval.boxed` — depth interval / box label custody handoff
- `core.gamma.scan` — natural-gamma reading bound to a hole/depth
