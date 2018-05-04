package ret_parser

import "fmt"
import "bytes"
import "math/big"
import "github.com/ethereum/go-ethereum/common/math"
// token constant
const (
    BOOL = iota
    ADDRESS
    BYTES
    ENUM
    STRING
    FIX_ARRAY_START
    DYN_ARRAY_START
    STRUCT_START
    STRUCT_END
    INT
    INT8
    INT16
    INT24
    INT32
    INT40
    INT48
    INT56
    INT64
    INT72
    INT80
    INT88
    INT96
    INT104
    INT112
    INT120
    INT128
    INT136
    INT144
    INT152
    INT160
    INT168
    INT176
    INT184
    INT192
    INT200
    INT208
    INT216
    INT224
    INT232
    INT240
    INT248
    INT256
    UINT
    UINT8
    UINT16
    UINT24
    UINT32
    UINT40
    UINT48
    UINT56
    UINT64
    UINT72
    UINT80
    UINT88
    UINT96
    UINT104
    UINT112
    UINT120
    UINT128
    UINT136
    UINT144
    UINT152
    UINT160
    UINT168
    UINT176
    UINT184
    UINT192
    UINT200
    UINT208
    UINT216
    UINT224
    UINT232
    UINT240
    UINT248
    UINT256
    BYTE1
    BYTE2
    BYTE3
    BYTE4
    BYTE5
    BYTE6
    BYTE7
    BYTE8
    BYTE9
    BYTE10
    BYTE11
    BYTE12
    BYTE13
    BYTE14
    BYTE15
    BYTE16
    BYTE17
    BYTE18
    BYTE19
    BYTE20
    BYTE21
    BYTE22
    BYTE23
    BYTE24
    BYTE25
    BYTE26
    BYTE27
    BYTE28
    BYTE29
    BYTE30
    BYTE31
    BYTE32    
)

// need type parsing 
var complexType = map[byte]bool{
    FIX_ARRAY_START: true,
    DYN_ARRAY_START: true,
    STRUCT_START: true,
    STRING: true,
}

// in bytes
// only for value type
var dataLen = map[byte]int{
    BOOL:1,
    INT:32,
    UINT:32,
    INT8:1,
    INT16:2,
    INT24:3,
    INT32:4,
    INT40:5,
    INT48:6,
    INT56:7,
    INT64:8,
    INT72:9,
    INT80:10,
    INT88:11,
    INT96:12,
    INT104:13,
    INT112:14,
    INT120:15,
    INT128:16,
    INT136:17,
    INT144:18,
    INT152:19,
    INT160:20,
    INT168:21,
    INT176:22,
    INT184:23,
    INT192:24,
    INT200:25,
    INT208:26,
    INT216:27,
    INT224:28,
    INT232:29,
    INT240:30,
    INT248:31,
    INT256:32,
    UINT8:1,
    UINT16:2,
    UINT24:3,
    UINT32:4,
    UINT40:5,
    UINT48:6,
    UINT56:7,
    UINT64:8,
    UINT72:9,
    UINT80:10,
    UINT88:11,
    UINT96:12,
    UINT104:13,
    UINT112:14,
    UINT120:15,
    UINT128:16,
    UINT136:17,
    UINT144:18,
    UINT152:19,
    UINT160:20,
    UINT168:21,
    UINT176:22,
    UINT184:23,
    UINT192:24,
    UINT200:25,
    UINT208:26,
    UINT216:27,
    UINT224:28,
    UINT232:29,
    UINT240:30,
    UINT248:31,
    UINT256:32,
    BYTE1:1,
    BYTE2:2,
    BYTE3:3,
    BYTE4:4,
    BYTE5:5,
    BYTE6:6,
    BYTE7:7,
    BYTE8:8,
    BYTE9:9,
    BYTE10:10,
    BYTE11:11,
    BYTE12:12,
    BYTE13:13,
    BYTE14:14,
    BYTE15:15,
    BYTE16:16,
    BYTE17:17,
    BYTE18:18,
    BYTE19:19,
    BYTE20:20,
    BYTE21:21,
    BYTE22:22,
    BYTE23:23,
    BYTE24:24,
    BYTE25:25,
    BYTE26:26,
    BYTE27:27,
    BYTE28:28,
    BYTE29:29,
    BYTE30:30,
    BYTE31:31,
    BYTE32:32}

// TODO
func Parse(type_info []byte, jsonStr string) ([]byte) {
    json := []byte(jsonStr)
    skip_ws(&json)
    expect(&json, '[')
    var dataBuf bytes.Buffer
    for i:=0; 0<len(type_info); i++ {
        if 0<i {
            skip_ws(&json)
            expect(&json, ',')
        }
        type_info, json = parse_type(type_info, &dataBuf, []byte(json))
    }
    // dataBuf.Write(data)
    skip_ws(&json)
    expect(&json, ']')
    return dataBuf.Bytes()
}

// assuming that data are packed
func parse_type(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
    t := type_info[0]
    if complexType[t]{
        if t==FIX_ARRAY_START{
            type_info, json = parse_fix_array(type_info, data, json)
        }else if t==DYN_ARRAY_START{
            type_info, json = parse_dyn_array(type_info, data, json)
        }else if t==STRUCT_START{
            type_info, json = parse_struct(type_info, data, json)
        }else if t==STRING{
            type_info, json = parse_string(type_info, data, json)
        }else{// error

        }
    } else {// value type
        type_info, json = parse_value(type_info, data, json)
    }
    return type_info, json
}

func parse_string(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
    type_info = type_info[1:] // string
    skip_ws(&json)
    expect(&json, '"')
    length := int64(0)
    for json[length]!='"'{length++}

    data.Write(math.PaddedBigBytes(big.NewInt(length), 32))
    data.Write(json[:length])
    if length%32>0 { data.Write(make([]byte, 32-length))}
    json = json[length:]
    expect(&json, '"')
    return type_info, json
}

func parse_dyn_array(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte){
    // TODO
    return type_info, json
}

// parsing int32 not finished
func parse_fix_array(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte){
    type_info = type_info[1:] // fix_array_start
    skip_ws(&json)
    expect(&json, '[')
    leng := int(type_info[0]) // TODO: parsing a 32-byte integer
    type_info = type_info[32:] // 

    for i:=0; i<leng; i++{
        if i==leng-1 {
            type_info, json = parse_type(type_info, data, json)
        }else{
            skip_ws(&json)
            expect(&json, ',')
            _, json = parse_type(type_info, data, json)
        }
    }

    skip_ws(&json)
    expect(&json, ']')
    return type_info, json
}

func parse_struct(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte){
    skip_ws(&json)
    expect(&json, '[')
    for i:=0; ; i++{
        t := type_info[0]
        skip_ws(&json)
        expect(&json, ',')
        if t!=STRUCT_END{
            type_info, json = parse_type(type_info, data, json)
        }
    }
    type_info = type_info[1:] // struct_end
    skip_ws(&json)
    expect(&json, ']')
    return type_info, json
}

func parse_value(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte){
    t := type_info[0]
    skip_ws(&json)
    if t==BOOL{
        if have(&json,'t') {
            expect(&json,'r')
            expect(&json,'u')
            expect(&json,'e')
            data.WriteByte(byte(1))
        }else if have(&json,'f'){
            expect(&json,'a')
            expect(&json,'l')
            expect(&json,'s')
            expect(&json,'e')
            data.WriteByte(byte(0))
        }else{// err

        }
    }else if INT<=t && t<=INT256{ // signed integer
        i := 0
        ojson := json
        if have(&json,'-') { i++ }
        for have_digit(&json) { i++ }
        var n big.Int
        n.SetString(string(ojson[0:i]), 10)
        b := math.PaddedBigBytes(n, 32)
        data.Write(b)
    }else if (UINT<=t && t<=UINT256) || (BYTE1<=t && t<=BYTE32){// unsigned integer
        i := 0
        ojson := json
        for have_digit(&json) { i++ }
        var n big.Int
        n.SetString(string(ojson[0:i]), 10)
        b := math.PaddedBigBytes(n, 32)
        data.Write(b)
    }
    type_info = type_info[1:]
    return type_info, json
}


func have(json *[]byte, c byte) (bool){
    if (*json)[0]==c {
        *json = (*json)[1:]
        return true
    }else{
        return false
    }
}

func have_digit(json *[]byte) (bool){
    if (*json)[0]>='9' && (*json)[0]<='9' {
        *json = (*json)[1:]
        return true
    }else{
        return false
    }
}

func expect(json *[]byte, c byte){
    expectMsg(json, c, fmt.Sprintf("expected '%c', found '%c'\n", c, (*json)[0]) )
}

func expectMsg(json *[]byte, c byte, err string){
    if((*json)[0]!=c){
        print(err)
    }else{
        *json = (*json)[1:]
    }
}

func skip_ws(json *[]byte){
    for 1<len(*json) && ((*json)[0]==' ' || (*json)[0]=='\t' || (*json)[0]=='\n' || (*json)[0]=='\r'){
        *json = (*json)[1:]
    }
}
