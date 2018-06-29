package eni

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
)

type retConverter struct{}

// ConvertReturnValue converts return value from JSON to ENI encoding
func ConvertReturnValue(typeInfo []byte, jsonStr string) (ret []byte, err error) {
	var cvt *retConverter
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Return Parser Error: " + r.(string))
		}
	}()
	json := []byte(jsonStr)
	cvt.skipWS(&json)
	cvt.expect(&json, '[')
	var dataBuf bytes.Buffer
	for i := 0; 0 < len(typeInfo); i++ {
		if 0 < i {
			cvt.skipWS(&json)
			cvt.expect(&json, ',')
		}
		typeInfo, json = cvt.parseType(typeInfo, &dataBuf, []byte(json))
	}
	cvt.skipWS(&json)
	cvt.expect(&json, ']')
	return dataBuf.Bytes(), err
}

// assuming that data are packed
func (cvt *retConverter) parseType(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	t := typeInfo[0]
	if ComplexType[t] {
		if t == FIX_ARRAY_START {
			typeInfo, json = cvt.parseFixArray(typeInfo, data, json)
		} else if t == DYN_ARRAY_START {
			typeInfo, json = cvt.parseDynArray(typeInfo, data, json)
		} else if t == STRUCT_START {
			typeInfo, json = cvt.parseStruct(typeInfo, data, json)
		} else if t == STRING {
			typeInfo, json = cvt.parseString(typeInfo, data, json)
		}
	} else { // value type
		typeInfo, json = cvt.parseValue(typeInfo, data, json)
	}
	return typeInfo, json
}

func (cvt *retConverter) parseString(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	typeInfo = typeInfo[1:] // string
	cvt.skipWS(&json)
	cvt.expect(&json, '"')
	length := int64(0)
	var buf bytes.Buffer
	for json[0] != '"' {
		if json[0] == '\\' {
			buf.WriteByte(cvt.parseEscape(&json))
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
	cvt.expect(&json, '"')
	return typeInfo, json
}

func (cvt *retConverter) parseDynArray(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	// TODO
	panic("dynamic array not implemented yet!")
}

func (cvt *retConverter) parseFixArray(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	typeInfo = typeInfo[1:] // fix_array_start
	cvt.skipWS(&json)
	cvt.expect(&json, '[')
	leng := new(big.Int).SetBytes(typeInfo[:32]).Int64()
	typeInfo = typeInfo[32:]

	for i := int64(0); i < leng; i++ {
		if i > 0 {
			cvt.skipWS(&json)
			cvt.expect(&json, ',')
		}
		if i == leng-1 {
			typeInfo, json = cvt.parseType(typeInfo, data, json)
		} else {
			_, json = cvt.parseType(typeInfo, data, json)
		}
	}

	cvt.skipWS(&json)
	cvt.expect(&json, ']')
	return typeInfo, json
}

func (cvt *retConverter) parseStruct(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	cvt.skipWS(&json)
	cvt.expect(&json, '[')
	for i := 0; ; i++ {
		t := typeInfo[0]
		cvt.skipWS(&json)
		cvt.expect(&json, ',')
		if t != STRUCT_END {
			typeInfo, json = cvt.parseType(typeInfo, data, json)
			break
		}
	}
	typeInfo = typeInfo[1:] // struct_end
	cvt.skipWS(&json)
	cvt.expect(&json, ']')
	return typeInfo, json
}

func (cvt *retConverter) parseValue(typeInfo []byte, data *bytes.Buffer, json []byte) ([]byte, []byte) {
	t := typeInfo[0]
	cvt.skipWS(&json)
	if t == BOOL {
		if cvt.have(&json, 't') {
			cvt.expect(&json, 'r')
			cvt.expect(&json, 'u')
			cvt.expect(&json, 'e')
			data.Write(make([]byte, 31, 31))
			data.WriteByte(byte(1))
		} else if cvt.have(&json, 'f') {
			cvt.expect(&json, 'a')
			cvt.expect(&json, 'l')
			cvt.expect(&json, 's')
			cvt.expect(&json, 'e')
			data.Write(make([]byte, 32, 32))
		} else { // err
			panic(fmt.Sprintf("expected boolean, found '%c'", json[0]))
		}
	} else if IsSint(t) { // signed integer
		i := 0
		ojson := json
		if cvt.have(&json, '-') {
			i++
		}
		for cvt.haveDigit(&json) {
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
	} else if IsUint(t) { // unsigned integer
		i := 0
		ojson := json
		for cvt.haveDigit(&json) {
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
		panic(fmt.Sprintf("encoding error - unknown or not implemented type: %d", t))
	}
	typeInfo = typeInfo[1:]
	return typeInfo, json
}

func (cvt *retConverter) parseEscape(json *[]byte) (ch byte) {
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

func (cvt *retConverter) have(json *[]byte, c byte) bool {
	if (*json)[0] == c {
		*json = (*json)[1:]
		return true
	} else {
		return false
	}
}

func (cvt *retConverter) haveDigit(json *[]byte) bool {
	if len(*json) > 0 && (*json)[0] >= '0' && (*json)[0] <= '9' {
		*json = (*json)[1:]
		return true
	} else {
		return false
	}
}

func (cvt *retConverter) expect(json *[]byte, c byte) {
	cvt.expectMsg(json, c, fmt.Sprintf("expected '%c', found '%c'", c, (*json)[0]))
}

func (cvt *retConverter) expectMsg(json *[]byte, c byte, errMsg string) {
	if (*json)[0] != c {
		panic(errMsg)
	} else {
		*json = (*json)[1:]
	}
}

func (cvt *retConverter) skipWS(json *[]byte) {
	for 1 < len(*json) && ((*json)[0] == ' ' || (*json)[0] == '\t' || (*json)[0] == '\n' || (*json)[0] == '\r') {
		*json = (*json)[1:]
	}
}
