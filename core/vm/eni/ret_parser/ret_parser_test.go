package ret_parser
import "fmt"
func ExampleNegInt(){
    f := [1]byte{INT}
    d := Parse(f[:], "[-123]")

    fmt.Printf("%v\n",d)
    // Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}
