package eni

import "fmt"

func printOrError(json string, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(json)
	}
}

// single positive integer (big endian)
func ExampleConvertArguments_posInt() {
	var f, d []byte

	f = make([]byte, 1, 1)
	d = make([]byte, 70, 70)
	f[0] = INT
	d[31] = uint8(72) // 32-byte big endian
	json, err := ConvertArguments(f, d)
	printOrError(json, err)
	// Output: [72]
}

func ExampleConvertArguments_bool() {
	f := [1]byte{BOOL}
	var d [32]byte

	json, err := ConvertArguments(f[:], d[:])
	printOrError(json, err)

	for i := 0; i < 32; i++ {
		copy(d[:], make([]byte, 32, 32))
		d[i] = uint8(72) // 32-byte big endian
		json, err = ConvertArguments(f[:], d[:])
		printOrError(json, err)
	}

	// Output: [false]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
	// [true]
}

// a int and a bool
func ExampleConvertArguments_intBool() {
	f := [2]byte{INT, BOOL}
	var d [70]byte
	d[31] = uint8(72) // 32-byte big endian
	json, err := ConvertArguments(f[:], d[:])
	printOrError(json, err)

	// Output: [72,false]
}

// a negative int
func ExampleConvertArguments_negInt() {
	f := [1]byte{INT}
	var d [70]byte
	for i := 0; i < 32; i++ {
		d[i] = uint8(255)
	}

	json, err := ConvertArguments(f[:], d[:])

	printOrError(json, err)
	// Output: [-1]
}

// two strings
func ExampleConvertArguments_string() {
	f := [2]byte{STRING, STRING}
	var d [160]byte
	d[31] = uint8(3) // 32-byte big endian
	strA := "abcd"
	copy(d[32:], []byte(strA))
	d[95] = uint8(50)
	strB := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
	copy(d[96:], []byte(strB))

	json, err := ConvertArguments(f[:], d[:])

	printOrError(json, err)
	// Output: ["abc","abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx"]
}

func ExampleConvertArguments_escapedString() {
	f := [1]byte{STRING}
	var d [64]byte
	d[31] = uint8(7) // 32-byte big endian
	strA := "abc\"d\\e"
	copy(d[32:], []byte(strA))

	json, _ := ConvertArguments(f[:], d[:])
	fmt.Println(json)

	// Output: ["abc\"d\\e"]
}

func ExampleConvertArguments_controlEscapedString() {
	f := [1]byte{STRING}
	var d [64]byte
	d[31] = uint8(9) // 32-byte big endian
	strA := "abc\"d\\\b\x00e"
	copy(d[32:], []byte(strA))

	json, _ := ConvertArguments(f[:], d[:])
	fmt.Println(json)

	// Output: ["abc\"d\\\u0008\u0000e"]
}

// encoding grammaer error
// This happens when Lity byte code generates wrong encoding
func ExampleConvertArguments_errorEncoding1() {
	f := [1]byte{155}
	var d [70]byte
	for i := 0; i < 32; i++ {
		d[i] = uint8(255)
	}

	json, err := ConvertArguments(f[:], d[:])

	printOrError(json, err)
	// Output: Argument Parser Error: encoding error - unknown or not implemented type: 155
}

// encoding grammaer error
// This happens when Lity byte code generates wrong encoding
func ExampleConvertArguments_errorEncoding2() {
	f := [2]byte{STRUCT_START, INT}
	var d [70]byte
	for i := 0; i < 32; i++ {
		d[i] = uint8(255)
	}
	json, err := ConvertArguments(f[:], d[:])

	printOrError(json, err)
	// Output: Argument Parser Error: runtime error: index out of range
}

// data length mismatches ENI type encoding
// This happens when Lity byte code generates wrong data length
func ExampleConvertArguments_errorDataLength() {
	// TODO
}
