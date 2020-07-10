package wavm

// #cgo CFLAGS: -I/usr/local/include/
// #cgo LDFLAGS: -L/usr/local/lib/ -lWAVM
// #include "WAVM/wavm-c/wavm-c.h"
// int lessbot_call(wasm_store_t* store, wasm_func_t* fn, const char* src, int slen, char *dst) {
// 	wasm_val_t args[3];
// 	wasm_val_t results[1];
// 	args[0].i32 = *((int*)src);
// 	args[1].i32 = slen;
// 	args[2].i32 = 1024;
// 	wasm_func_call(store, fn, args, results);
// 	return results[0].i32;
// }
import "C"
import (
	"fmt"
	"unsafe"
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

func (e *Engine) Call(moduleName string, req []byte) ([]byte, error) {
	fn, ok := e.funcs[moduleName]
	if !ok {
		return nil, fmt.Errorf("invalid module %s", moduleName)
	}
	buf := make([]byte, 1024)
	ret := C.lessbot_call(e.store, fn, (*C.char)(unsafe.Pointer(&req[0])), C.int(len(req)), (*C.char)(unsafe.Pointer(&buf[0])))
	return buf[:ret], nil
}
