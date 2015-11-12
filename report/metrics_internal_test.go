package report

import (
	"testing"
	"time"
)

func TestFirstLast(t *testing.T) {
	zero, t1, t2 := time.Time{}, time.Now(), time.Now().Add(1*time.Minute)
	tests := []struct {
		arg1, arg2, first, last time.Time
	}{
		{zero, zero, zero, zero},
		{t1, zero, t1, t1},
		{zero, t1, t1, t1},
		{t1, t1, t1, t1},
		{t1, t2, t1, t2},
		{t2, t1, t1, t2},
	}
	for _, test := range tests {
		if got := first(test.arg1, test.arg2); !got.Equal(test.first) {
			t.Errorf("first(%q, %q) => %q, Expected: %q", test.arg1, test.arg2, got, test.first)
		}

		if got := last(test.arg1, test.arg2); !got.Equal(test.last) {
			t.Errorf("last(%q, %q) => %q, Expected: %q", test.arg1, test.arg2, got, test.last)
		}
	}
}
