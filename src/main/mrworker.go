package main

//
// start a worker process, which is implemented
// in ../mr/worker.go. typically there will be
// multiple worker processes, talking to one coordinator.
//
// go run mrworker.go wc.so
//
// Please do not change this file.
//
import (

	"github.com/sbinet/go-python"
	"6.824/mr"
	"fmt"
	"log"
	"os"
)
var PyStr = python.PyString_FromString
var GoStr = python.PyString_AS_STRING

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrworker xxx\n")
		os.Exit(1)
	}
	module := ImportModule("/root/mapreduce/src/main", os.Args[1])

	if module == nil {
		log.Fatal(os.Args[1])
		log.Fatal("PyImport Fails\n")
	}

	fmt.Fprintf(os.Stderr, "Import Success\n")

	mapf, reducef := loadPlugin(module)

	mr.Worker(mapf, reducef)
}

//
// load the application Map and Reduce functions
// from a plugin file, e.g. ../mrapps/wc.so
//
func loadPlugin(module *python.PyObject) (*python.PyObject, *python.PyObject) {
	mapf := module.GetAttrString("map")
	defer mapf.DecRef()
	reducef := module.GetAttrString("reduce")
	defer reducef.DecRef()
	return mapf, reducef

}
func ImportModule(dir, name string) *python.PyObject {
    sysModule := python.PyImport_ImportModule("sys") // import sys
    path := sysModule.GetAttrString("path")                    // path = sys.path
    python.PyList_Insert(path, 0, PyStr(dir))                     // path.insert(0, dir)
    return python.PyImport_ImportModule(name)            // return __import__(name)
}