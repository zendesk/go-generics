package test

import (
	"errors"
	"testing"
)

type mockT struct {
	FatalfCalled bool
	FatalCalled  bool
	ErrorfCalled bool
}

func (t *mockT) Errorf(format string, args ...any) {
	t.ErrorfCalled = true
}

func (t *mockT) Fatalf(format string, args ...any) {
	t.FatalfCalled = true
}

func (t *mockT) Fatal(args ...any) {
	t.FatalCalled = true
}

func TestCheckErr(t *testing.T) {
	mock := &mockT{}
	err := errors.New("test error")
	CheckErr(err, "CheckErr failed", mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}

	mock = &mockT{}
	CheckErr(nil, "CheckErr failed", mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}
}

func TestCheckErrs(t *testing.T) {
	mock := &mockT{}
	errs := []error{errors.New("error1"), errors.New("error2")}
	CheckErrs(errs, "CheckErrs failed", mock)

	if !mock.FatalfCalled {
		t.Errorf("Expected Fatalf to be called, but it was not")
	}

	mock = &mockT{}
	CheckErrs(nil, "CheckErrs failed", mock)

	if mock.FatalfCalled {
		t.Errorf("Expected Fatalf not to be called, but it was")
	}
}

func TestExpectNil(t *testing.T) {
	mock := &mockT{}
	value := "not nil"
	ExpectNil(value, mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}

	mock = &mockT{}
	var nilValue interface{} = nil
	ExpectNil(nilValue, mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}
}

func TestExpectNotEmpty(t *testing.T) {
	mock := &mockT{}
	value := ""
	ExpectNotEmpty(value, "value", mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}

	mock = &mockT{}
	nonEmptyValue := "not empty"
	ExpectNotEmpty(nonEmptyValue, "value", mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}
}

func TestExpectNotNil(t *testing.T) {
	mock := &mockT{}
	var value interface{} = nil
	ExpectNotNil("value", value, mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}

	mock = &mockT{}
	nonNilValue := "not nil"
	ExpectNotNil("value", nonNilValue, mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}
}

func TestExpectErr(t *testing.T) {
	mock := &mockT{}
	err := errors.New("test error")
	ExpectErr(err, "ExpectErr failed", mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}

	mock = &mockT{}
	ExpectErr(nil, "ExpectErr failed", mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}
}

func TestCheckEqual(t *testing.T) {
	mock := &mockT{}
	result := 42
	expected := 42
	CheckEqual(result, "result", expected, mock)

	if mock.FatalfCalled {
		t.Errorf("Expected Fatalf not to be called, but it was")
	}

	mock = &mockT{}
	expected = 43
	CheckEqual(result, "result", expected, mock)

	if !mock.FatalfCalled {
		t.Errorf("Expected Fatalf to be called, but it was not")
	}
}

func TestCheckContains(t *testing.T) {
	mock := &mockT{}
	list := []int{1, 2, 3}
	item := 2
	CheckContains(list, item, mock)

	if mock.FatalfCalled {
		t.Errorf("Expected Fatalf not to be called, but it was")
	}

	mock = &mockT{}
	item = 4
	CheckContains(list, item, mock)

	if !mock.FatalfCalled {
		t.Errorf("Expected Fatalf to be called, but it was not")
	}
}

func TestCheckOk(t *testing.T) {
	mock := &mockT{}
	CheckOk(true, "CheckOk failed", mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}

	mock = &mockT{}
	CheckOk(false, "CheckOk failed", mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}
}

func TestCheckNotOk(t *testing.T) {
	mock := &mockT{}
	CheckNotOk(false, "CheckNotOk failed", mock)

	if mock.FatalCalled {
		t.Errorf("Expected Fatal not to be called, but it was")
	}

	mock = &mockT{}
	CheckNotOk(true, "CheckNotOk failed", mock)

	if !mock.FatalCalled {
		t.Errorf("Expected Fatal to be called, but it was not")
	}
}
