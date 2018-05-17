// type format (type_info) grammar
// type_info describe the types with encoding
// type: bool | int | uint | address | bytes | enum | string | fix_array | dyn_array | struct
// fix_array: fix_array_start [0-9]+ type
// dyn_array: dyn_array_start type
// struct: struct_start type+ struct_end

// TODO: bytes
package arg_parser

import "math/big"
import "bytes"
import "fmt"
import "errors"

import "github.com/ethereum/go-ethereum/core/vm/eni/typecodes"

func Parse(type_info []byte, data []byte) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint("Argument Parser Error: ", r))
		}
	}()
	var json bytes.Buffer
	parse_entry_point(type_info, data, &json)
	return json.String(), err
}

func parse_entry_point(type_info []byte, data []byte, json *bytes.Buffer) {
	json.WriteString("[")
	for i := 0; 0 < len(type_info); i++ {
		if 0 < i {
			json.WriteString(",")
		}
		type_info, data = parse_type(type_info, data, json)
	}
	json.WriteString("]")
}

// assuming that data are packed
func parse_type(type_info []byte, data []byte, json *bytes.Buffer) ([]byte, []byte) {
	t := type_info[0]
	if typecodes.ComplexType[t] {
		if t == typecodes.FIX_ARRAY_START {
			type_info, data = parse_fix_array(type_info, data, json)
		} else if t == typecodes.DYN_ARRAY_START {
			type_info, data = parse_dyn_array(type_info, data, json)
		} else if t == typecodes.STRUCT_START {
			type_info, data = parse_struct(type_info, data, json)
		} else if t == typecodes.STRING {
			type_info, data = parse_string(type_info, data, json)
		}
	} else { // value type
		type_info, data = parse_value(type_info, data, json)
	}
	return type_info, data
}

func parse_string(type_info []byte, data []byte, json *bytes.Buffer) ([]byte, []byte) {
	type_info = type_info[1:] // string
	leng := new(big.Int).SetBytes(data[:32]).Int64()
	data = data[32:]

	var buffer bytes.Buffer
	for i := int64(0); i < leng; i++ {
		if data[i] == '\\' || data[i] == '"' {
			buffer.WriteByte('\\')
			buffer.WriteByte(data[i])
		} else if data[i] < 0x20 { // control characters
			buffer.WriteString(fmt.Sprintf("\\u%04X", data[i]))
		} else {
			buffer.WriteByte(data[i])
		}
	}
	json.WriteString("\"")
	json.WriteString(buffer.String())
	json.WriteString("\"")
	data = data[leng:]
	if leng%32 > 0 {
		data = data[32-leng%32:]
	}
	return type_info, data
}

// parsing int32 not finished
func parse_fix_array(type_info []byte, data []byte, json *bytes.Buffer) ([]byte, []byte) {
	type_info = type_info[1:] // fix_array_start
	json.WriteString("[")
	leng := new(big.Int).SetBytes(type_info[:32]).Int64()
	type_info = type_info[32:]

	for i := int64(0); i < leng; i++ {
		if i == leng-1 {
			type_info, data = parse_type(type_info, data, json)
		} else {
			json.WriteString(", ")
			_, data = parse_type(type_info, data, json)
		}
	}

	json.WriteString("]")
	return type_info, data
}

// dynamic array
func parse_dyn_array(type_info []byte, data []byte, json *bytes.Buffer) ([]byte, []byte) {
	panic(fmt.Sprintf("dynamic array not implemented yet!"))
}

func parse_struct(type_info []byte, data []byte, json *bytes.Buffer) ([]byte, []byte) {
	type_info = type_info[1:] // struct_start
	json.WriteString("[")
	for i := 0; 0 < len(type_info); i++ {
		t := type_info[0]
		if 0 < i {
			json.WriteString(", ")
		}
		if t != typecodes.STRUCT_END {
			type_info, data = parse_type(type_info, data, json)
		}
	}
	if type_info[0] != typecodes.STRUCT_START {
		panic("encoding error - expected struct_end token")
	}
	type_info = type_info[1:] // struct_end
	json.WriteString("]")
	return type_info, data
}

// bool, int
func parse_value(type_info []byte, data []byte, json *bytes.Buffer) ([]byte, []byte) {
	t := type_info[0]
	if t == typecodes.BOOL {
		var boolVal bool
		for i := 0; i < 32; i++ {
			if data[i] != 0 {
				boolVal = true
			}
		}
		json.WriteString(fmt.Sprint(boolVal))
	} else if typecodes.IsSint(t) { // signed integer
		n := new(big.Int)
		var b [32]byte
		copy(b[:], data[:32])
		if b[0] >= 128 { // negative value, two's complement
			n.SetBytes(b[:])
			n = n.Sub(n, big.NewInt(int64(1)))
			copy(b[:], n.Bytes())
			for i := 0; i < 32; i++ {
				b[i] ^= 255
			}
			n.SetBytes(b[:])
			n = n.Mul(n, big.NewInt(int64(-1)))
			json.WriteString(n.String())
		} else { // positive value
			n.SetBytes(b[:])
			json.WriteString(n.String())
		}

	} else if typecodes.IsUint(t) { // unsigned integer
		n := new(big.Int)
		n.SetBytes(data[:32]) // big endian
		json.WriteString(n.String())
	} else {
		panic(fmt.Sprintf("encoding error - unknown or not implemented type: %d", t))
	}
	type_info = type_info[1:]
	data = data[32:]
	return type_info, data
}
