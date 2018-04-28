package eni

/*
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>
#cgo LDFLAGS: -ldl -leni

typedef void* eni_create_t(char* pArgStr);
typedef void  eni_destroy_t(void* pFunctor);

extern uint64_t eni_gas(void* pFunctor);
extern char* eni_run(void *pFunctor);

void* functor;

void set_functor(void* f) {
  functor = f;
}

void eni_create(eni_create_t* f, char *data) {
  functor = f(data);
}

void eni_destroy(eni_destroy_t* f) {
  f(functor);
}
*/
import "C"

import (
	"debug/elf"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ethereum/go-ethereum/node"
)

type ENI struct {
	// functions & dynamic libraries mapping used by ENI
	functions map[string]string
}

func NewENI() *ENI {
	return &ENI{
		functions: getEniFunctions(),
	}
}

func (eni *ENI) ExecuteENI(eniFunction string, data string) string {
	// Check If function exists.
	dynamicLib := make(map[string]string)
	var ok bool
	dynamicLib["create"], ok = eni.functions[eniFunction+"_create"]
	if !ok {
		panic("create function not exists: " + eniFunction)
	}
	dynamicLib["destroy"], ok = eni.functions[eniFunction+"_destroy"]
	if !ok {
		panic("destroy function not exists: " + eniFunction)
	}

	// Check create & destroy method in the same dynamic library
	if dynamicLib["create"] != dynamicLib["destroy"] {
		panic("create and destroy method are not in the same dynamic library")
	}
	dynamicLibName := dynamicLib["create"]

	// Load dynamic library.
	handler := C.dlopen(C.CString(dynamicLibName), C.RTLD_LAZY)
	if handler == nil {
		panic("dlopen failed: " + dynamicLibName)
	}

	// Prepare create & destroy functions.
	eni_create := C.dlsym(handler, C.CString(eniFunction+"_create"))
	if eni_create == nil {
		panic("dlsym failed: " + eniFunction)
	}
	eni_destroy := C.dlsym(handler, C.CString(eniFunction+"_destroy"))
	if eni_destroy == nil {
		panic("dlsym failed: " + eniFunction)
	}

	// Create functor.
	C.eni_create((*C.eni_create_t)(eni_create), C.CString(data))

	// Calculate gas usage.
	_ = uint64(C.eni_gas(C.functor))

	// Run ENI function.
	outputCString := C.eni_run(C.functor)
	outputGoString := C.GoString(outputCString)
	defer C.free(unsafe.Pointer(outputCString))

	// Destroy functor.
	C.eni_destroy((*C.eni_destroy_t)(eni_destroy))

	return outputGoString
}

func getEniFunctions() map[string]string {
	if runtime.GOOS != "linux" {
		return nil
	}

	// Get dynamic library path.
	libPath := filepath.Join(node.DefaultDataDir(), "eni", "lib")
	if val, ok := os.LookupEnv("ENI_LIBRARY_PATH"); ok {
		libPath = val
	}
	var dynamicLibs []string
	fileinfo, err := os.Stat(libPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		panic(err)
	}
	if !fileinfo.Mode().IsDir() {
		panic("Can't file dynamic library path: " + libPath)
	}

	// Get all dynamic libraries from library path.
	filepath.Walk(libPath, func(path string, fileinfo os.FileInfo, err error) error {
		if fileinfo.Mode().IsRegular() && filepath.Ext(path) == ".so" {
			dynamicLibs = append(dynamicLibs, path)
		}
		return nil
	})

	// Build mapping for functions and libraries
	functions := make(map[string]string)
	for _, lib := range dynamicLibs {
		elf, err := elf.Open(lib)
		if err != nil {
			panic(err)
		}

		symbols, err := elf.DynamicSymbols()
		if err != nil {
			panic(err)
		}

		for _, symbol := range symbols {
			functions[symbol.Name] = lib
		}
	}

	return functions
}
