package binenc_test

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/binenc"
)

func TestReaderAlignBlockBoundary232(t *testing.T) {
	for pos := 0; pos < 32; pos++ {
		for bias := -8; bias <= 8; bias++ {
			t.Run(fmt.Sprintf("position_%d/bias_%d", pos, bias), func(t *testing.T) {
				r := binenc.NewReader(make([]byte, 48))
				if _, ok := r.ReadNext(pos); !ok {
					t.Fatal("initial cursor")
				}
				want := pos
				for (want+bias)%8 != 0 {
					want++
				}
				if !r.Align(bias) || r.Offset() != want {
					t.Fatalf("aligned cursor=%d want=%d", r.Offset(), want)
				}
				if !r.Align(bias) || r.Offset() != want {
					t.Fatalf("repeated alignment moved cursor=%d want=%d", r.Offset(), want)
				}
			})
		}
	}
}

func TestReaderAlignAtEOF232(t *testing.T) {
	for _, size := range []int{0, 8, 16} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			r := binenc.NewReader(make([]byte, size))
			if _, ok := r.ReadNext(size); !ok {
				t.Fatal("initial cursor")
			}
			if !r.Align(0) || r.Offset() != size || r.Remaining() != 0 {
				t.Fatalf("aligned EOF cursor=%d remaining=%d", r.Offset(), r.Remaining())
			}
		})
	}
}

func TestReaderAlignTruncatedPadding232(t *testing.T) {
	r := binenc.NewReader([]byte{1})
	_, _ = r.ReadNext(1)
	if r.Align(0) || r.Offset() != 1 || r.Remaining() != 0 {
		t.Fatalf("truncated padding advanced cursor=%d remaining=%d", r.Offset(), r.Remaining())
	}
}
