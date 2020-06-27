package wavm

// #cgo CFLAGS: -I/usr/local/include/
// #cgo LDFLAGS: -L/usr/local/lib/ -lWAVM
// #include "WAVM/wavm-c/wavm-c.h"
import "C"
import (
	"fmt"
)

type Engine struct {
	engine      *C.wasm_engine_t
	compartment *C.wasm_compartment_t
	store       *C.wasm_store_t
	funcs       map[string]*C.wasm_func_t
}

func NewEngine() *Engine {
	engine := C.wasm_engine_new()
	compartment := C.wasm_compartment_new(engine, C.CString("compartment"))
	store := C.wasm_store_new(compartment, C.CString("store"))
	return &Engine{engine, compartment, store, make(map[string]*C.wasm_func_t)}
}

func (e *Engine) LoadModule(moduleName, funcName string, code []byte) error {
	m := C.wasm_module_new(e.engine, (*C.char)(C.CBytes(code)), (C.ulong)(len(code)))
	if m == nil {
		return fmt.Errorf("load %s module fail", moduleName)
	}
	ins := C.wasm_instance_new(e.store, m, nil, nil, C.CString("instance"))
	for i := 0; i < int(C.wasm_module_num_exports(m)); i++ {
		var exp C.wasm_export_t
		C.wasm_module_export(m, C.ulong(i), &exp)
		if C.GoString(exp.name) == funcName {
			ext := C.wasm_instance_export(ins, C.ulong(i))
			if fn := C.wasm_extern_as_func(ext); fn != nil {
				e.funcs[moduleName] = fn
				break
			}
		}
	}
	C.wasm_module_delete(m)
	C.wasm_instance_delete(ins)
	if e.funcs[moduleName] == nil {
		return fmt.Errorf("module %s not found %s func", moduleName, funcName)
	}
	return nil
}

func (e *Engine) Execute(moduleName, req string) (string, error) {
	fn, ok := e.funcs[moduleName]
	if !ok {
		return "", fmt.Errorf("module %s not found", moduleName)
	}
	args := make([]C.wasm_val_t, 1)
	results := make([]C.wasm_val_t, 1)
	C.wasm_func_call(e.store, fn, &args[0], &results[0])
	return "", nil
}

func (e *Engine) Call(moduleName string, a, b int) int {
	fn, ok := e.funcs[moduleName]
	if !ok {
		return 0
	}
	args := make([]C.wasm_val_t, 2)
	args[0] = C.wasm_val_t{byte(a)}
	args[1] = C.wasm_val_t{byte(b)}
	results := make([]C.wasm_val_t, 1)
	C.wasm_func_call(e.store, fn, &args[0], &results[0])
	return int(results[0][0])
}
