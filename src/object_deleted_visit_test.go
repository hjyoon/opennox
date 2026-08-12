package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestVisitDeletedObjects4E5F00Empty(t *testing.T) {
	visitDeletedObjects_4E5F00(nil, func() uint32 {
		t.Fatal("empty visit read frame")
		return 0
	}, nil, -1)
}

func TestVisitDeletedObjects4E5F00CurrentFrameDoesNotCallNilVisitor(t *testing.T) {
	obj := &server.Object{DeletedAt: 42}
	frameCalls := 0
	visitDeletedObjects_4E5F00(obj, func() uint32 {
		frameCalls++
		return 42
	}, nil, -1)
	if frameCalls != 1 {
		t.Fatalf("frame calls = %d, want 1", frameCalls)
	}
}

func TestVisitDeletedObjects4E5F00OrderFilterAndArgument(t *testing.T) {
	const (
		frame = 100
		arg   = int32(-0x1234567)
	)
	first := &server.Object{DeletedAt: frame}
	second := &server.Object{DeletedAt: frame - 1}
	third := &server.Object{DeletedAt: frame}
	fourth := &server.Object{DeletedAt: frame + 1}
	first.DeletedNext = second
	second.DeletedNext = third
	third.DeletedNext = fourth

	frameCalls := 0
	want := []*server.Object{second, fourth}
	var got []*server.Object
	visitDeletedObjects_4E5F00(first, func() uint32 {
		frameCalls++
		return frame
	}, func(obj *server.Object, gotArg int32) int32 {
		if gotArg != arg {
			t.Fatalf("callback argument = %#x, want %#x", gotArg, arg)
		}
		got = append(got, obj)
		return int32(len(got))
	}, arg)
	if frameCalls != 4 {
		t.Fatalf("frame calls = %d, want 4", frameCalls)
	}
	if len(got) != len(want) {
		t.Fatalf("callback count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("callback %d = %p, want %p", i, got[i], want[i])
		}
	}
}

func TestVisitDeletedObjects4E5F00ReadOrder(t *testing.T) {
	const frame = 7
	first := &server.Object{DeletedAt: frame}
	originalNext := &server.Object{DeletedAt: frame - 1}
	replacementNext := &server.Object{DeletedAt: frame - 1}
	final := &server.Object{DeletedAt: frame - 1}
	first.DeletedNext = originalNext
	replacementNext.DeletedNext = originalNext

	frameCalls := 0
	var got []*server.Object
	visitDeletedObjects_4E5F00(first, func() uint32 {
		frameCalls++
		if frameCalls == 1 {
			first.DeletedAt = frame + 1
			first.DeletedNext = replacementNext
		}
		return frame
	}, func(obj *server.Object, _ int32) int32 {
		got = append(got, obj)
		if obj == replacementNext {
			replacementNext.DeletedNext = final
		}
		return -1
	}, 0)

	want := []*server.Object{replacementNext, final}
	if frameCalls != 3 {
		t.Fatalf("frame calls = %d, want 3", frameCalls)
	}
	if len(got) != len(want) {
		t.Fatalf("callback count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("callback %d = %p, want %p", i, got[i], want[i])
		}
	}
}
