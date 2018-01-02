package arg_parser
import "testing"
import "bytes"
import "fmt"

// single positive integer (big endian)
func ExampleA(){
    var arg_parser arg_parser_t
    var f, d []byte
    var buf bytes.Buffer;

    f = make([]byte, 1, 1)
    d = make([]byte, 70, 70)
    f[0] = INT
    for i:=0; i<70; i++ { d[i] = uint8(0) }
    d[31] = uint8(72) // 32-byte big endian
    arg_parser.parse_entry_point(f, d, &buf)
    fmt.Println(buf.String())
    // Output: [72]
}

// single positive integer (big endian)
func TestB(t *testing.T){
    var arg_parser arg_parser_t
    var f, d []byte
    var buf bytes.Buffer;

    f = make([]byte, 1, 1)
    d = make([]byte, 70, 70)
    f[0] = INT
    for i:=0; i<70; i++ { d[i] = uint8(0) }
    a:= -72
    d[31] = byte(a) // 32-byte big endian
    arg_parser.parse_entry_point(f, d, &buf)
}

// a int and a bool
func ExampleC(){
    var arg_parser arg_parser_t
    var f, d []byte
    var buf bytes.Buffer;

    f = make([]byte, 2, 2)
    d = make([]byte, 70, 70)
    f[0] = INT
    f[1] = BOOL
    for i:=0; i<70; i++ { d[i] = uint8(0) }
    d[31] = uint8(72) // 32-byte big endian
    arg_parser.parse_entry_point(f, d, &buf)
    fmt.Println(buf.String())

    // Output: [72,false]
}

