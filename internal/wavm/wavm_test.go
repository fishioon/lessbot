package wavm

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestWAVM(t *testing.T) {
	rawModule, err := ioutil.ReadFile("sum.wasm")
	if err != nil {
		panic(err)
	}
	engine := NewEngine()
	engine.LoadModule("echo", "sum", rawModule)
	res, _ := engine.Call("echo", []byte("hello"))
	log.Print(res)
}
