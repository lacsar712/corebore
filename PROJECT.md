# CoreBore

Geological **core sample custody relay**. Field crews post signed `CoreEvent` messages (core intervals and gamma scans). The service journals them and forwards each matched subscription to a **lab HTTP intake** with retry, circuit breaker, rate limit, dead-letter queue, and operator replay.

This repository is a **healthy runnable baseline** with no pre-planted defects on `main`.

## 1. What this is (and is not)

This is **custody event relay infrastructure**, not:

- Hospital booking / 预约
- CRM or sales pipeline
- Warehouse inventory / WMS
- Game, CLI file tool, ecommerce, RBAC, OA, food delivery, IM, parking, auction
- Analytics dashboard, fitness tracker, recipe app, weather, Pomodoro, habit tracker

Product boundary: a signed field event arrives, is verified and routed, then this process is responsible for **posting JSON to lab intake URLs**. The operator console only manages intakes, journal, DLQ, and replay.

## 2. Roles and happy path

| Role | Action |
|------|--------|
| Field crew / logger | POST `/api/v1/events` with HMAC headers and Idempotency-Key |
| Lab intake | HTTPS/HTTP URL registered by operators; receives JSON core events |
| Operator | Open `/` to register lab intakes, inspect journal, replay failures |

Happy path:

1. Operator registers a lab intake (URL + shared secret + event type filters).
2. Field crew posts a `CoreEvent` (`core.interval.boxed` or `core.gamma.scan`).
3. Engine verifies signature window, nonce, idempotency; matches intakes; enqueues deliveries.
4. Worker posts to lab intake via the lab HTTP client; success is journaled; retryable failures back off; exhausted attempts go to DLQ.
5. Operator can replay from journal or DLQ.

## 3. Domain types

- `CoreEvent` — `{ "type", "payload" }` envelope
- `Interval` — hole_id, from/to depth (m), box_label, optional lithology
- `GammaScan` — hole_id, depth_m, cps, probe_id, optional operator/notes
- Lab intake client — `internal/lab` outbound HTTP with the same HMAC scheme

## 4. Inbound signature

- Algorithm: `HMAC-SHA256(secret, canonical)`, lowercase hex.
- Canonical: `v1.{timestamp}.{nonce}.{sha256_hex(raw_body)}`.
- Headers: `X-Core-Timestamp`, `X-Core-Nonce`, `X-Core-Signature` (`v1=<hex>`), `X-Core-Source-Key`, `Idempotency-Key`.
- Skew window default ±300s; nonce unique within window.
- Ingest secrets are separate from lab outbound secrets.

## 5. Idempotency

- Required `Idempotency-Key` (8–128 chars).
- Same key + same body hash → replay prior accept result.
- Same key + different body hash → 409 Conflict.
- TTL default 24h.

## 6. Fanout to lab intakes

- Each destination: id, name, URL, secret, enabled, type prefixes, ordered flag, rate/burst, max in-flight.
- One inbound event may fan out to many intakes; each has its own queue and journal trail.

## 7. Outbound lab POST

- POST JSON with outbound HMAC using the **lab** secret.
- Extra headers: `X-Core-Event-Id`, `X-Core-Delivery-Id`, `X-Core-Attempt`, `X-Core-Destination`, `X-Core-Body-Sha256`.
- Timeout default 10s; no cross-host redirect follow.
- 2xx success; 408/429/5xx and network/timeout retryable; other 4xx terminal → DLQ.

## 8. Retry / circuit / rate limit / DLQ / replay

- Full-jitter exponential backoff; default maxAttempts=8.
- Per-intake circuit breaker (closed/open/half-open).
- Per-intake token bucket.
- DLQ retains exhausted deliveries; operator replay re-enqueues with a new delivery id lineage via `ReplayOf`.

## 9. Layout

| Path | Role |
|------|------|
| `cmd/corebore` | process entry |
| `internal/accept` | ingest pipeline |
| `internal/event` | CoreEvent / Interval / GammaScan |
| `internal/lab` | lab HTTP intake client |
| `internal/worker` | delivery engine |
| `internal/circuit` | breaker |
| `internal/ratelimit` | token bucket |
| `internal/sign` | HMAC verify/sign |
| `internal/nonce` | nonce book |
| `internal/idempotency` | idem store |
| `internal/replay` | journal/DLQ → job |
| `internal/app` | wiring + replay API |
| `web/` | static operator UI |

## 10. Run

```text
set GOTOOLCHAIN=local
set CGO_ENABLED=0
go test ./...
go run ./cmd/corebore
```

Open http://127.0.0.1:8080/
