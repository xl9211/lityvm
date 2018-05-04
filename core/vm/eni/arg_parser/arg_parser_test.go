package arg_parser
import "bytes"
import "fmt"

// single positive integer (big endian)
func ExamplePosInt(){
    var f, d []byte
    var buf bytes.Buffer;

    f = make([]byte, 1, 1)
    d = make([]byte, 70, 70)
    f[0] = INT
    d[31] = uint8(72) // 32-byte big endian
    parse_entry_point(f, d, &buf)
    fmt.Println(buf.String())
    // Output: [72]
}

// a int and a bool
func ExampleC(){
    var buf bytes.Buffer;

    f := [2]byte{INT, BOOL}
    var d [70] byte
    d[31] = uint8(72) // 32-byte big endian
    parse_entry_point(f[:], d[:], &buf)
    fmt.Println(buf.String())

    // Output: [72,false]
}

// a int and a bool
func ExampleNegInt(){
    var buf bytes.Buffer;
    
    f := [1]byte{INT}
    var d [70] byte
    for i:=0; i<32; i++ { d[i] = uint8(255) }

    parse_entry_point(f[:], d[:], &buf)
    fmt.Println(buf.String())

    // Output: [-1]
}

// two strings
func ExampleString(){
    var buf bytes.Buffer;

    f := [2]byte{STRING, STRING}
    var d [160] byte
    d[31] = uint8(3) // 32-byte big endian
    strA := "abcd"
    copy(d[32:], []byte(strA))
    d[95] = uint8(50)
    strB := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
    copy(d[96:], []byte(strB))

    parse_entry_point(f[:], d[:], &buf)
    fmt.Println(buf.String())

    // Output: ["abc","abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwx"]
}

// an escaped strings
func ExampleEscapedString(){
    var buf bytes.Buffer;

    f := [1]byte{STRING}
    var d [64] byte
    d[31] = uint8(7) // 32-byte big endian
    strA := "abc\"d\\e"
    copy(d[32:], []byte(strA))

    parse_entry_point(f[:], d[:], &buf)
    fmt.Println(buf.String())

    // Output: ["abc\"d\\e"]
}
