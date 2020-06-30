package httpkit

import (
	"fmt"
	"testing"
)

func TestWrapError(t *testing.T) {
	t.Run("Caller", func(t *testing.T) {
		err := WrapError(fmt.Errorf("test wrap caller"))

		f, ok := err.Caller()
		if !ok {
			t.Fatal("wrap caller not exist")
		}

		if actual := 10; f.Line != actual {
			t.Fatalf("caller line, Expected = %d, Actual = %d", actual, f.Line)
		} else if actual := "github.com/joyparty/httpkit.TestWrapError.func1"; f.Function != actual {
			t.Fatalf("caller function, Expected = %s, Actual = %s", f.Function, actual)
		}
	})
}
