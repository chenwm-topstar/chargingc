package slices

import (
	"reflect"
	"strconv"
	"unsafe"
)

// StrToBytes will return the same array in given string.
// Modifying the byte array will directly change underlying memory.
// NOTE: Should be used only to avoid copy for large string.
func StrToBytes(str string) []byte {
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&str))
	bytesHeader := &reflect.SliceHeader{
		Data: stringHeader.Data,
		Len:  stringHeader.Len,
		Cap:  stringHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bytesHeader))
}

// BytesToStr will return string using the byte array as underlying memory.
// Modifying the wrapped byte array will change the string content.
// NOTE: Should be used only to avoid copy for large string.
func BytesToStr(bytes []byte) string {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := &reflect.StringHeader{
		Data: bytesHeader.Data,
		Len:  bytesHeader.Len,
	}
	return *(*string)(unsafe.Pointer(stringHeader))
}

func Int32ToStrSlice(arr []int32) []string {
	s := make([]string, len(arr))
	for i, n := range arr {
		s[i] = strconv.Itoa(int(n))
	}
	return s
}

func Int64ToStrSlice(arr []int64) []string {
	s := make([]string, len(arr))
	for i, n := range arr {
		s[i] = strconv.FormatInt(n, 10)
	}
	return s
}

// LastString returns the last element in a slice.
func LastString(strSlice []string) string {
	if len(strSlice) == 0 {
		return ""
	}
	return strSlice[len(strSlice)-1]
}

// Clone copies the underlying array of given `strSlice` to a new slice.
func Clone(strSlice []string) []string {
	if strSlice == nil {
		return strSlice
	}
	return append(strSlice[:0:0], strSlice...)
}

// Repeat returns a slice with `n` repeated same `element` string.
func Repeat(n int, element string) []string {
	if n <= 0 {
		return nil
	}
	ret := make([]string, n)
	for i := 0; i < n; i++ {
		ret[i] = element
	}
	return ret
}

func Str2InterfaceSlice(arr []string) []interface{} {
	ret := make([]interface{}, len(arr))
	for i := range arr {
		ret[i] = arr[i]
	}
	return ret
}
