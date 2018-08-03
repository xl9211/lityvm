package eni

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/node"
)

type OTAInfo struct {
	libName  string
	version  string // The format of version should be vX.Y.Z where X, Y, Z are all integers. E.g. v1.0.0, v3.2.0
	url      string // URL to retrieve the library file
	checksum string // SHA512 checksum to check the health of the library
}

type OTAInstance struct {
	availableInfos map[string]OTAInfo
	enableInfos    map[string]OTAInfo
}

func NewOTAInstance() *OTAInstance {
	ota := OTAInstance{
		availableInfos: make(map[string]OTAInfo),
		enableInfos:    make(map[string]OTAInfo),
	}
	return &ota
}

// Use OTAInfo to generate real path name of a single library
func generateFileName(info OTAInfo) string {
	fileName := info.libName + "_" + info.version + ".so"
	return fileName
}

type Version struct {
	major int
	minor int
	patch int
}

func NewVersion() *Version {
	v := Version{
		major: 0,
		minor: 0,
		patch: 0,
	}
	return &v
}

// Convert version string to a Version struct
func (v *Version) BuildFromString(version string) error {
	// Check the version format
	versionReg := regexp.MustCompile(`\Av\d+\.\d+\.\d+\z`)
	if !versionReg.MatchString(version) {
		return errors.New("The format of version is invalid")
	}

	versionSlice := strings.Split(version[1:], ".")
	v.major, _ = strconv.Atoi(versionSlice[0])
	v.minor, _ = strconv.Atoi(versionSlice[1])
	v.patch, _ = strconv.Atoi(versionSlice[2])
	return nil
}

// Compare version between two OTAInfos
// Return:
//   this > version -> 1
//   this = version -> 0
//   this < version -> -1
func (v *Version) Compare(a Version) int {
	if v.major > a.major {
		return +1
	} else if v.major < a.major {
		return -1
	} else {
		if v.minor > a.minor {
			return +1
		} else if v.minor < a.minor {
			return -1
		} else {
			if v.patch > a.patch {
				return +1
			} else if v.patch < a.patch {
				return -1
			} else {
				return 0
			}
		}
	}
}

// Check a given OTAInfo is valid to upgrade system setting
func (ota *OTAInstance) IsValidUpgrade(info OTAInfo) (bool, error) {
	currentVersion := NewVersion()
	err := currentVersion.BuildFromString(ota.enableInfos[info.libName].version)
	if err != nil {
		return false, err
	}
	nextVersion := NewVersion()
	err = nextVersion.BuildFromString(info.version)
	if err != nil {
		return false, err
	}

	if nextVersion.Compare(*currentVersion) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// Download the library to staging folder
func (ota *OTAInstance) Download(info OTAInfo) (err error) {
	hashKey := info.libName + info.version
	// Cache OTAInfo to available list.
	if _, exist := ota.availableInfos[hashKey]; !exist {
		ota.availableInfos[hashKey] = info
	}

	err = downloadFromUrl(info.url, generateFileName(info))
	if err != nil {
		return err
	}
	return nil
}

// Verify downloaded staging libraries.
func (ota *OTAInstance) Verify(info OTAInfo) (err error) {
	libPath, err := getLibPath()
	if err != nil {
		return err
	}

	stagingLibPath := filepath.Join(libPath, "staging")
	libFile, err := os.Open(filepath.Join(stagingLibPath, generateFileName(info)))
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
		os.Remove(filepath.Join(stagingLibPath, generateFileName(info)))
		return errors.New("Library " + info.libName + " checksum doesn't match")
	}
	return nil
}

// Register staging libraries to lib.
func (ota *OTAInstance) Register(info OTAInfo) (err error) {
	// Overwrite old libraries by libName.
	ota.enableInfos[info.libName] = info
	libPath, err := getLibPath()
	if err != nil {
		return err
	}
	stagingLibPath := filepath.Join(libPath, "staging")
	if err != nil {
		return err
	}
	err = os.Rename(
		filepath.Join(stagingLibPath, generateFileName(info)),
		filepath.Join(libPath, generateFileName(info)))
	if err != nil {
		return err
	}
	return nil
}

// Remove unused libraries from lib and staging folder
func (ota *OTAInstance) Destroy(info OTAInfo) (err error) {
	libPath, err := getLibPath()
	if err != nil {
		return err
	}
	stagingLibPath := filepath.Join(libPath, "staging")
	if err != nil {
		return err
	}

	// Check lib folder first.
	fileName := filepath.Join(libPath, generateFileName(info))
	if _, err := os.Stat(fileName); err == nil {
		err = os.Remove(fileName)
		if err != nil {
			return err
		}
	}

	// Check staging folder first.
	fileName = filepath.Join(stagingLibPath, generateFileName(info))
	if _, err := os.Stat(fileName); err == nil {
		err = os.Remove(fileName)
		if err != nil {
			return err
		}
	}

	return nil
}

// Get libPath from default data path or ENI_LIBRARY_PATH
func getLibPath() (libPath string, err error) {
	libPath = filepath.Join(node.DefaultDataDir(), "eni", "lib")
	if val, ok := os.LookupEnv("ENI_LIBRARY_PATH"); ok {
		libPath = val
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
}

// Download the library from given url. The library will
// be saved in ENI_LIBRARY_PATH/staging named OTAInfo.libName.
func downloadFromUrl(url string, libName string) (err error) {
	libPath, err := getLibPath()
	stagingLibPath := filepath.Join(libPath, "staging")
	if err != nil {
		return err
	}

	// If file is existed, we don't need to download it again.
	fileName := filepath.Join(stagingLibPath, libName)
	if _, err := os.Stat(fileName); err == nil {
		return nil
	}

	// Create the output file.

	// If the output file is broken, remove it.
	defer func() {
		if rec := recover(); rec != nil {
			os.Remove(fileName)
			err = rec.(error)
		}
	}()

	output, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		panic(err)
	}

	return nil
}
