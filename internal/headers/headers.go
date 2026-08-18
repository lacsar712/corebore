package headers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	Timestamp     = "X-Core-Timestamp"
	Nonce         = "X-Core-Nonce"
	Signature     = "X-Core-Signature"
	Idempotency   = "Idempotency-Key"
	SourceKey     = "X-Core-Source-Key"
	EventID       = "X-Core-Event-Id"
	DeliveryID    = "X-Core-Delivery-Id"
	Attempt       = "X-Core-Attempt"
	DestinationID = "X-Core-Destination"
)

type Inbound struct {
	Timestamp int64
	Nonce     string
	Signature string
	IdemKey   string
	SourceKey string
}

func ParseInbound(h http.Header) (Inbound, error) {
	var in Inbound
	ts := strings.TrimSpace(h.Get(Timestamp))
	if ts == "" {
		return in, fmt.Errorf("missing %s", Timestamp)
	}
	n, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || n <= 0 {
		return in, fmt.Errorf("invalid %s", Timestamp)
	}
	in.Timestamp = n
	in.Nonce = strings.TrimSpace(h.Get(Nonce))
	if in.Nonce == "" {
		return in, fmt.Errorf("missing %s", Nonce)
	}
	in.Signature = strings.TrimSpace(h.Get(Signature))
	if in.Signature == "" {
		return in, fmt.Errorf("missing %s", Signature)
	}
	in.IdemKey = strings.TrimSpace(h.Get(Idempotency))
	if in.IdemKey == "" {
		return in, fmt.Errorf("missing %s", Idempotency)
	}
	in.SourceKey = strings.TrimSpace(h.Get(SourceKey))
	if in.SourceKey == "" {
		in.SourceKey = "ingest"
	}
	return in, nil
}
