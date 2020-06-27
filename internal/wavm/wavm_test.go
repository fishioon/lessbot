package wavm

import (
	"io/ioutil"
	"log"
	"testing"
)

const (
	echoFunc = `(module
  (func (export "run") (param i32) (result i32)
    (call $1 (local.get 0))
  )
);`
)

func TestWAVM(t *testing.T) {
	rawModule, err := ioutil.ReadFile("sum.wasm")
	if err != nil {
		panic(err)
	}
	engine := NewEngine()
	engine.LoadModule("echo", "sum", rawModule)
	res := engine.Call("echo", 2, 3)
	log.Print(res)
}
