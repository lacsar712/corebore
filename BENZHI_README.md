# Corebore

宀╁績 custody lab 杞彂

## Build

```bash
export GOTOOLCHAIN=local
go build ./...
```

## Test

```bash
export GOTOOLCHAIN=local
go test ./... -count=1
```

## Docker (benzhi)

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh corebore linux/amd64
./build_benzhi_docker.sh corebore linux/arm64
```