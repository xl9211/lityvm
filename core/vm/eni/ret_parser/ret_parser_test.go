package ret_parser

import "fmt"
import "github.com/ethereum/go-ethereum/core/vm/eni/typecodes"

func printOrError(d []byte, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v", d)
	}
}

func ExampleParse_negInt() {
	f := [1]byte{typecodes.INT}
	d, _ := Parse(f[:], "[-123]")

	fmt.Printf("%v\n", d)
	// Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}

func ExampleParse_string() {
	f := [1]byte{typecodes.STRING}
	d, _ := Parse(f[:], "[\"-123abc1\"]")

	fmt.Printf("%v\n", d)
	// Output: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 8 45 49 50 51 97 98 99 49 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleParse_escapedString() {
	f := [1]byte{typecodes.STRING}
	json := "[\"-123\\\"\"]"
	d, err := Parse(f[:], json)

	printOrError(d, err)
	// Output: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 5 45 49 50 51 34 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleParse_controlEscapedString() {
	f := [1]byte{typecodes.STRING}
	json := "[\"-123\\\"\\n\\u0010\"]"
	d, err := Parse(f[:], json)

	printOrError(d, err)
	// Output: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 7 45 49 50 51 34 10 16 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleParse_unicodeEscapedString() {
	f := [1]byte{typecodes.STRING}
	json := "[\"-123\\\"\\n\\u7122\"]"
	d, err := Parse(f[:], json)

	printOrError(d, err)
	// Output: Return Parser Error: UTF-8 not implemented yet!
}

func ExampleParse_fixArray() {
	var f [34]byte
	f[0] = typecodes.FIX_ARRAY_START
	f[32] = 1
	f[33] = typecodes.INT
	d, err := Parse(f[:], "[[-123]]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}

func ExampleParse_error1() {
	f := [1]byte{typecodes.INT}
	d, err := Parse(f[:], "-123")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: Return Parser Error: expected '[', found '-'
}

func ExampleParse_error2() {
	f := [1]byte{typecodes.UINT}
	d, err := Parse(f[:], "[-123]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: Return Parser Error: expected uint, found '-'
}

func ExampleParse_error3() {
	f := [1]byte{typecodes.STRING}
	d, err := Parse(f[:], "[-123]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: Return Parser Error: expected '"', found '-'
}

func ExampleParse_errorInt() {
	f := [1]byte{typecodes.INT}
	d, err := Parse(f[:], "[-]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	f = [1]byte{typecodes.UINT}
	d, err = Parse(f[:], "[-123]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: Return Parser Error: expected int, found '-'
	// Return Parser Error: expected uint, found '-'
}

func ExampleParse_errorBool() {
	f := [1]byte{typecodes.BOOL}
	d, err := Parse(f[:], "[tree]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	f = [1]byte{typecodes.BOOL}
	d, err = Parse(f[:], "[jizz]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: Return Parser Error: expected 'u', found 'e'
	// Return Parser Error: expected boolean, found 'j'
}

func ExampleParse_errorFixArray() {
	var f [34]byte
	f[0] = typecodes.FIX_ARRAY_START
	f[32] = 4
	f[33] = typecodes.INT
	d, err := Parse(f[:], "[[-123, 7122, a, 45]]")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\n", d)
	}
	// Output: Return Parser Error: expected int, found 'a'
}

// Json not matched with ENI type encoding
func ExampleParse_errorEcodingJsonMismatch() {
	// TODO:
}

// JSON format error
// This happens when libeni returns wrong JSON
func ExampleParse_errorJsonFormat1() {
	// TODO:
}

// JSON integer error
// This happens when libeni returns any too large integers
func ExampleParse_errorJsonFormat2() {
	// TODO:
}
