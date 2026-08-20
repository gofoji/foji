package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gofoji/foji/runtime"
)

type nilTestStruct struct{ A int }

type nilTestError struct{}

func (*nilTestError) Error() string { return "" }

// TestIsNil pins the exact semantics the template helpers `isNil` / `isNotNil`
// have always had: only pointer-shaped kinds can report nil. A nil slice is
// deliberately false — it was false under the previous unsafe implementation
// that inspected the interface data word, and templates depend on that.
func TestIsNil(t *testing.T) {
	var (
		nilPointer   *int
		nilMap       map[string]int
		nilChannel   chan int
		nilFunc      func()
		nilSlice     []int
		nilInterface any
		nilError     error
	)

	value := 5

	tests := []struct {
		name string
		in   any
		want bool
	}{
		{"nil interface", nilInterface, true},
		{"nil pointer", nilPointer, true},
		{"nil map", nilMap, true},
		{"nil channel", nilChannel, true},
		{"nil func", nilFunc, true},
		{"nil error interface", nilError, true},
		{"typed nil in error interface", error((*nilTestError)(nil)), true},

		{"nil slice", nilSlice, false},
		{"empty slice", []int{}, false},
		{"populated slice", []int{1}, false},
		{"non-nil pointer", &value, false},
		{"non-nil map", map[string]int{}, false},
		{"zero int", 0, false},
		{"int", 5, false},
		{"empty string", "", false},
		{"string", "a", false},
		{"false", false, false},
		{"zero struct", nilTestStruct{}, false},
		{"pointer to struct", &nilTestStruct{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, runtime.IsNil(tt.in))
			assert.Equal(t, !tt.want, runtime.IsNotNil(tt.in))
		})
	}
}
