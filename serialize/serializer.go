package serialize

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type DynamicType string

// wrapper allows us to support both []T and T
type wrapper[T any] struct {
	T T
}

const (
	B64String  DynamicType = "b64string"
	JSONString DynamicType = "jsonstring"
	String     DynamicType = "string"
	Bytes      DynamicType = "bytes"
	B64Bytes   DynamicType = "b64bytes"
	Struct     DynamicType = "struct"
	Reflect    DynamicType = "reflection" // if provided, T is used

	// custom prefix added to encoded b64 strings to prevent possible misidentification of b64 strings when using Dynamic lookups
	B64Prefix string = "b64:"
)

type Serializer[T any] struct {
	isFromT          bool
	fromT            wrapper[T]
	isFromBytes      bool
	fromBytes        wrapper[[]byte]
	isFromJsonString bool
	isFromString     bool
	isFromB64String  bool
	isFromB64Bytes   bool
	fromString       wrapper[string]

	// isFromFailure is set when the from cannot be determine dynamically
	isFromFailure bool
}

func NewSerializer[T any]() *Serializer[T] {
	return &Serializer[T]{}
}

func (s *Serializer[T]) reset() {
	var t T
	s.isFromBytes = false
	s.isFromJsonString = false
	s.isFromB64String = false
	s.isFromT = false
	s.isFromB64Bytes = false
	s.isFromFailure = false
	s.isFromT = false
	s.fromT = wrapper[T]{T: t}
	s.fromBytes = wrapper[[]byte]{T: []byte{}}
	s.fromString = wrapper[string]{T: ""}
}

func (s *Serializer[T]) FromT(t T) *Serializer[T] {
	s.reset()
	s.isFromT = true
	s.fromT = wrapper[T]{T: t}
	return s
}

func (s *Serializer[T]) FromBytes(b []byte) *Serializer[T] {
	s.reset()
	s.isFromBytes = true
	s.fromBytes = wrapper[[]byte]{T: b}
	return s
}

func (s *Serializer[T]) FromJsonString(str string) *Serializer[T] {
	s.reset()
	s.isFromJsonString = true
	s.fromString = wrapper[string]{T: str}
	return s
}

func (s *Serializer[T]) FromString(str string) *Serializer[T] {
	s.reset()
	s.isFromString = true
	s.fromString = wrapper[string]{T: str}
	return s

}

func (s *Serializer[T]) FromB64String(str string) *Serializer[T] {
	s.reset()
	s.isFromB64String = true
	s.fromString = wrapper[string]{T: str}
	return s
}

func (s *Serializer[T]) FromB64Bytes(b []byte) *Serializer[T] {
	s.reset()
	s.isFromB64Bytes = true
	s.fromBytes = wrapper[[]byte]{T: b}
	return s
}

func (s *Serializer[T]) fromFailure() *Serializer[T] {
	s.reset()
	s.isFromFailure = true
	return s
}

// FromDynamicType will attempt to infer T and call the appropriate From method. If the type is not inferable, it will return the serializer as is, which
// will result in an error when calling a To function.
func (s *Serializer[T]) FromDynamicType(t any) *Serializer[T] {
	typ := reflect.TypeOf(t)
	switch typ.Kind() {
	case reflect.Slice:
		_, isByteSlice := t.([]byte)

		if !isByteSlice {
			converted, isT := t.(T)
			if isT {
				return s.FromT(converted)
			}

			return s.fromFailure()
		}

		fallthrough
	case reflect.Struct, reflect.String:
		return s.setFrom(t)
	default:
		return s.fromFailure()
	}
}

func (s *Serializer[T]) setFrom(t any) *Serializer[T] {
	typ := reflect.TypeOf(t)
	switch typ.Kind() {
	case reflect.String:
		str, _ := t.(string)
		if isJSON(str) {
			return s.FromJsonString(str)
		}

		if isB64IntermediateEncoding(str) {
			return s.FromB64String(str)
		}

		return s.FromString(str)
	case reflect.Slice:
		b, ok := t.([]byte)
		if !ok {
			return s.fromFailure()
		}

		if isB64BytesIntermediateEncoding(b) {
			return s.FromB64Bytes(b)
		}

		return s.FromBytes(b)
	case reflect.Struct:
		converted := t.(T)
		return s.FromT(converted)
	default:
		return s.fromFailure()
	}
}

// ToT converts From to T
func (s *Serializer[T]) ToT() (T, error) {
	var t T

	if s.isFromBytes {
		err := decodeBytesToAny(s.fromBytes.T, &t)
		return t, err
	}

	if s.isFromJsonString || s.isFromString {
		strBytes := []byte(s.fromString.T)
		err := json.Unmarshal(strBytes, &t)
		if err != nil {
			return t, fmt.Errorf("Serializer: Failed to unmarhsal string to struct. %w. Is the string valid JSON?", err)
		}
		return t, nil
	}

	if s.isFromB64String {
		err := decodeFromBase64StringToAny(s.fromString.T, &t)
		if err != nil {
			return t, fmt.Errorf("Serializer: Failed to decode b64 string to struct. %w. Is the b64 encoded string valid JSON? Use this serialize to create it.", err)
		}
		return t, nil
	}

	if s.isFromB64Bytes {
		err := decodeFromBase64BytesToAny(s.fromBytes.T, &t)
		if err != nil {
			return t, fmt.Errorf("Serializer: Failed to decode b64 bytes to struct. %w. Are the bytes b64 encoded and valid JSON? Use this serialize to create it.", err)
		}
		return t, nil
	}

	if s.isFromT {
		return s.fromT.T, nil
	}

	return t, errors.New("No From() method was defined. Please call From() before calling To(). e.g. foo, err := NewSerializer[Foo].FromBytes(b).ToT()")
}

// ToBytes converts the from input as a byte slice
func (s *Serializer[T]) ToBytes() ([]byte, error) {
	if s.isFromT {
		b, err := encodeAnyToBytes(s.fromT.T)
		if err != nil {
			return b, fmt.Errorf("Serializer: Failed to encode struct to bytes. %w", err)
		}
		return b, nil
	}

	if s.isFromB64Bytes {
		return decodeFromBase64ToBytes(s.fromBytes.T)
	}
	if s.isFromB64String {
		return decodeFromBase64StringToBytes(s.fromString.T)
	}

	if s.isFromString {
		return []byte(s.fromString.T), nil
	}

	if s.isFromJsonString {
		return []byte(s.fromString.T), nil
	}

	if s.isFromBytes {
		return s.fromBytes.T, nil
	}

	return []byte{}, errors.New("No From() method was defined. Please call From() before calling To(). e.g. b, err := NewSerializer[Foo].FromT(f).ToBytes()")
}

// ToB64String converts the from input as a b64 encoded string
func (s *Serializer[T]) ToB64String() (string, error) {
	if s.isFromBytes {
		return encodeBytesToB64String(s.fromBytes.T), nil
	}

	if s.isFromT {
		return encodeToB64String(s.fromT.T)
	}

	if s.isFromB64String {
		return s.fromString.T, nil
	}

	if s.isFromB64Bytes {
		return string(s.fromBytes.T), nil
	}

	if s.isFromJsonString {
		return encodeStringToB64String(s.fromString.T), nil
	}

	if s.isFromString {
		return encodeStringToB64String(s.fromString.T), nil
	}

	return "", errors.New("No From() method was defined. Please call From() before calling To(). e.g. str, err := NewSerializer[Foo].FromT(f).ToB64String()")
}

// ToB64Bytes encodes the from input as a b64 encoded byte slice
func (s *Serializer[T]) ToB64Bytes() ([]byte, error) {
	if s.isFromT {
		return encodeToB64Bytes(s.fromT.T)
	}

	if s.isFromB64String {
		return []byte(s.fromString.T), nil
	}

	if s.isFromB64Bytes {
		return s.fromBytes.T, nil
	}

	if s.isFromJsonString || s.isFromString {
		return encodeStringToB64(s.fromString.T), nil
	}

	if s.isFromBytes {
		return encodeBytesToB64(s.fromBytes.T), nil
	}

	return []byte{}, errors.New("No From() method was defined. Please call From() before calling To(). e.g. str, err := NewSerializer[Foo].FromT(f).ToB64Bytes()")
}

// ToJsonString encodes the from input as a json string
func (s *Serializer[T]) ToJsonString() (string, error) {
	if s.isFromBytes {
		return string(s.fromBytes.T), nil
	}

	if s.isFromT {
		return encodeAnyToJson(s.fromT.T)
	}

	if s.isFromB64String {
		jsonStr, err := decodeFromBase64StringToJsonString(s.fromString.T)
		return jsonStr, err
	}

	if s.isFromJsonString || s.isFromString {
		return s.fromString.T, nil
	}

	if s.isFromB64Bytes {
		jsonStr, err := decodeFromBase64ToJsonString(s.fromBytes.T)
		return jsonStr, err

	}

	return "", errors.New("No From() method was defined. Please call From() before calling To(). e.g. str, err := NewSerializer[Foo].FromT(f).ToJsonString()")
}

// ToString encodes the from input as a string
func (s *Serializer[T]) ToString() (string, error) {
	if s.isFromBytes {
		return string(s.fromBytes.T), nil
	}

	if s.isFromT {
		return encodeAnyToJson(s.fromT.T)
	}

	if s.isFromB64String {
		b, err := decodeFromBase64StringToBytes(s.fromString.T)
		return string(b), err
	}

	if s.isFromJsonString || s.isFromString {
		return s.fromString.T, nil
	}

	if s.isFromB64Bytes {
		b, err := decodeFromBase64ToBytes(s.fromBytes.T)
		return string(b), err
	}

	return "", errors.New("No From() method was defined. Please call From() before calling To(). e.g. str, err := NewSerializer[Foo].FromT(f).ToJsonString()")
}

// ToDynamicType allows a single method to dynamically determine the toType depending on dynamic type. If "reflect" is provided, then parameter
// typ will be used to determine the target type to convert to. This will be a best-effort conversion.
func (s *Serializer[T]) ToDynamicType(dynamicType DynamicType, typ any) (any, error) {
	switch dynamicType {
	case B64String:
		return s.ToB64String()
	case JSONString:
		return s.ToJsonString()
	case String:
		return s.ToString()
	case Bytes:
		return s.ToBytes()
	case B64Bytes:
		return s.ToB64Bytes()
	case Struct:
		return s.ToT()
	case Reflect:
		if typ == nil {
			return nil, fmt.Errorf("Serializer: Failed to infer type. nil provided for parameter: typ")
		}
		switch reflect.TypeOf(typ).Kind() {
		case reflect.String:
			return s.ToString()
		case reflect.Slice:
			_, isByteSlice := typ.([]byte)
			if isByteSlice {
				// always assume to is bytes, there is no way to anticipate from types that we want b64 encoded bytes or normal bytes since a random
				// []byte might pass the encoding test even though the provided `typ` was not intended to be b64 encoded.
				return s.ToBytes()
			}

			_, isT := typ.(T)
			if isT {
				return s.ToT()
			}

			return fmt.Errorf("Serializer: Unsupported type: %s. Slice is not []byte or of type T", reflect.ValueOf(typ).Kind().String()), nil
		case reflect.Struct:
			return s.ToT()
		default:
			return fmt.Errorf("Serializer: Unsupported type: %s", reflect.ValueOf(typ).Kind().String()), nil
		}
	default:
		return nil, fmt.Errorf("Serializer: Failed to infer type. Unknown dynamic type provided: %s", dynamicType)
	}

}

func decodeFromBase64StringToAny(from string, to any) error {
	fromB, err := decodeFromBase64StringToBytes(from)
	if err != nil {
		return fmt.Errorf("Serializer: Failed to decode b64 string to struct. %w", err)
	}
	return json.Unmarshal(fromB, to)
}

func decodeFromBase64BytesToAny(from []byte, to any) error {
	b, err := decodeFromBase64ToBytes(from)
	if err != nil {
		return fmt.Errorf("Serializer: Failed to decode from b64 to bytes. %w", err)
	}
	return json.NewDecoder(bytes.NewBuffer(b)).Decode(to)
}

func decodeFromBase64StringToJsonString(from string) (string, error) {
	bytes, err := decodeFromBase64StringToBytes(from)
	if err != nil {
		return "", fmt.Errorf("Serializer: Failed to decode b64 string to json string. %w", err)
	}
	jsonStr := string(bytes)
	if !isJSON(jsonStr) {
		return "", fmt.Errorf("Serializer: Failed to decode b64 string to json string. The b64 encoded string is not valid JSON. %s", jsonStr)
	}

	return jsonStr, nil
}
func decodeFromBase64ToBytes(from []byte) ([]byte, error) {
	return decodeFromBase64StringToBytes(string(from))
}

func decodeFromBase64ToJsonString(from []byte) (string, error) {
	jsonStr, err := decodeB64ToString(from)
	if err != nil {
		return "", fmt.Errorf("Serializer: Failed to decode b64 to json string. %w", err)
	}

	if !isJSON(jsonStr) {
		return "", fmt.Errorf("Serializer: Failed to decode b64 to json string. The encoded bytes are not valid JSON. %s", jsonStr)
	}

	return jsonStr, nil
}

func fromAnyToJson(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("Serializer: Failed to marshal to json. %w", err)
	}
	return string(b), err
}

func encodeToB64String(v any) (string, error) {
	// first encode to json
	jsonStr, err := fromAnyToJson(v)
	if err != nil {
		return "", fmt.Errorf("Serializer: Failed to encode to b64 string. %w", err)
	}
	return encodeStringToB64String(jsonStr), nil
}

func encodeToB64Bytes(v any) ([]byte, error) {
	// first encode to json
	jsonStr, err := fromAnyToJson(v)
	if err != nil {
		return []byte(jsonStr), fmt.Errorf("Serializer: Failed to encode to b64 bytes. %w", err)
	}

	return encodeStringToB64(jsonStr), nil
}

func encodeStringToB64String(v string) string {
	return encodeBytesToB64String([]byte(v))
}

func encodeBytesToB64String(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeFromBase64StringToBytes(from string) ([]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(from)
	if err != nil {
		return bytes, fmt.Errorf("Serializer: Failed to decode b64 string to bytes. %w", err)
	}
	return bytes, nil
}

func encodeStringToB64(v string) []byte {
	return []byte(encodeStringToB64String(v))
}

func encodeAnyToBytes(v any) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return buf.Bytes(), fmt.Errorf("Serializer: Failed to encode to bytes. %w", err)
	}
	return buf.Bytes(), nil
}

func decodeBytesToAny(b []byte, to any) error {
	dec := gob.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(to)
	if err != nil {
		return fmt.Errorf("Serializer: Failed to decode bytes to struct. Are the source bytes a previously encoded struct? %w", err)
	}

	return nil
}

func decodeStringToType[T any](s string) (T, error) {
	var t T
	err := json.Unmarshal([]byte(s), &t)
	return t, err
}

func decodeB64ToString(b []byte) (string, error) {
	decodedBytes, err := decodeFromBase64StringToBytes(string(b))
	if err != nil {
		return "", fmt.Errorf("Serializer: Failed to decode b64 to string. %w", err)
	}
	return string(decodedBytes), err
}

func encodeBytesToB64(v []byte) []byte {
	return encodeStringToB64(string(v))
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func isB64IntermediateEncoding(s string) bool {
	return strings.HasPrefix(s, B64Prefix)
}

func isB64BytesIntermediateEncoding(b []byte) bool {
	str := string(b)
	return isB64IntermediateEncoding(str) && len(b) > 0
}

func encodeAnyToJson(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return string(b), fmt.Errorf("Serializer: Failed to encode struct to json. %w", err)
	}
	return string(b), nil
}
