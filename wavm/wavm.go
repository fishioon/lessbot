package wavm

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -lWAVM
// #include "WAVM/wavm-c/wavm-c.h"
// int lessbot_call(wasm_store_t* store, wasm_func_t* fn, int offset, int len) {
// 	wasm_val_t args[2];
// 	wasm_val_t results[1];
// 	args[0].i32 = offset;
// 	args[1].i32 = len;
// 	wasm_func_call(store, fn, args, results);
// 	return results[0].i32;
// }
import "C"
import (
	"fmt"
)

const (
	botFuncName = "lessbot"
)

type Mod struct {
	fn          *C.wasm_func_t
	mem         *C.wasm_memory_t
	offset, max int
}

func (m *Mod) read(offset, size int) string {
	data := C.wasm_memory_data(m.mem)
	return C.GoStringN(data, C.int(size))
}

func (m *Mod) write(req string) (int, int) {
	// data := C.wasm_memory_data(m.mem)
	offset := m.offset
	size := len(req)
	/*
		for i := 0; i < size; i++ {
			data[i] = C.char(req[i])
		}
	*/
	return offset, size
}

type Engine struct {
	engine      *C.wasm_engine_t
	compartment *C.wasm_compartment_t
	store       *C.wasm_store_t
	mods        map[string]*Mod
}

func NewEngine() *Engine {
	engine := C.wasm_engine_new()
	compartment := C.wasm_compartment_new(engine, C.CString("compartment"))
	store := C.wasm_store_new(compartment, C.CString("store"))
	return &Engine{engine, compartment, store, make(map[string]*Mod)}
}

func (e *Engine) LoadModule(name string, code []byte) error {
	initMem := func() *C.wasm_memory_t {
		mt := C.wasm_memorytype_new(&C.wasm_limits_t{min: 1, max: 1}, 1)
		return C.wasm_memory_new(e.compartment, mt, C.CString("memory"))
	}
	m := C.wasm_module_new(e.engine, (*C.char)(C.CBytes(code)), (C.ulong)(len(code)))
	if m == nil {
		return fmt.Errorf("load %s module fail", name)
	}
	mem := initMem()
	imports := []*C.wasm_extern_t{C.wasm_memory_as_extern(mem)}
	ins := C.wasm_instance_new(e.store, m, &imports[0], nil, C.CString("instance"))
	for i := 0; i < int(C.wasm_module_num_exports(m)); i++ {
		var exp C.wasm_export_t
		C.wasm_module_export(m, C.ulong(i), &exp)
		if C.GoString(exp.name) == botFuncName {
			ext := C.wasm_instance_export(ins, C.ulong(i))
			if fn := C.wasm_extern_as_func(ext); fn != nil {
				e.mods[name] = &Mod{
					fn:  fn,
					mem: mem,
				}
				break
			}
		}
	}
	C.wasm_module_delete(m)
	C.wasm_instance_delete(ins)
	if e.mods[name] == nil {
		return fmt.Errorf("module %s not found %s func", name, botFuncName)
	}
	return nil
}

func (e *Engine) Call(name string, req string) (string, error) {
	mod, ok := e.mods[name]
	if !ok {
		return "", fmt.Errorf("invalid module %s", name)
	}
	offset, size := mod.write(req)
	ret := C.lessbot_call(e.store, mod.fn, C.int(offset), C.int(size))
	return mod.read(offset, int(ret)), nil
}
