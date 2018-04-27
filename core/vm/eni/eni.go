package eni

import (
	"debug/elf"
	"os"
	"path/filepath"
	"runtime"
)

func GetEniFunctions() map[string]string {
	if runtime.GOOS != "linux" {
		return nil
	}

	libPath := "./lib"
	var dynamicLibs []string
	fileinfo, err := os.Stat(libPath)
	if err != nil {
		panic(err)
	}
	if !fileinfo.Mode().IsDir() {
		panic("Can't file dynamic library path: " + libPath)
	}

	filepath.Walk(libPath, func(path string, fileinfo os.FileInfo, err error) error {
		if fileinfo.Mode().IsRegular() && filepath.Ext(path) == ".so" {
			dynamicLibs = append(dynamicLibs, path)
		}
		return nil
	})

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
