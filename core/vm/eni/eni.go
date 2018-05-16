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
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ethereum/go-ethereum/node"
)

type ENI struct {
	// functions & dynamic libraries mapping used by ENI
	functions map[string]string
	// native functions for create & destroy functor
	create  unsafe.Pointer
	destroy unsafe.Pointer
}

func NewENI() *ENI {
	return &ENI{}
}

func (eni *ENI) InitENI(eniFunction string, data string) error {
	// Get functions & dynamic libraries mappings
	functions, err := getEniFunctions()
	if err != nil {
		return err
	}
	eni.functions = functions

	// Check If function exists.
	dynamicLib := make(map[string]string)
	var ok bool
	dynamicLib["create"], ok = eni.functions[eniFunction+"_create"]
	if !ok {
		return errors.New("create function not exists: " + eniFunction)
	}
	dynamicLib["destroy"], ok = eni.functions[eniFunction+"_destroy"]
	if !ok {
		return errors.New("destroy function not exists: " + eniFunction)
	}

	// Check create & destroy method in the same dynamic library
	if dynamicLib["create"] != dynamicLib["destroy"] {
		return errors.New("create and destroy method are not in the same dynamic library")
	}
	dynamicLibName := dynamicLib["create"]

	// Load dynamic library.
	handler := C.dlopen(C.CString(dynamicLibName), C.RTLD_LAZY)
	if handler == nil {
		return errors.New("dlopen failed: " + dynamicLibName)
	}

	// Prepare create & destroy functions.
	eni.create = C.dlsym(handler, C.CString(eniFunction+"_create"))
	if eni.create == nil {
		return errors.New("dlsym failed: " + eniFunction + "_create")
	}
	eni.destroy = C.dlsym(handler, C.CString(eniFunction+"_destroy"))
	if eni.destroy == nil {
		return errors.New("dlsym failed: " + eniFunction + "_destroy")
	}

	// Create functor.
	C.eni_create((*C.eni_create_t)(eni.create), C.CString(data))

	return nil
}

func (eni *ENI) Gas() (uint64, error) {
	return uint64(C.eni_gas(C.functor)), nil
}

func (eni *ENI) ExecuteENI() (string, error) {
	// Run ENI function.
	outputCString := C.eni_run(C.functor)
	outputGoString := C.GoString(outputCString)
	defer C.free(unsafe.Pointer(outputCString))

	// Destroy functor.
	C.eni_destroy((*C.eni_destroy_t)(eni.destroy))

	return outputGoString, nil
}

func getEniFunctions() (map[string]string, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("currently ENI is only supported on Linux")
	}

	// Get dynamic library path.
	libPath := filepath.Join(node.DefaultDataDir(), "eni", "lib")
	if val, ok := os.LookupEnv("ENI_LIBRARY_PATH"); ok {
		libPath = val
	}
	var dynamicLibs []string
	fileinfo, err := os.Stat(libPath)
	if err != nil {
		return nil, err
	}
	if !fileinfo.Mode().IsDir() {
		return nil, errors.New("can't find dynamic library path: " + libPath)
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
			return nil, err
		}

		symbols, err := elf.DynamicSymbols()
		if err != nil {
			return nil, err
		}

		for _, symbol := range symbols {
			functions[symbol.Name] = lib
		}
	}

	return functions, nil
}
