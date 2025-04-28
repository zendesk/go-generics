// build +test
package serialize

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/zendesk/go-generics/serialize/internal/test"
)

func validateB64(s string, t *testing.T) {
	_, err := base64.StdEncoding.DecodeString(s)
	test.CheckErr(err, fmt.Sprintf("%s is not valid base64.", s), t)
}

func Test_Serializer_Struct(t *testing.T) {
	type Foo struct {
		Value string
	}

	// Test serialize to and from struct
	from := Foo{Value: "abcd"}

	// Test to and from json string
	toStr, err := NewSerializer[Foo]().FromT(from).ToJsonString()
	test.CheckErr(err, "Failed to go from struct to json string", t)
	toStruct, err := NewSerializer[Foo]().FromJsonString(toStr).ToT()
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(toStruct, "Struct", from, t)

	// Test to and from B64String
	toStr, err = NewSerializer[Foo]().FromT(from).ToB64String()
	test.CheckErr(err, "Failed to go from struct to b64 string string", t)
	validateB64(toStr, t)
	toStruct, err = NewSerializer[Foo]().FromB64String(toStr).ToT()
	test.CheckErr(err, "Failed to go from json b64 string to struct", t)
	test.CheckEqual(toStruct, "Struct", from, t)

	// Test to and from B64
	toBytes, err := NewSerializer[Foo]().FromT(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from struct to b64", t)
	validateB64(string(toBytes), t)
	toStruct, err = NewSerializer[Foo]().FromB64Bytes(toBytes).ToT()
	test.CheckErr(err, "Failed to go from b64 to struct", t)
	test.CheckEqual(toStruct, "Struct", from, t)

	// Test to and from bytes
	toBytes, err = NewSerializer[Foo]().FromT(from).ToBytes()
	test.CheckErr(err, "Failed to go from struct to bytes", t)
	toStruct, err = NewSerializer[Foo]().FromBytes(toBytes).ToT()
	test.CheckErr(err, "Failed to go from bytes to struct", t)
	test.CheckEqual(toStruct, "Struct", from, t)

	// Test to and from stryct
	toStruct, err = NewSerializer[Foo]().FromT(from).ToT()
	test.CheckErr(err, "Failed to go from struct to struct", t)
	test.CheckEqual(toStruct, "Struct", from, t)
}

func Test_Serializer_JSONString(t *testing.T) {
	type Foo struct {
		Value string
	}

	// Test serialize to and from struct
	from := "{\"Value\":\"abcd\"}"

	// Test to and from json string
	toStruct, err := NewSerializer[Foo]().FromJsonString(from).ToT()
	test.CheckErr(err, "Failed to go from json string to json string", t)
	toJson, err := NewSerializer[Foo]().FromT(toStruct).ToJsonString()
	test.CheckErr(err, "Failed to go from json string to json string", t)
	test.CheckEqual(toJson, "json -> struct", from, t)

	// Test to and from B64String
	toStr, err := NewSerializer[any]().FromJsonString(from).ToB64String()
	test.CheckErr(err, "Failed to go from json string to b64 string string", t)
	toJson, err = NewSerializer[any]().FromB64String(toStr).ToJsonString()
	test.CheckErr(err, "Failed to go from json b64 string to json string", t)
	test.CheckEqual(toJson, "json -> b64str", from, t)

	// Test to and from B64
	toBytes, err := NewSerializer[any]().FromJsonString(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json string to b64", t)
	validateB64(string(toBytes), t)
	toJson, err = NewSerializer[any]().FromB64Bytes(toBytes).ToJsonString()
	test.CheckErr(err, "Failed to go from b64 to json string", t)
	test.CheckEqual(toJson, "json -> b64", from, t)

	//Test to and from bytes
	toBytes, err = NewSerializer[Foo]().FromJsonString(from).ToBytes()
	test.CheckErr(err, "Failed to go from json string to bytes", t)
	toJson, err = NewSerializer[Foo]().FromBytes(toBytes).ToJsonString()
	test.CheckErr(err, "Failed to go from bytes to json string", t)
	test.CheckEqual(toJson, "Struct", from, t)

	// Test to and from struct
	toJson, err = NewSerializer[Foo]().FromJsonString(from).ToJsonString()
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(toJson, "Json str", from, t)
}

func Test_Serializer_B64String(t *testing.T) {
	type Foo struct {
		Value string
	}

	foo := Foo{Value: "foo1234"}
	from, err := NewSerializer[Foo]().FromT(foo).ToB64String()
	test.CheckErr(err, "Failed to init from", t)

	// Test serialize to and from b64String

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromB64String(from).ToT()
	test.CheckErr(err, "Failed to go from b64str string to json string", t)
	result, err := NewSerializer[Foo]().FromT(toStruct).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to b64str", t)
	test.CheckEqual(result, "b64str -> struct", from, t)

	// Test to and from json
	toStr, err := NewSerializer[any]().FromB64String(from).ToJsonString()
	test.CheckErr(err, "Failed to go from b64str string to b64 string string", t)
	result, err = NewSerializer[any]().FromJsonString(toStr).ToB64String()
	test.CheckErr(err, "Failed to go from b64str b64 string to b64str", t)
	test.CheckEqual(result, "b64str -> json", from, t)

	// Test to and from B64
	toBytes, err := NewSerializer[any]().FromB64String(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	result, err = NewSerializer[any]().FromB64Bytes(toBytes).ToB64String()
	test.CheckErr(err, "Failed to go from b64 to b64str", t)
	test.CheckEqual(result, "json -> b64", from, t)

	//Test to and from bytes
	toBytes, err = NewSerializer[any]().FromB64String(from).ToBytes()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	result, err = NewSerializer[any]().FromBytes(toBytes).ToB64String()
	test.CheckErr(err, "Failed to go from bytes to b64str", t)
	test.CheckEqual(result, "b64str -> bytes", from, t)

	// Test to and from b64
	result, err = NewSerializer[any]().FromB64String(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to b64String", t)
	test.CheckEqual(result, "b64 str", from, t)
}

func Test_Serializer_B64Bytes(t *testing.T) {
	type Foo struct {
		Value string
	}

	foo := Foo{Value: "foo1234"}
	from, err := NewSerializer[Foo]().FromT(foo).ToB64Bytes()
	test.CheckErr(err, "Failed to init from", t)

	// Test serialize to and from b64String

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromB64Bytes(from).ToT()
	test.CheckErr(err, "Failed to go from b64str string to json string", t)
	result, err := NewSerializer[Foo]().FromT(toStruct).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64str", t)
	test.CheckEqual(result, "b64str -> struct", from, t)

	// Test to and from json
	toStr, err := NewSerializer[any]().FromB64Bytes(from).ToJsonString()
	test.CheckErr(err, "Failed to go from b64str string to b64 string string", t)
	result, err = NewSerializer[any]().FromJsonString(toStr).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str b64 string to b64str", t)
	test.CheckEqual(result, "b64str -> json", from, t)

	// Test to and from B64
	toStr, err = NewSerializer[any]().FromB64Bytes(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	result, err = NewSerializer[any]().FromB64String(toStr).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64 to b64str", t)
	test.CheckEqual(result, "json -> b64", from, t)

	//Test to and from bytes
	toBytes, err := NewSerializer[any]().FromB64Bytes(from).ToBytes()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	result, err = NewSerializer[any]().FromBytes(toBytes).ToB64Bytes()
	test.CheckErr(err, "Failed to go from bytes to b64str", t)
	test.CheckEqual(result, "b64 -> bytes", from, t)

	// Test to and from b64
	result, err = NewSerializer[any]().FromB64Bytes(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64String", t)
	test.CheckEqual(result, "b64", from, t)
}

func Test_Serializer_Bytes(t *testing.T) {
	type Foo struct {
		Value string
	}

	foo := Foo{Value: "foo1234"}
	from, err := NewSerializer[Foo]().FromT(foo).ToBytes()
	test.CheckErr(err, "Failed to init from", t)

	// Test serialize to and from b64String

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromBytes(from).ToT()
	test.CheckErr(err, "Failed to go from b64str string to json string", t)
	result, err := NewSerializer[Foo]().FromT(toStruct).ToBytes()
	test.CheckErr(err, "Failed to go from b64str string to b64str", t)
	test.CheckEqual(result, "bytes -> struct", from, t)

	// Test to and from json
	toStr, err := NewSerializer[any]().FromBytes(from).ToJsonString()
	test.CheckErr(err, "Failed to go from b64str string to b64 string string", t)
	result, err = NewSerializer[any]().FromJsonString(toStr).ToBytes()
	test.CheckErr(err, "Failed to go from b64str b64 string to b64str", t)
	test.CheckEqual(result, "bytes -> json", from, t)

	// Test to and from B64
	toBytes, err := NewSerializer[any]().FromBytes(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	result, err = NewSerializer[any]().FromB64Bytes(toBytes).ToBytes()
	test.CheckErr(err, "Failed to go from b64 to b64str", t)
	test.CheckEqual(result, "bytes -> b64", from, t)

	//Test to and from bytes
	toStr, err = NewSerializer[any]().FromBytes(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	result, err = NewSerializer[any]().FromB64String(toStr).ToBytes()
	test.CheckErr(err, "Failed to go from bytes to b64str", t)
	test.CheckEqual(result, "bytes -> b64str", from, t)

	// Test to and from b64
	result, err = NewSerializer[any]().FromBytes(from).ToBytes()
	test.CheckErr(err, "Failed to go from b64str string to b64String", t)
	test.CheckEqual(result, "bytes", from, t)
}

func Test_Serializer_Slice_String(t *testing.T) {
	// Test serialize to and from self
	from := []string{"a", "b", "cde"}
	to, err := NewSerializer[[]string]().FromT(from).ToT()
	test.CheckErr(err, "Failed to init from", t)
	test.CheckEqual(to, "From -> To", from, t)

	// Test to and from bytes
	toBytes, err := NewSerializer[[]string]().FromT(from).ToBytes()
	test.CheckErr(err, "Failed to go from slice to bytes", t)
	toSlice, err := NewSerializer[[]string]().FromBytes(toBytes).ToT()
	test.CheckErr(err, "Failed to go from bytes to slice", t)
	test.CheckEqual(toSlice, "slice -> bytes", from, t)

	// Test to and from json
	toJsonStr, err := NewSerializer[[]string]().FromT(from).ToJsonString()
	test.CheckErr(err, "Failed to go from slice string to json string", t)
	toSlice, err = NewSerializer[[]string]().FromJsonString(toJsonStr).ToT()
	test.CheckErr(err, "Failed to go from jsonStr to slice", t)
	test.CheckEqual(toSlice, "slice -> Json", from, t)

	// Test to and from B64
	toB64Bytes, err := NewSerializer[[]string]().FromT(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	toSlice, err = NewSerializer[[]string]().FromB64Bytes(toB64Bytes).ToT()
	test.CheckErr(err, "Failed to go from b64 to slice", t)
	test.CheckEqual(toSlice, "slice -> b64", from, t)

	//Test to and from b64String
	toB64Str, err := NewSerializer[[]string]().FromT(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	toSlice, err = NewSerializer[[]string]().FromB64String(toB64Str).ToT()
	test.CheckErr(err, "Failed to go from b64str to slice", t)
	test.CheckEqual(toSlice, "slice -> b64str", from, t)
}

func Test_Serializer_Slice_Struct(t *testing.T) {
	type Foo struct {
		Value string
	}
	// Test serialize to and from self
	from := []Foo{{Value: "one"}, {Value: "two"}, {Value: "three"}}
	to, err := NewSerializer[[]Foo]().FromT(from).ToT()
	test.CheckErr(err, "Failed to init from", t)
	test.CheckEqual(to, "From -> To", from, t)

	// Test to and from bytes
	toBytes, err := NewSerializer[[]Foo]().FromT(from).ToBytes()
	test.CheckErr(err, "Failed to go from slice to bytes", t)
	toSlice, err := NewSerializer[[]Foo]().FromBytes(toBytes).ToT()
	test.CheckErr(err, "Failed to go from bytes to slice", t)
	test.CheckEqual(toSlice, "slice -> bytes", from, t)

	// Test to and from json
	toJsonStr, err := NewSerializer[[]Foo]().FromT(from).ToJsonString()
	test.CheckErr(err, "Failed to go from slice string to json string", t)
	toSlice, err = NewSerializer[[]Foo]().FromJsonString(toJsonStr).ToT()
	test.CheckErr(err, "Failed to go from jsonStr to slice", t)
	test.CheckEqual(toSlice, "slice -> Json", from, t)

	// Test to and from B64
	toB64Bytes, err := NewSerializer[[]Foo]().FromT(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	toSlice, err = NewSerializer[[]Foo]().FromB64Bytes(toB64Bytes).ToT()
	test.CheckErr(err, "Failed to go from b64 to slice", t)
	test.CheckEqual(toSlice, "slice -> b64", from, t)

	//Test to and from b64String
	toB64Str, err := NewSerializer[[]Foo]().FromT(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	toSlice, err = NewSerializer[[]Foo]().FromB64String(toB64Str).ToT()
	test.CheckErr(err, "Failed to go from b64str to slice", t)
	test.CheckEqual(toSlice, "slice -> b64str", from, t)
}

func Test_Serializer_Slice_PointerToT(t *testing.T) {
	type Foo struct {
		Value string
	}
	// Test serialize to and from self
	from := []*Foo{{Value: "one"}, {Value: "two"}, {Value: "three"}}
	to, err := NewSerializer[[]*Foo]().FromT(from).ToT()
	test.CheckErr(err, "Failed to init from", t)
	test.CheckEqual(to, "From -> To", from, t)

	// Test to and from bytes
	toBytes, err := NewSerializer[[]*Foo]().FromT(from).ToBytes()
	test.CheckErr(err, "Failed to go from slice to bytes", t)
	toSlice, err := NewSerializer[[]*Foo]().FromBytes(toBytes).ToT()
	test.CheckErr(err, "Failed to go from bytes to slice", t)
	test.CheckEqual(toSlice, "slice -> bytes", from, t)

	// Test to and from json
	toJsonStr, err := NewSerializer[[]*Foo]().FromT(from).ToJsonString()
	test.CheckErr(err, "Failed to go from slice string to json string", t)
	toSlice, err = NewSerializer[[]*Foo]().FromJsonString(toJsonStr).ToT()
	test.CheckErr(err, "Failed to go from jsonStr to slice", t)
	test.CheckEqual(toSlice, "slice -> Json", from, t)

	// Test to and from B64
	toB64Bytes, err := NewSerializer[[]*Foo]().FromT(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	toSlice, err = NewSerializer[[]*Foo]().FromB64Bytes(toB64Bytes).ToT()
	test.CheckErr(err, "Failed to go from b64 to slice", t)
	test.CheckEqual(toSlice, "slice -> b64", from, t)

	//Test to and from b64String
	toB64Str, err := NewSerializer[[]*Foo]().FromT(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	toSlice, err = NewSerializer[[]*Foo]().FromB64String(toB64Str).ToT()
	test.CheckErr(err, "Failed to go from b64str to slice", t)
	test.CheckEqual(toSlice, "slice -> b64str", from, t)
}

func Test_Serializer_Slice_Bytes(t *testing.T) {
	// Test serialize to and from self
	from := [][]byte{{1, 2}, {3, 4}, {5, 6}}
	to, err := NewSerializer[[][]byte]().FromT(from).ToT()
	test.CheckErr(err, "Failed to init from", t)
	test.CheckEqual(to, "From -> To", from, t)

	// Test to and from bytes
	toBytes, err := NewSerializer[[][]byte]().FromT(from).ToBytes()
	test.CheckErr(err, "Failed to go from slice to bytes", t)
	toSlice, err := NewSerializer[[][]byte]().FromBytes(toBytes).ToT()
	test.CheckErr(err, "Failed to go from bytes to slice", t)
	test.CheckEqual(toSlice, "slice -> bytes", from, t)

	// Test to and from json
	toJsonStr, err := NewSerializer[[][]byte]().FromT(from).ToJsonString()
	test.CheckErr(err, "Failed to go from slice string to json string", t)
	toSlice, err = NewSerializer[[][]byte]().FromJsonString(toJsonStr).ToT()
	test.CheckErr(err, "Failed to go from jsonStr to slice", t)
	test.CheckEqual(toSlice, "slice -> Json", from, t)

	// Test to and from B64
	toB64Bytes, err := NewSerializer[[][]byte]().FromT(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64str string to b64", t)
	toSlice, err = NewSerializer[[][]byte]().FromB64Bytes(toB64Bytes).ToT()
	test.CheckErr(err, "Failed to go from b64 to slice", t)
	test.CheckEqual(toSlice, "slice -> b64", from, t)

	//Test to and from b64String
	toB64Str, err := NewSerializer[[][]byte]().FromT(from).ToB64String()
	test.CheckErr(err, "Failed to go from b64str string to bytes", t)
	toSlice, err = NewSerializer[[][]byte]().FromB64String(toB64Str).ToT()
	test.CheckErr(err, "Failed to go from b64str to slice", t)
	test.CheckEqual(toSlice, "slice -> b64str", from, t)
}

func Test_Serializer_InferredType_Struct(t *testing.T) {
	type Foo struct {
		Value string
	}

	// Test serialize to and from struct
	from := Foo{Value: "abcd"}
	var typ Foo

	// Test to and from json string
	toStr, err := NewSerializer[Foo]().FromDynamicType(from).ToJsonString()
	test.CheckErr(err, "Failed to go from struct to json string", t)
	toStruct, err := NewSerializer[Foo]().FromDynamicType(toStr).ToDynamicType(Struct, typ)
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(toStruct, "struct -> json", from, t)

	// Test to and from B64String
	toStr, err = NewSerializer[Foo]().FromDynamicType(from).ToB64String()
	test.CheckErr(err, "Failed to go from struct to b64 string string", t)
	toStruct, err = NewSerializer[Foo]().FromB64String(toStr).ToDynamicType(Struct, typ)
	test.CheckErr(err, "Failed to go from json b64 string to struct", t)
	test.CheckEqual(toStruct, "struct -> b64str", from, t)

	// Test to and from B64
	toBytes, err := NewSerializer[Foo]().FromDynamicType(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from struct to b64", t)
	toStruct, err = NewSerializer[Foo]().FromB64Bytes(toBytes).ToDynamicType(Struct, typ)
	test.CheckErr(err, "Failed to go from b64 to struct", t)
	test.CheckEqual(toStruct, "struct -> b64b", from, t)

	// Test to and from bytes
	toBytes, err = NewSerializer[Foo]().FromDynamicType(from).ToBytes()
	test.CheckErr(err, "Failed to go from struct to bytes", t)
	toStruct, err = NewSerializer[Foo]().FromDynamicType(toBytes).ToDynamicType(Struct, typ)
	test.CheckErr(err, "Failed to go from bytes to struct", t)
	test.CheckEqual(toStruct, "struct -> bytes", from, t)

	// Test to and from stryct
	toStruct, err = NewSerializer[Foo]().FromDynamicType(from).ToDynamicType(Struct, typ)
	test.CheckErr(err, "Failed to go from struct to struct", t)
	test.CheckEqual(toStruct, "struct -> struct", from, t)
}

func Test_Serializer_InferredType_JSONString(t *testing.T) {
	type Foo struct {
		Value string
	}

	// Test serialize to and from struct
	from := "{\"Value\":\"abcd\"}"

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromDynamicType(from).ToT()
	test.CheckErr(err, "Failed to go from json string to json string", t)
	toJson, err := NewSerializer[Foo]().FromDynamicType(toStruct).ToJsonString()
	test.CheckErr(err, "Failed to go from json string to json string", t)
	test.CheckEqual(toJson, "json -> struct", from, t)

	// Test to and from B64String
	toStr, err := NewSerializer[string]().FromDynamicType(from).ToB64String()
	test.CheckErr(err, "Failed to go from json string to b64 string string", t)
	result, err := NewSerializer[string]().FromB64String(toStr).ToDynamicType(JSONString, nil)
	toJson = result.(string)
	test.CheckErr(err, "Failed to go from json b64 string to json string", t)
	test.CheckEqual(toJson, "json -> b64str", from, t)

	// Test to and from B64
	b64Bytes, err := NewSerializer[string]().FromDynamicType(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json string to b64", t)
	result, err = NewSerializer[string]().FromB64Bytes(b64Bytes).ToDynamicType(JSONString, nil)
	toJson = result.(string)
	test.CheckErr(err, "Failed to go from b64 to json string", t)
	test.CheckEqual(toJson, "json -> b64", from, t)

	//Test to and from bytes
	bytes, err := NewSerializer[string]().FromDynamicType(from).ToBytes()
	test.CheckErr(err, "Failed to go from json string to bytes", t)
	result, err = NewSerializer[string]().FromDynamicType(bytes).ToDynamicType(JSONString, nil)
	toJson = result.(string)
	test.CheckErr(err, "Failed to go from bytes to json string", t)
	test.CheckEqual(toJson, "json -> bytes", from, t)

	// Test to and from json
	result, err = NewSerializer[string]().FromDynamicType(from).ToDynamicType(JSONString, nil)
	toJson = result.(string)
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(toJson, "json -> json", from, t)
}

func Test_Serializer_InferredType_B64String(t *testing.T) {
	type Foo struct {
		Value string
	}

	foo := Foo{Value: "foo1234"}
	from, err := NewSerializer[Foo]().FromT(foo).ToB64String()
	test.CheckErr(err, "Failed to init from", t)

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromB64String(from).ToT()
	test.CheckErr(err, "Failed to go 1", t)
	toStructResult, err := NewSerializer[Foo]().FromDynamicType(toStruct).ToB64String()
	test.CheckErr(err, "Failed to go 2", t)
	test.CheckEqual(toStructResult, "b64str -> struct", from, t)

	// Test to and from Json string
	toStr, err := NewSerializer[string]().FromB64String(from).ToJsonString()
	test.CheckErr(err, "Failed to go from json string to b64 string string", t)
	b64result, err := NewSerializer[string]().FromDynamicType(toStr).ToB64String()
	test.CheckErr(err, "Failed to go from json b64 string to json string", t)
	test.CheckEqual(b64result, "b64str -> json", from, t)

	// Test to and from B64
	b64Bytes, err := NewSerializer[string]().FromB64String(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json string to b64", t)
	b64StrResult, err := NewSerializer[string]().FromB64Bytes(b64Bytes).ToB64String()
	test.CheckErr(err, "Failed to go from b64 to json string", t)
	test.CheckEqual(b64StrResult, "b64str -> b64", from, t)

	//Test to and from bytes
	bytes, err := NewSerializer[string]().FromB64String(from).ToBytes()
	test.CheckErr(err, "Failed to go from json string to bytes", t)
	b64StrResult, err = NewSerializer[string]().FromDynamicType(bytes).ToB64String()
	test.CheckErr(err, "Failed to go from bytes to json string", t)
	test.CheckEqual(b64StrResult, "b64str -> bytes", from, t)

	// Test to and from b64str
	b64StrResult, err = NewSerializer[string]().FromB64String(from).ToB64String()
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(b64StrResult, "b64str -> b64str", from, t)
}

func Test_Serializer_InferredType_B64Bytes(t *testing.T) {
	type Foo struct {
		Value string
	}

	foo := Foo{Value: "foo1234"}
	from, err := NewSerializer[Foo]().FromT(foo).ToB64Bytes()
	test.CheckErr(err, "Failed to init from", t)

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromB64Bytes(from).ToT()
	test.CheckErr(err, "Failed to go from json inferred to struct", t)
	tob64b, err := NewSerializer[Foo]().FromDynamicType(toStruct).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json string to json string", t)
	test.CheckEqual(tob64b, "b64 -> struct", from, t)

	// Test to and from Json string
	toStr, err := NewSerializer[[]byte]().FromB64Bytes(from).ToJsonString()
	test.CheckErr(err, "Failed to go from inferred to json string", t)
	b64result, err := NewSerializer[[]byte]().FromDynamicType(toStr).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json b64 string to json string", t)
	test.CheckEqual(b64result, "b64 -> json", from, t)

	// Test to and from B64
	b64Bytes, err := NewSerializer[[]byte]().FromB64Bytes(from).ToB64String()
	test.CheckErr(err, "Failed to go from json string to b64", t)
	result, err := NewSerializer[[]byte]().FromB64String(b64Bytes).ToB64Bytes()
	test.CheckErr(err, "Failed to go from b64 to json string", t)
	test.CheckEqual(result, "b64 -> b64str", from, t)

	// Test to and from bytes
	bytes, err := NewSerializer[[]byte]().FromB64Bytes(from).ToBytes()
	test.CheckErr(err, "Failed to go from json string to bytes", t)
	anyResult, err := NewSerializer[[]byte]().FromDynamicType(bytes).ToDynamicType(B64Bytes, nil) // Inferred should be b64
	converted := anyResult.([]byte)
	test.CheckErr(err, "Failed to go from bytes to json string", t)
	test.CheckEqual(converted, "b64 -> bytes", from, t)

	// Test to and from b64b
	result, err = NewSerializer[[]byte]().FromB64Bytes(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(result, "b64 -> b64", from, t)
}

func Test_Serializer_InferredType_Bytes(t *testing.T) {
	type Foo struct {
		Value string
	}

	foo := Foo{Value: "foo1234"}
	from, err := NewSerializer[Foo]().FromT(foo).ToBytes()
	test.CheckErr(err, "Failed to init from", t)

	// Test to and from struct
	toStruct, err := NewSerializer[Foo]().FromDynamicType(from).ToT()
	test.CheckErr(err, "Failed to go from json inferred to struct", t)
	tob64b, err := NewSerializer[Foo]().FromDynamicType(toStruct).ToBytes()
	test.CheckErr(err, "Failed to go from json string to json string", t)
	test.CheckEqual(tob64b, "bytes -> struct", from, t)

	// Test to and from Json string
	toStr, err := NewSerializer[[]byte]().FromDynamicType(from).ToJsonString()
	test.CheckErr(err, "Failed to go from inferred to json string", t)
	b64result, err := NewSerializer[[]byte]().FromDynamicType(toStr).ToBytes()
	test.CheckErr(err, "Failed to go from json b64 string to json string", t)
	test.CheckEqual(b64result, "bytes -> json", from, t)

	// Test to and from B64
	b64str, err := NewSerializer[[]byte]().FromDynamicType(from).ToB64String()
	test.CheckErr(err, "Failed to go from json string to b64", t)
	result, err := NewSerializer[[]byte]().FromB64String(b64str).ToDynamicType(Bytes, nil)
	bConverted := result.([]byte)
	test.CheckErr(err, "Failed to go from b64 to b64str string", t)
	test.CheckEqual(bConverted, "bytes -> b64str", from, t)

	// Test to and from B64Bytes
	b64bytes, err := NewSerializer[[]byte]().FromDynamicType(from).ToB64Bytes()
	test.CheckErr(err, "Failed to go from json string to bytes", t)
	result, err = NewSerializer[[]byte]().FromB64Bytes(b64bytes).ToBytes()
	converted := result.([]byte)
	t.Logf("%s : %s", string(converted), string(from))
	test.CheckErr(err, "Failed to go from bytes to json string", t)
	test.CheckEqual(converted, "bytes -> b64bytes", from, t)

	// Test to and from b64b
	result, err = NewSerializer[[]byte]().FromDynamicType(from).ToBytes()
	converted = result.([]byte)
	test.CheckErr(err, "Failed to go from bytes", t)
	test.CheckEqual(result, "bytes -> bytes", from, t)
}

func Test_Serializer_DynamicType_Reflect(t *testing.T) {
	type Foo struct {
		Value string
	}

	// Test serialize to and from struct
	from := Foo{Value: "abcd"}
	var typ Foo

	// Test reflecting to struct
	toStr, err := NewSerializer[Foo]().FromDynamicType(from).ToJsonString()
	test.CheckErr(err, "Failed to go from struct to json string", t)
	t.Logf("Type: %+v", typ)
	toStruct, err := NewSerializer[Foo]().FromDynamicType(toStr).ToDynamicType(Reflect, typ)
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(toStruct, "struct -> json", from, t)

	// test reflecting to string
	fromBytes, err := NewSerializer[string]().FromString("My string").ToBytes()
	test.CheckErr(err, "Failed to convert to bytes", t)
	anyResult, err := NewSerializer[string]().FromDynamicType(fromBytes).ToDynamicType(Reflect, "string")
	strResult := anyResult.(string)
	test.CheckErr(err, "error converting dynamic type (1)", t)
	test.CheckEqual(strResult, "bytes -> dynamic string", string(fromBytes), t)

	// test reflecting to bytes
	fromStr := "from string value"
	var bytes []byte
	test.CheckErr(err, "Failed to convert to bytes", t)
	anyBytrResult, err := NewSerializer[string]().FromDynamicType(fromStr).ToDynamicType(Reflect, bytes)
	bytes = anyBytrResult.([]byte)
	test.CheckErr(err, "err converting dynamic type (2)", t)
	test.CheckEqual(bytes, "string -> dynamic bytes", []byte(fromStr), t)
}

func Test_Serializer_DynamicType_PointerSlice_Reflect(t *testing.T) {
	type Foo struct {
		Value string
	}

	// Test serialize to and from struct
	from := []*Foo{
		{Value: "abcd"},
		{Value: "1234"},
		{Value: "zzzz"},
	}

	var typ []*Foo

	// Test reflecting to struct
	toStr, err := NewSerializer[[]*Foo]().FromDynamicType(from).ToJsonString()
	test.CheckErr(err, "Failed to go from struct to json string", t)
	t.Logf("ToStr: %+v", toStr)
	toStruct, err := NewSerializer[[]*Foo]().FromDynamicType(toStr).ToDynamicType(Reflect, typ)
	test.CheckErr(err, "Failed to go from json string to struct", t)
	test.CheckEqual(toStruct, "struct -> json", from, t)

	// test reflecting to string
	fromBytes, err := NewSerializer[string]().FromString("My string").ToBytes()
	test.CheckErr(err, "Failed to convert to bytes", t)
	anyResult, err := NewSerializer[string]().FromDynamicType(fromBytes).ToDynamicType(Reflect, "string")
	strResult := anyResult.(string)
	test.CheckErr(err, "error converting dynamic type (1)", t)
	test.CheckEqual(strResult, "bytes -> dynamic string", string(fromBytes), t)

	// test reflecting to bytes
	fromStr := "from string value"
	var bytes []byte
	test.CheckErr(err, "Failed to convert to bytes", t)
	anyBytrResult, err := NewSerializer[string]().FromDynamicType(fromStr).ToDynamicType(Reflect, bytes)
	bytes = anyBytrResult.([]byte)
	test.CheckErr(err, "err converting dynamic type (2)", t)
	test.CheckEqual(bytes, "string -> dynamic bytes", []byte(fromStr), t)
}
