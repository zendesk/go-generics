package serialize

import (
	"math/rand"
	"testing"

	"github.com/zendesk/go-generics/internal/test"
)

func Fuzz_Dynamic_Bytes(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{22})
	f.Add([]byte{22, 21, 20})

	f.Fuzz(func(t *testing.T, value []byte) {

		// To the same type
		result1, err := NewSerializer[[]byte]().FromDynamicType(value).ToBytes()
		test.CheckErr(err, "result 1 error", t)
		test.CheckEqual(result1, "result 1", value, t)

		result2, err := NewSerializer[[]byte]().FromDynamicType(result1).ToDynamicType(Reflect, value)
		test.CheckErr(err, "result 2 error", t)
		test.CheckEqual(result2, "result 2", value, t)
		test.CheckEqual(result2, "result 2 (2)", result1, t)

		// To json
		jsonResult, err := NewSerializer[[]byte]().FromDynamicType(value).ToJsonString()
		test.CheckErr(err, "(JSON) result error", t)

		jsonToBytes, err := NewSerializer[[]byte]().FromDynamicType(jsonResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(JSON) jsonToBytes error", t)
		test.CheckEqual(jsonToBytes, "JsonToBytes", value, t)

		// To bytes
		byteResult, err := NewSerializer[[]byte]().FromDynamicType(value).ToBytes()
		test.CheckErr(err, "(byteResult) result error", t)

		bytes, err := NewSerializer[[]byte]().FromDynamicType(byteResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(byteResult) b64Result error", t)
		test.CheckEqual(bytes, "byteResult", value, t)

		// To string
		stringResult, err := NewSerializer[[]byte]().FromDynamicType(value).ToString()
		test.CheckErr(err, "(stringResult) result error", t)

		stringToBytes, err := NewSerializer[[]byte]().FromDynamicType(stringResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(stringResult) stringResult error", t)
		test.CheckEqual(stringToBytes, "stringResult", value, t)
	})
}

func Fuzz_Dynamic_Struct(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{22})
	f.Add([]byte{22, 21, 20})

	f.Fuzz(func(t *testing.T, valBytes []byte) {
		type Foo struct {
			Value []byte
		}

		value := Foo{Value: valBytes}

		// To the same type
		result1, err := NewSerializer[Foo]().FromDynamicType(value).ToT()
		test.CheckErr(err, "result 1 error", t)
		test.CheckEqual(result1, "result 1", value, t)

		result2, err := NewSerializer[Foo]().FromDynamicType(result1).ToDynamicType(Reflect, value)
		test.CheckErr(err, "result 2 error", t)
		test.CheckEqual(result2, "result 2", value, t)
		test.CheckEqual(result2, "result 2 (2)", result1, t)

		// To json
		jsonResult, err := NewSerializer[Foo]().FromDynamicType(value).ToJsonString()
		test.CheckErr(err, "(JSON) result error", t)

		jsonToT, err := NewSerializer[Foo]().FromDynamicType(jsonResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(JSON) jsonToT error", t)
		test.CheckEqual(jsonToT, "jsonToT", value, t)

		// To bytes
		byteResult, err := NewSerializer[Foo]().FromDynamicType(value).ToBytes()
		test.CheckErr(err, "(byteResult) result error", t)

		bytes, err := NewSerializer[Foo]().FromDynamicType(byteResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(byteResult) byteResult error", t)
		test.CheckEqualEquateEmpty(bytes, "byteResult", value, t)

		// To b64
		b64Result, err := NewSerializer[Foo]().FromDynamicType(value).ToB64Bytes()
		test.CheckErr(err, "(b64Bytes) result error", t)

		b64Bytes, err := NewSerializer[Foo]().FromB64Bytes(b64Result).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(b64Bytes) b64Bytes error", t)
		test.CheckEqualEquateEmpty(b64Bytes, "b64Bytes", value, t)

		// To string
		stringResult, err := NewSerializer[Foo]().FromDynamicType(value).ToString()
		test.CheckErr(err, "(stringResult) result error", t)

		stringToT, err := NewSerializer[Foo]().FromDynamicType(stringResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(stringResult) stringResult error", t)
		test.CheckEqual(stringToT, "stringResult", value, t)

	})
}

func Fuzz_Dynamic_String(f *testing.F) {
	f.Add("")
	f.Add("")
	f.Add("")

	f.Fuzz(func(t *testing.T, value string) {
		// To the same type
		result1, err := NewSerializer[string]().FromDynamicType(value).ToString()
		test.CheckErr(err, "result 1 error", t)
		test.CheckEqual(result1, "result 1", value, t)

		result2, err := NewSerializer[string]().FromDynamicType(result1).ToDynamicType(Reflect, value)
		test.CheckErr(err, "result 2 error", t)
		test.CheckEqual(result2, "result 2", value, t)
		test.CheckEqual(result2, "result 2 (2)", result1, t)

		// To json
		jsonResult, err := NewSerializer[string]().FromDynamicType(value).ToJsonString()
		test.CheckErr(err, "(JSON) result error", t)

		fromJson, err := NewSerializer[string]().FromDynamicType(jsonResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(JSON) jsonToBytes error", t)
		test.CheckEqual(fromJson, "JsonToBytes", value, t)

		// To bytes
		byteResult, err := NewSerializer[string]().FromDynamicType(value).ToBytes()
		test.CheckErr(err, "(byteResult) result error", t)

		fromBytes, err := NewSerializer[string]().FromDynamicType(byteResult).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(byteResult) byteResult error", t)
		test.CheckEqualEquateEmpty(fromBytes, "byteResult", value, t)

		// To b64
		b64Result, err := NewSerializer[string]().FromDynamicType(value).ToB64Bytes()
		test.CheckErr(err, "(b64Bytes) result error", t)

		t.Logf("RESULT: %+v", string(b64Result))

		b64Bytes, err := NewSerializer[string]().FromB64Bytes(b64Result).ToDynamicType(Reflect, value)
		test.CheckErr(err, "(b64Bytes) b64Bytes error", t)
		test.CheckEqualEquateEmpty(b64Bytes, "b64Bytes", value, t)
	})
}

func Fuzz_Dynamic_SliceOfStructs(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 22})

	f.Fuzz(func(t *testing.T, value []byte) {
		type Foo struct {
			Value []byte
		}
		var foos []*Foo
		for i := 0; i < rand.Int()%10; i++ {
			foos = append(foos, &Foo{Value: value})
		}

		// To the same type
		result1, err := NewSerializer[[]*Foo]().FromDynamicType(foos).ToT()
		test.CheckErr(err, "result 1 error", t)
		test.CheckEqual(result1, "result 1", foos, t)

		result2, err := NewSerializer[[]*Foo]().FromDynamicType(result1).ToDynamicType(Reflect, foos)
		test.CheckErr(err, "result 2 error", t)
		test.CheckEqual(result2, "result 2", foos, t)
		test.CheckEqual(result2, "result 2 (2)", result1, t)

		// To json
		jsonResult, err := NewSerializer[[]*Foo]().FromDynamicType(foos).ToJsonString()
		test.CheckErr(err, "(JSON) result error", t)

		fromJson, err := NewSerializer[[]*Foo]().FromDynamicType(jsonResult).ToDynamicType(Reflect, foos)
		test.CheckErr(err, "(JSON) jsonToSlice error", t)
		converted := fromJson.([]*Foo)
		for i, val := range converted {
			t.Logf("From1 %d, %+v", i, []byte(val.Value))
		}

		for i, val := range foos {
			t.Logf("From2 %d, %+v", i, []byte(val.Value))
		}
		test.CheckEqualEquateEmpty(fromJson, "JsonToSlice", foos, t)

		// To bytes
		byteResult, err := NewSerializer[[]*Foo]().FromDynamicType(foos).ToBytes()
		test.CheckErr(err, "(byteResult) result error", t)

		fromBytes, err := NewSerializer[[]*Foo]().FromDynamicType(byteResult).ToDynamicType(Reflect, foos)
		test.CheckErr(err, "(byteResult) byteResult error", t)
		test.CheckEqualEquateEmpty(fromBytes, "byteResult", foos, t)

		// To b64
		b64Result, err := NewSerializer[[]*Foo]().FromDynamicType(foos).ToB64Bytes()
		test.CheckErr(err, "(b64Bytes) result error", t)

		b64Bytes, err := NewSerializer[[]*Foo]().FromB64Bytes(b64Result).ToDynamicType(Reflect, foos)
		test.CheckErr(err, "(b64Bytes) b64Bytes error", t)
		test.CheckEqualEquateEmpty(b64Bytes, "b64Bytes", foos, t)
	})
}
