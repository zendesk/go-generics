package datastructures

import "reflect"

type Tuple[A any, B any] struct {
	A A
	B B
}

type TupleWithErr[A any, B any] struct {
	A A
	B B
	E error
}

func (t TupleWithErr[A, B]) HasError() bool {
	return t.E != nil
}

// HasValue returns true if the tuple has a non-zero value set for either A or B
func (t TupleWithErr[A, B]) HasValue() bool {
	return !reflect.ValueOf(t.A).IsZero() || !reflect.ValueOf(t.B).IsZero()
}
