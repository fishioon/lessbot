package wavm

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestWAVM(t *testing.T) {
	rawModule, err := ioutil.ReadFile("hello.wasm")
	if err != nil {
		panic(err)
	}
	engine := NewEngine()
	engine.LoadModule("echo", rawModule)
	res, _ := engine.Call("echo", "hello")
	log.Print(res)
}
