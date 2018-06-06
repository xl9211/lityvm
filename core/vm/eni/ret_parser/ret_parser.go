package ret_parser

import "fmt"
import "bytes"
import "math/big"
import "github.com/ethereum/go-ethereum/common/math"
import "errors"

import "github.com/ethereum/go-ethereum/core/vm/eni/typecodes"

func Parse(typeInfo []byte, jsonStr string) (ret []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Return Parser Error: " + r.(string))
		}
	}()
	json := []byte(jsonStr)
	skipWS(&json)
	expect(&json, '[')
	var dataBuf bytes.Buffer
	for i := 0; 0 < len(typeInfo); i++ {
		if 0 < i {
			skipWS(&json)
			expect(&json, ',')
		}
		typeInfo, json = parseType(typeInfo, &dataBuf, []byte(json))
	}
	skipWS(&json)
	expect(&json, ']')
	return dataBuf.Bytes(), err
}

// assuming that data are packed
func parseType(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	t := typeInfo[0]
	if typecodes.ComplexType[t] {
		if t == typecodes.FIX_ARRAY_START {
			typeInfo, json = parseFixArray(typeInfo, data, json)
		} else if t == typecodes.DYN_ARRAY_START {
			typeInfo, json = parseDynArray(typeInfo, data, json)
		} else if t == typecodes.STRUCT_START {
			typeInfo, json = parseStruct(typeInfo, data, json)
		} else if t == typecodes.STRING {
			typeInfo, json = parseString(typeInfo, data, json)
		}
	} else { // value type
		typeInfo, json = parseValue(typeInfo, data, json)
	}
	return typeInfo, json
}

func parseString(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	typeInfo = typeInfo[1:] // string
	skipWS(&json)
	expect(&json, '"')
	length := int64(0)
	var buf bytes.Buffer
	for json[0] != '"' {
		if json[0] == '\\' {
			buf.WriteByte(parseEscape(&json))
		} else {
			buf.WriteByte(json[0])
			json = json[1:]
		}
		length++
	}

	data.Write(math.PaddedBigBytes(big.NewInt(length), 32))
	data.Write(buf.Bytes())
	if length%32 > 0 {
		data.Write(make([]byte, 32-length%32))
	}
	expect(&json, '"')
	return typeInfo, json
}

func parseDynArray(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	// TODO
	return typeInfo, json
}

func parseFixArray(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	typeInfo = typeInfo[1:] // fix_array_start
	skipWS(&json)
	expect(&json, '[')
	leng := new(big.Int).SetBytes(typeInfo[:32]).Int64()
	typeInfo = typeInfo[32:]

	for i := int64(0); i < leng; i++ {
		if i > 0 {
			skipWS(&json)
			expect(&json, ',')
		}
		if i == leng-1 {
			typeInfo, json = parseType(typeInfo, data, json)
		} else {
			_, json = parseType(typeInfo, data, json)
		}
	}

	skipWS(&json)
	expect(&json, ']')
	return typeInfo, json
}

func parseStruct(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	skipWS(&json)
	expect(&json, '[')
	for i := 0; ; i++ {
		t := typeInfo[0]
		skipWS(&json)
		expect(&json, ',')
		if t != typecodes.STRUCT_END {
			typeInfo, json = parseType(typeInfo, data, json)
			break
		}
	}
	typeInfo = typeInfo[1:] // struct_end
	skipWS(&json)
	expect(&json, ']')
	return typeInfo, json
}

func parseValue(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	t := typeInfo[0]
	skipWS(&json)
	if t == typecodes.BOOL {
		if have(&json, 't') {
			expect(&json, 'r')
			expect(&json, 'u')
			expect(&json, 'e')
			data.Write(make([]byte, 31, 31))
			data.WriteByte(byte(1))
		} else if have(&json, 'f') {
			expect(&json, 'a')
			expect(&json, 'l')
			expect(&json, 's')
			expect(&json, 'e')
			data.Write(make([]byte, 32, 32))
		} else { // err
			panic(fmt.Sprintf("expected boolean, found '%c'", json[0]))
		}
	} else if typecodes.IsSint(t) { // signed integer
		i := 0
		ojson := json
		if have(&json, '-') {
			i++
		}
		for haveDigit(&json) {
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
	} else if typecodes.IsUint(t) { // unsigned integer
		i := 0
		ojson := json
		for haveDigit(&json) {
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
	typeInfo = typeInfo[1:]
	return typeInfo, json
}

func parseEscape(json *[]byte) (ch byte) {
	if (*json)[1] == '\\' || (*json)[1] == '"' {
		ch = (*json)[1]
	} else if (*json)[1] == '/' {
		ch = '/'
	} else if (*json)[1] == 'b' {
		ch = '\b'
	} else if (*json)[1] == 'f' {
		ch = '\f'
	} else if (*json)[1] == 'n' {
		ch = '\n'
	} else if (*json)[1] == 'r' {
		ch = '\r'
	} else if (*json)[1] == 't' {
		ch = '\t'
	} else if (*json)[1] == 'u' {
		str := string((*json)[2:6])
		var code int
		fmt.Sscanf(str, "%x", &code)
		if code < 128 {
			ch = byte(code)
		} else {
			//TODO: UTF-8
			panic("UTF-8 not implemented yet!")
		}
	} else {
		panic("invalid escape sequence")
	}

	if (*json)[1] == 'u' {
		(*json) = (*json)[6:]
	} else {
		(*json) = (*json)[2:]
	}
	return ch
}

func have(json *[]byte, c byte) bool {
	if (*json)[0] == c {
		*json = (*json)[1:]
		return true
	} else {
		return false
	}
}

func haveDigit(json *[]byte) bool {
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

func skipWS(json *[]byte) {
	for 1 < len(*json) && ((*json)[0] == ' ' || (*json)[0] == '\t' || (*json)[0] == '\n' || (*json)[0] == '\r') {
		*json = (*json)[1:]
	}
}
