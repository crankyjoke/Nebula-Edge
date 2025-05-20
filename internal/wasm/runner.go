package wasm

import (
	"context"
	"errors"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// FunctionRegistry stores wasm modules in-memory.
type FunctionRegistry struct {
	mu   sync.RWMutex
	code map[string][]byte
}

func NewRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		code: make(map[string][]byte),
	}
}

func (r *FunctionRegistry) Register(name string, wasmBytes []byte) {
	r.mu.Lock()
	r.code[name] = wasmBytes
	r.mu.Unlock()
}

func (r *FunctionRegistry) Execute(ctx context.Context, name string, args ...uint64) (uint64, error) {
	r.mu.RLock()
	bytes, ok := r.code[name]
	r.mu.RUnlock()
	if !ok {
		return 0, errors.New("function not found")
	}
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		return 0, err
	}

	compiled, err := runtime.CompileModule(ctx, bytes)
	if err != nil {
		return 0, err
	}

	cfg := wazero.NewModuleConfig().WithStartFunctions()
	mod, err := runtime.InstantiateModule(ctx, compiled, cfg)

	if err != nil {
		return 0, err
	}
	fn := mod.ExportedFunction(name)
	if fn == nil {
		return 0, errors.New("export not found")
	}
	results, err := fn.Call(ctx, args...)
	if err != nil {
		return 0, err
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return 0, nil
}
