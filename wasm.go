package main

import (
	"fmt"

	"github.com/wasmerio/wasmer-go/wasmer"
)

var (
	engine *wasmer.Engine
	store  *wasmer.Store
)

func init() {
	engine = wasmer.NewEngine()
	store = wasmer.NewStore(engine)
}

type WasmBot struct {
	module   *wasmer.Module
	instance *wasmer.Instance
	lessbot  func(...interface{}) (interface{}, error)
}

func LoadWasm(code string) (*WasmBot, error) {
	module, err := wasmer.NewModule(store, []byte(code))

	if err != nil {
		return nil, fmt.Errorf("Failed to compile module: %s", err.Error())
	}

	importObject := wasmer.NewImportObject()
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return nil, fmt.Errorf("Failed to instantiate the module: %s", err.Error())
	}

	lessbot, err := instance.Exports.GetFunction("lessbot")
	if err != nil {
		return nil, fmt.Errorf("Failed to get function lessbot: %s", err.Error())
	}
	return &WasmBot{
		module:   module,
		instance: instance,
		lessbot:  lessbot,
	}, err
}

func (w *WasmBot) Lessbot(msg string) (string, error) {
	w.lessbot(msg)
	return "", nil
}
