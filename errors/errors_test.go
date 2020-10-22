// +build !debug

package errors

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

func TestDebug(t *testing.T) {
	// Test with -tags debug to run the tests in debug_test.go
	cmd := exec.Command("go", "test", "-tags", "prod")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("external go test failed: %v", err)
	}
}

func TestMarshal(t *testing.T) {
	// Single error. No user is set, so we will have a zero-length field inside.
	e1 := E(Op("Get"), Network, "caching in progress")

	// Nested error.
	e2 := E(Op("Read"), Undefined, e1)

	b := MarshalError(e2)
	e3 := UnmarshalError(b)

	in := e2.(*Error)
	out := e3.(*Error)

	// Compare elementwise.
	if in.Op != out.Op {
		t.Errorf("expected Op %q; got %q", in.Op, out.Op)
	}
	if in.Kind != out.Kind {
		t.Errorf("expected kind %d; got %d", in.Kind, out.Kind)
	}
	// Note that error will have lost type information, so just check its Error string.
	if in.Err.Error() != out.Err.Error() {
		t.Errorf("expected Err %q; got %q", in.Err, out.Err)
	}
}

func TestSeparator(t *testing.T) {
	defer func(prev string) {
		Separator = prev
	}(Separator)
	Separator = ":: "

	// Single error. No user is set, so we will have a zero-length field inside.
	e1 := E(Op("Get"), Network, "network error")

	// Nested error.
	e2 := E(Op("Get"), Network, e1)

	want := "Get: Network error:: Get: network error"
	if errorAsString(e2) != want {
		t.Errorf("expected %q; got %q", want, e2)
	}
}

func TestDoesNotChangePreviousError(t *testing.T) {
	err := E(Network)
	err2 := E(Op("I will NOT modify err"), err)

	expected := "I will NOT modify err: Network error"
	if errorAsString(err2) != expected {
		t.Fatalf("Expected %q, got %q", expected, err2)
	}
	kind := err.(*Error).Kind
	if kind != Network {
		t.Fatalf("Expected kind %v, got %v", Network, kind)
	}
}

//func TestNoArgs(t *testing.T) {
//	defer func() {
//		err := recover()
//		if err == nil {
//			t.Fatal("E() did not panic")
//		}
//	}()
//	_ = E()
//}

type matchTest struct {
	err1, err2 error
	matched    bool
}

const (
	op  = Op("Op")
	op1 = Op("Op1")
	op2 = Op("Op2")
)

var matchTests = []matchTest{
	// Errors not of type *Error fail outright.
	{nil, nil, false},
	{io.EOF, io.EOF, false},
	{E(io.EOF), io.EOF, false},
	{io.EOF, E(io.EOF), false},
	// Success. We can drop fields from the first argument and still match.
	{E(io.EOF), E(io.EOF), true},
	{E(op, Other, io.EOF), E(op, Other, io.EOF), true},
	{E(op, Other, io.EOF, "test"), E(op, Other, io.EOF, "test", "test"), true},
	{E(op, Other), E(op, Other, io.EOF, "test", "test"), true},
	{E(op), E(op, Other, io.EOF, "test", "test"), true},
	// Failure.
	{E(io.EOF), E(io.ErrClosedPipe), false},
	{E(op1), E(op2), false},
	{E(Other), E(Network), false},
	{E("test"), E("test1"), false},
	{E(fmt.Errorf("error")), E(fmt.Errorf("error1")), false},
	{E(op, Other, io.EOF, "test", "test1"), E(op, Other, io.EOF, "test", "test"), false},
	{E("test", Str("something")), E("test"), false}, // Test nil error on rhs.
	// Nested *Errors.
	{E(op1, E("test")), E(op1, "1", E(op2, "2", "test")), true},
	{E(op1, "test"), E(op1, "1", E(op2, "2", "test")), false},
	{E(op1, E("test")), E(op1, "1", Str(E(op2, "2", "test").Error())), false},
}

func TestMatch(t *testing.T) {
	for _, test := range matchTests {
		matched := Match(test.err1, test.err2)
		if matched != test.matched {
			t.Errorf("Match(%q, %q)=%t; want %t", test.err1, test.err2, matched, test.matched)
		}
	}
}

type kindTest struct {
	err  error
	kind Kind
	want bool
}

var kindTests = []kindTest{
	//Non-Error errors.
	{nil, Network, false},
	{Str("not an *Error"), Network, false},

	// Basic comparisons.
	{E(Network), Network, true},
	{E(Test), Network, false},
	{E("no kind"), Network, false},
	{E("no kind"), Other, false},

	// Nested *Error values.
	{E("Nesting", E(Network)), Network, true},
	{E("Nesting", E(Test)), Network, false},
	{E("Nesting", E("no kind")), Network, false},
	{E("Nesting", E("no kind")), Other, false},
}

func TestKind(t *testing.T) {
	for _, test := range kindTests {
		got := Is(test.kind, test.err)
		if got != test.want {
			t.Errorf("Is(%q, %q)=%t; want %t", test.kind, test.err, got, test.want)
		}
	}
}

func errorAsString(err error) string {
	if e, ok := err.(*Error); ok {
		e2 := *e
		e2.stack = stack{}
		return e2.Error()
	}
	return err.Error()
}
