package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/example/nebula-edge/internal/discovery"
	"github.com/example/nebula-edge/internal/gossip"
	"github.com/example/nebula-edge/internal/wasm"
)

// Agent combines discovery, gossip and wasm execution.
type Agent struct {
	ctx       context.Context
	cancel    context.CancelFunc
	registry  *wasm.FunctionRegistry
	discover  *discovery.Service
	gossip    *gossip.Gossip
	port      int
	peers     []string
	peersLock sync.RWMutex

	execCounter prometheus.Counter
}

func New(port int) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	reg := wasm.NewRegistry()
	execCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasm_exec_total",
		Help: "Total wasm function executions",
	})
	prometheus.MustRegister(execCounter)
	return &Agent{
		ctx:         ctx,
		cancel:      cancel,
		registry:    reg,
		discover:    discovery.NewService("_nebula-edge._tcp", port),
		port:        port,
		execCounter: execCounter,
	}
}

// Start launches HTTP server, mDNS and gossip.
func (a *Agent) Start() error {
	if err := a.discover.Start(a.ctx); err != nil {
		return err
	}
	g, err := gossip.New(a.ctx)
	if err != nil {
		return err
	}
	a.gossip = g

	mux := http.NewServeMux()
	mux.HandleFunc("/deploy", a.handleDeploy)
	mux.HandleFunc("/exec", a.handleExec)
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Printf("edge agent listening on :%d", a.port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", a.port), mux); err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

type deployRequest struct {
	Name string `json:"name"`
	Wasm string `json:"wasm"` // base64
}

func (a *Agent) handleDeploy(w http.ResponseWriter, r *http.Request) {
	var req deployRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	wasmBytes, err := base64.StdEncoding.DecodeString(req.Wasm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.registry.Register(req.Name, wasmBytes)
	if a.gossip != nil {
		_ = a.gossip.Broadcast(wasmBytes) // ignore error
	}
	w.WriteHeader(http.StatusOK)
}

type execRequest struct {
	Name string `json:"name"`
	Arg  uint64 `json:"arg"`
}

type execResponse struct {
	Result uint64 `json:"result"`
}

func (a *Agent) handleExec(w http.ResponseWriter, r *http.Request) {
	var req execRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := a.registry.Execute(a.ctx, req.Name, req.Arg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.execCounter.Inc()
	_ = json.NewEncoder(w).Encode(execResponse{Result: res})
}
