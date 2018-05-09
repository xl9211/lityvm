package ret_parser
import "fmt"
import "github.com/ethereum/go-ethereum/core/vm/eni/typecodes"
func ExampleNegInt(){
    f := [1]byte{typecodes.INT}
    d, _ := Parse(f[:], "[-123]")

    fmt.Printf("%v\n",d)
    // Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}

func ExampleFixArray(){
    var f  [34]byte
    f[0] = typecodes.FIX_ARRAY_START
    f[32] = 1
    f[33] = typecodes.INT
    d, err := Parse(f[:], "[[-123]]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // Output: [255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 133]
}

func ExampleError1(){
    f := [1]byte{typecodes.INT}
    d, err := Parse(f[:], "-123")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // Output: expected '[', found '-' 
}

func ExampleError2(){
    f := [1]byte{typecodes.UINT}
    d, err := Parse(f[:], "[-123]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // expected uint, found '-'
}

func ExampleError3(){
    f := [1]byte{typecodes.STRING}
    d, err := Parse(f[:], "[-123]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // Output: expected '"', found '-' 
}

func ExampleErrorInt(){
    f := [1]byte{typecodes.INT}
    d, err := Parse(f[:], "[-]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    f = [1]byte{typecodes.UINT}
    d, err = Parse(f[:], "[-123]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // Output: expected int, found '-'
    // expected uint, found '-'
}

func ExampleErrorBool(){
    f := [1]byte{typecodes.BOOL}
    d, err := Parse(f[:], "[tree]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    f = [1]byte{typecodes.BOOL}
    d, err = Parse(f[:], "[jizz]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // Output: expected 'u', found 'e' 
    // expected boolean, found 'j'
}

func ExampleErrorFixArray(){
    var f  [34]byte
    f[0] = typecodes.FIX_ARRAY_START
    f[32] = 4
    f[33] = typecodes.INT
    d, err := Parse(f[:], "[[-123, 7122, a, 45]]")

    if(err!=nil){
    	fmt.Println(err)
    }else{
    	fmt.Printf("%v\n",d)
    }
    // Output: expected int, found 'a'
}
