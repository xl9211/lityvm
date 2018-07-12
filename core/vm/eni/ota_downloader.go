package eni

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/node"
)

type OTAInfo struct {
	libName           string
	version           string
	url               string // URL to retrieve the library file
	checksum          string // SHA512 checksum to check the health of the library
	deployedTimestamp string // Timestamp to all nodes to download the library after proposal is approved
	reason            string
	expiredTimestamp  string // Remain timestamp before expire
}

type OTAInstance struct {
	availableInfos map[string]OTAInfo
	enableInfos    map[string]OTAInfo
}

func NewOTAInstance() *OTAInstance {
	ota := OTAInstance{availableInfos: make(map[string]OTAInfo), enableInfos: make(map[string]OTAInfo)}
	return &ota
}

// Download the library to staging folder
func (ota *OTAInstance) Download(info OTAInfo) (err error) {
	hashKey := info.libName + info.version
	// Cache OTAInfo to available list.
	if _, exist := ota.availableInfos[hashKey]; !exist {
		ota.availableInfos[hashKey] = info
	}

	err = downloadFromUrl(info.url, info.libName)
	if err != nil {
		return err
	}
	return nil
}

// Verify downloaded staging libraries.
func (ota *OTAInstance) Verify(info OTAInfo) (err error) {
	libPath, err := getLibPath("staging")
	if err != nil {
		return err
	}

	libFile, err := os.Open(filepath.Join(libPath, info.libName))
	if err != nil {
		return err
	}
	defer libFile.Close()

	hasher := sha512.New()
	if _, err := io.Copy(hasher, libFile); err != nil {
		return err
	}

	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	if checksum != info.checksum {
		return errors.New("Library " + info.libName + " checksum doesn't match")
	}
	return nil
}

// Register staging libraries to lib.
func (ota *OTAInstance) Register(info OTAInfo) (err error) {
	// Overwrite old libraries by libName.
	ota.enableInfos[info.libName] = info
	stagingLibPath, err := getLibPath("staging")
	if err != nil {
		return err
	}
	libPath, err := getLibPath("lib")
	if err != nil {
		return err
	}
	err = os.Rename(filepath.Join(stagingLibPath, info.libName), filepath.Join(libPath, info.libName))
	if err != nil {
		return err
	}
	return nil
}

// Remove unused libraries from lib and staging folder
func (ota *OTAInstance) Destroy(info OTAInfo) (err error) {
	stagingLibPath, err := getLibPath("staging")
	if err != nil {
		return err
	}
	libPath, err := getLibPath("lib")
	if err != nil {
		return err
	}

	// Check lib folder first.
	fileName := filepath.Join(libPath, info.libName)
	if _, err := os.Stat(fileName); err == nil {
		err = os.Remove(fileName)
		if err != nil {
			return err
		}
	}

	// Check staging folder first.
	fileName = filepath.Join(stagingLibPath, info.libName)
	if _, err := os.Stat(fileName); err == nil {
		err = os.Remove(fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get libPath from default data path or ENI_LIBRARY_PATH
func getLibPath(phase string) (libPath string, err error) {
	if phase == "lib" || phase == "staging" {
		libPath = filepath.Join(node.DefaultDataDir(), "eni", phase)
		if val, ok := os.LookupEnv("ENI_LIBRARY_PATH"); ok {
			libPath = filepath.Join(val, "..", phase)
		}

		// Check if the ENI_LIBRARY_PATH is existed.
		fileinfo, err := os.Stat(libPath)
		if err != nil {
			return "", err
		}
		if !fileinfo.Mode().IsDir() {
			return "", errors.New("can't find dynamic library path: " + libPath)
		}
		return libPath, nil
	} else {
		return "", errors.New("Unknown phase")
	}
}

// Download the library from given url. The library will
// be saved in ENI_LIBRARY_PATH/../staging named OTAInfo.libName.
func downloadFromUrl(url string, libName string) (err error) {
	libPath, err := getLibPath("staging")
	if err != nil {
		return err
	}

	// If file is existed, we don't need to download it again.
	fileName := filepath.Join(libPath, libName)
	if _, err := os.Stat(fileName); err == nil {
		return nil
	}

	// Create the output file.
	output, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	return nil
}
