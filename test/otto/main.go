package main

import (
	"fmt"
	"github.com/robertkrimen/otto"
)

func main() {
	//d := encrypt.JsParser("./encrypt/encrypt.js", "encodeInp", "abc")
	//fmt.Println("d: ", d)
	get()

}
func get() {
	vm := otto.New()
	script := `
    var n = 100;
    console.log("hello-" + n);   
    n;
  `
	value, _ := vm.Run(script)
	fmt.Println("value:", value.String())
	fmt.Println()

}

func set() {
	vm := otto.New()
	vm.Set("def", 11)
	vm.Run(`
		console.log("The value of def is " + def);
	
	`)

}
