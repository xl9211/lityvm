package ret_parser

import "fmt"
import "bytes"
import "math/big"
import "github.com/ethereum/go-ethereum/common/math"
import "errors"

import "github.com/ethereum/go-ethereum/core/vm/eni/typecodes"

func Parse(type_info []byte, jsonStr string) (ret []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Return Parser Error: " + r.(string))
		}
	}()
	json := []byte(jsonStr)
	skip_ws(&json)
	expect(&json, '[')
	var dataBuf bytes.Buffer
	for i := 0; 0 < len(type_info); i++ {
		if 0 < i {
			skip_ws(&json)
			expect(&json, ',')
		}
		type_info, json = parse_type(type_info, &dataBuf, []byte(json))
	}
	skip_ws(&json)
	expect(&json, ']')
	return dataBuf.Bytes(), err
}

// assuming that data are packed
func parse_type(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	t := type_info[0]
	if typecodes.ComplexType[t] {
		if t == typecodes.FIX_ARRAY_START {
			type_info, json = parse_fix_array(type_info, data, json)
		} else if t == typecodes.DYN_ARRAY_START {
			type_info, json = parse_dyn_array(type_info, data, json)
		} else if t == typecodes.STRUCT_START {
			type_info, json = parse_struct(type_info, data, json)
		} else if t == typecodes.STRING {
			type_info, json = parse_string(type_info, data, json)
		}
	} else { // value type
		type_info, json = parse_value(type_info, data, json)
	}
	return type_info, json
}

func parse_string(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	type_info = type_info[1:] // string
	skip_ws(&json)
	expect(&json, '"')
	length := int64(0)
	var buf bytes.Buffer
	for json[length] != '"' {
		if json[length] == '\\' {
			length++
            buf.WriteByte(parse_escape(json[length:]))
		}else{
            buf.WriteByte(json[length])
        }
		length++
	}

	data.Write(math.PaddedBigBytes(big.NewInt(length), 32))
	data.Write(buf.Bytes())
	if length%32 > 0 {
		data.Write(make([]byte, 32-length))
	}
	json = json[length:]
	expect(&json, '"')
	return type_info, json
}

func parse_dyn_array(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	// TODO
	return type_info, json
}

// parsing int32 not finished
func parse_fix_array(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	type_info = type_info[1:] // fix_array_start
	skip_ws(&json)
	expect(&json, '[')
	leng := new(big.Int).SetBytes(type_info[:32]).Int64()
	type_info = type_info[32:]

	for i := int64(0); i < leng; i++ {
		if i > 0 {
			skip_ws(&json)
			expect(&json, ',')
		}
		if i == leng-1 {
			type_info, json = parse_type(type_info, data, json)
		} else {
			_, json = parse_type(type_info, data, json)
		}
	}

	skip_ws(&json)
	expect(&json, ']')
	return type_info, json
}

func parse_struct(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	skip_ws(&json)
	expect(&json, '[')
	for i := 0; ; i++ {
		t := type_info[0]
		skip_ws(&json)
		expect(&json, ',')
		if t != typecodes.STRUCT_END {
			type_info, json = parse_type(type_info, data, json)
		}
	}
	type_info = type_info[1:] // struct_end
	skip_ws(&json)
	expect(&json, ']')
	return type_info, json
}

func parse_value(type_info []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	t := type_info[0]
	skip_ws(&json)
	if t == typecodes.BOOL {
		if have(&json, 't') {
			expect(&json, 'r')
			expect(&json, 'u')
			expect(&json, 'e')
			data.WriteByte(byte(1))
		} else if have(&json, 'f') {
			expect(&json, 'a')
			expect(&json, 'l')
			expect(&json, 's')
			expect(&json, 'e')
			data.WriteByte(byte(0))
		} else { // err
			panic(fmt.Sprintf("expected boolean, found '%c'", json[0]))
		}
	} else if typecodes.INT <= t && t <= typecodes.INT256 { // signed integer
		i := 0
		ojson := json
		if have(&json, '-') {
			i++
		}
		for have_digit(&json) {
			i++
		}
		if i == 0 || (i == 1 && ojson[0] == '-') {
			panic(fmt.Sprintf("expected int, found '%c'", ojson[0]))
		}
		// two's complement
		n := new(big.Int)
		n.SetString(string(ojson[0:i]), 10)
		b := math.PaddedBigBytes(n, 32)
		for i := 0; i < 32; i++ {
			b[i] ^= byte(255)
		}
		n.SetBytes(b)
		n.Add(n, big.NewInt(int64(1)))
		b = math.PaddedBigBytes(n, 32)
		data.Write(b)
	} else if (typecodes.UINT <= t && t <= typecodes.UINT256) || (typecodes.BYTE1 <= t && t <= typecodes.BYTE32) { // unsigned integer
		i := 0
		ojson := json
		for have_digit(&json) {
			i++
		}
		if i == 0 {
			panic(fmt.Sprintf("expected uint, found '%c'", ojson[0]))
		}
		var n big.Int
		n.SetString(string(ojson[0:i]), 10)
		b := math.PaddedBigBytes(&n, 32)
		data.Write(b)
	} else {
		// TODO: unknown code
	}
	type_info = type_info[1:]
	return type_info, json
}

func parse_escape(json []byte) byte {
    if json[1]=='\\' || json[1]=='"'{
        return json[1]
    }else if json[1]=='/'{
        return '/'
    }else if json[1]=='b'{
        return '\b'
    }else if json[1]=='f'{
        return '\f'
    }else if json[1]=='n'{
        return '\n'
    }else if json[1]=='r'{
        return '\r'
    }else if json[1]=='t'{
        return '\t'
    }else if json[1]=='u'{
        //TODO: UTF-8
        panic("UTF-8 not implemented yet!")
    }else{
        panic("invalid escape sequence")
    }
}
func have(json *[]byte, c byte) bool {
	if (*json)[0] == c {
		*json = (*json)[1:]
		return true
	} else {
		return false
	}
}

func have_digit(json *[]byte) bool {
	if len(*json) > 0 && (*json)[0] >= '0' && (*json)[0] <= '9' {
		*json = (*json)[1:]
		return true
	} else {
		return false
	}
}

func expect(json *[]byte, c byte) {
	expectMsg(json, c, fmt.Sprintf("expected '%c', found '%c'", c, (*json)[0]))
}

func expectMsg(json *[]byte, c byte, errMsg string) {
	if (*json)[0] != c {
		panic(errMsg)
	} else {
		*json = (*json)[1:]
	}
}

func skip_ws(json *[]byte) {
	for 1 < len(*json) && ((*json)[0] == ' ' || (*json)[0] == '\t' || (*json)[0] == '\n' || (*json)[0] == '\r') {
		*json = (*json)[1:]
	}
}
