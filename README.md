# Nebula-Edge MVP

A minimal peer‑to‑peer WebAssembly function mesh written in **Go**.

```
tree -L 2
.
├── cmd
│   ├── edge        # edge agent binary
│   └── nebctl      # CLI for deploying wasm
├── internal
│   ├── agent
│   ├── discovery
│   ├── gossip
│   └── wasm
└── sample
    └── double      # TinyGo example
```

## Quick start

### 1. Build edge agent & CLI

```bash
go mod tidy
go build -o edge ./cmd/edge
go build -o nebctl ./cmd/nebctl
```

### 2. Start agent on two machines (or two terminals)

```bash
./edge -port 8080
./edge -port 8081
```

Agents advertise via mDNS (`_nebula-edge._tcp`) and exchange code over libp2p‑gossipsub.

### 3. Build sample wasm

```bash
cd sample/double
tinygo build -o double.wasm -target=wasi
```

### 4. Deploy

```bash
./nebctl -wasm sample/double/double.wasm -name double -addr http://localhost:8080
```

### 5. Execute

```bash
curl -X POST -d '{"name":"double","arg":21}' http://localhost:8081/exec
# => {"result":42}
```

Prometheus metrics exposed at `http://localhost:8080/metrics`.

