package httpkit

import (
	"fmt"
	"runtime"
	"testing"
)

func TestWrapError(t *testing.T) {
	t.Run("Caller", func(t *testing.T) {
		err := WrapError(fmt.Errorf("test wrap caller"))

		f, ok := err.Caller()
		if !ok {
			t.Fatal("wrap caller not exist")
		}

		if actual := 11; f.Line != actual {
			t.Fatalf("caller line, Expected = %d, Actual = %d", actual, f.Line)
		} else if actual := "github.com/joyparty/httpkit.TestWrapError.func1"; f.Function != actual {
			t.Fatalf("caller function, Expected = %s, Actual = %s", f.Function, actual)
		}
	})
}

func TestCaller(t *testing.T) {
	cases := []struct {
		Caller   *runtime.Frame
		Function string
	}{
		{
			Caller:   calleeFoo(),
			Function: "github.com/joyparty/httpkit.calleeFoo",
		},
		{
			Caller:   calleeBar(),
			Function: "github.com/joyparty/httpkit.calleeBar",
		},
		{
			Caller:   calleeFoobar(),
			Function: "github.com/joyparty/httpkit.calleeBar",
		},
		{
			Caller:   testCaller(),
			Function: "github.com/joyparty/httpkit.TestCaller",
		},
	}

	for _, c := range cases {
		if expected, actual := c.Function, c.Caller.Function; actual != expected {
			t.Fatalf("getCaller(), Expected=%s, Actual=%s", expected, actual)
		}
	}
}

func calleeFoobar() *runtime.Frame {
	return calleeBar()
}

func calleeFoo() *runtime.Frame {
	return testCaller()
}

func calleeBar() *runtime.Frame {
	return testCaller()
}

func testCaller() *runtime.Frame {
	return GetCaller()
}
