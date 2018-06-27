package eni

import (
	"fmt"
)

func retPrintOrError(d []byte, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v", d)
	}
}

func ExampleConvertReturnValue_negInt() {
	f := [1]byte{INT}
	d, _ := ConvertReturnValue(f[:], "[-123]")

	fmt.Printf("%v\n", d)
	// Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}

func ExampleConvertReturnValue_string() {
	f := [1]byte{STRING}
	d, _ := ConvertReturnValue(f[:], "[\"-123abc1\"]")

	fmt.Printf("%v\n", d)
	// Output: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 8 45 49 50 51 97 98 99 49 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleConvertReturnValue_escapedString() {
	f := [1]byte{STRING}
	json := "[\"-123\\\"\"]"
	d, err := ConvertReturnValue(f[:], json)

	retPrintOrError(d, err)
	// Output: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 5 45 49 50 51 34 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleConvertReturnValue_controlEscapedString() {
	f := [1]byte{STRING}
	json := "[\"-123\\\"\\n\\u0010\"]"
	d, err := ConvertReturnValue(f[:], json)

	retPrintOrError(d, err)
	// Output: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 7 45 49 50 51 34 10 16 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleConvertReturnValue_unicodeEscapedString() {
	f := [1]byte{STRING}
	json := "[\"-123\\\"\\n\\u7122\"]"
	d, err := ConvertReturnValue(f[:], json)

	retPrintOrError(d, err)
	// Output: Return Parser Error: UTF-8 not implemented yet!
}

func ExampleConvertReturnValue_fixArray() {
	var f [34]byte
	f[0] = FIX_ARRAY_START
	f[32] = 1
	f[33] = INT
	d, err := ConvertReturnValue(f[:], "[[-123]]")

	retPrintOrError(d, err)
	// Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}

func ExampleConvertReturnValue_error1() {
	f := [1]byte{INT}
	d, err := ConvertReturnValue(f[:], "-123")

	retPrintOrError(d, err)
	// Output: Return Parser Error: expected '[', found '-'
}

func ExampleConvertReturnValue_error2() {
	f := [1]byte{UINT}
	d, err := ConvertReturnValue(f[:], "[-123]")

	retPrintOrError(d, err)
	// Output: Return Parser Error: expected uint, found '-'
}

func ExampleConvertReturnValue_error3() {
	f := [1]byte{STRING}
	d, err := ConvertReturnValue(f[:], "[-123]")

	retPrintOrError(d, err)
	// Output: Return Parser Error: expected '"', found '-'
}

func ExampleConvertReturnValue_errorInt() {
	f := [1]byte{INT}
	d, err := ConvertReturnValue(f[:], "[-]")

	retPrintOrError(d, err)
	f = [1]byte{UINT}
	d, err = ConvertReturnValue(f[:], "[-123]")

	retPrintOrError(d, err)
	// Output: Return Parser Error: expected int, found '-'
	// Return Parser Error: expected uint, found '-'
}

func ExampleConvertReturnValue_errorBool() {
	f := [1]byte{BOOL}
	d, err := ConvertReturnValue(f[:], "[tree]")

	retPrintOrError(d, err)
	f = [1]byte{BOOL}
	d, err = ConvertReturnValue(f[:], "[jizz]")

	retPrintOrError(d, err)
	// Output: Return Parser Error: expected 'u', found 'e'
	// Return Parser Error: expected boolean, found 'j'
}

func ExampleConvertReturnValue_errorFixArray() {
	var f [34]byte
	f[0] = FIX_ARRAY_START
	f[32] = 4
	f[33] = INT
	d, err := ConvertReturnValue(f[:], "[[-123, 7122, a, 45]]")

	retPrintOrError(d, err)
	// Output: Return Parser Error: expected int, found 'a'
}

// Json not matched with ENI type encoding
func ExampleConvertReturnValue_errorEcodingJsonMismatch() {
	// TODO:
}

// JSON format error
// This happens when libeni returns wrong JSON
func ExampleConvertReturnValue_errorJsonFormat1() {
	// TODO:
}

// JSON integer error
// This happens when libeni returns any too large integers
func ExampleConvertReturnValue_errorJsonFormat2() {
	// TODO:
}
