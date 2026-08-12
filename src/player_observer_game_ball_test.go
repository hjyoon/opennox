package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerObserverFindGameBall4E6230SeedsBeforeEmptyList(t *testing.T) {
	var calls []string
	got := playerObserverFindGameBall_4E6230(
		func() { calls = append(calls, "seed") },
		func() uint32 {
			t.Fatal("read GameBall ID for an empty list")
			return 0
		},
		func() *server.Object {
			calls = append(calls, "first")
			return nil
		},
		func(*server.Object) *server.Object {
			t.Fatal("read next object for an empty list")
			return nil
		},
	)
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if want := []string{"seed", "first"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFindGameBall4E6230ReloadsCachedID(t *testing.T) {
	first := &server.Object{TypeInd: 2}
	second := &server.Object{TypeInd: 5}
	var calls []string
	reads := 0
	got := playerObserverFindGameBall_4E6230(
		func() { calls = append(calls, "seed") },
		func() uint32 {
			reads++
			calls = append(calls, "id")
			if reads == 1 {
				return 3
			}
			return 5
		},
		func() *server.Object {
			calls = append(calls, "first")
			return first
		},
		func(obj *server.Object) *server.Object {
			if obj != first {
				t.Fatalf("next argument = %p, want first object", obj)
			}
			calls = append(calls, "next")
			return second
		},
	)
	if got != second {
		t.Fatalf("result = %p, want second object %p", got, second)
	}
	if want := []string{"seed", "first", "id", "next", "id"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerObserverFindGameBall4E6230ReadsIDBeforeType(t *testing.T) {
	obj := &server.Object{TypeInd: 1}
	nextCalls := 0
	got := playerObserverFindGameBall_4E6230(
		func() {},
		func() uint32 {
			obj.TypeInd = 7
			return 7
		},
		func() *server.Object { return obj },
		func(*server.Object) *server.Object {
			nextCalls++
			return nil
		},
	)
	if got != obj {
		t.Fatalf("result = %p, want object %p", got, obj)
	}
	if nextCalls != 0 {
		t.Fatalf("next calls = %d, want 0 after a match", nextCalls)
	}
}

func TestPlayerObserverFindGameBall4E6230ZeroExtendsTypeIndex(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   uint32
		want bool
	}{
		{name: "sixteen bit match", id: 0x0000ffff, want: true},
		{name: "wide value does not match", id: 0x0001ffff, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &server.Object{TypeInd: 0xffff}
			got := playerObserverFindGameBall_4E6230(
				func() {},
				func() uint32 { return tc.id },
				func() *server.Object { return obj },
				func(*server.Object) *server.Object { return nil },
			)
			if (got != nil) != tc.want {
				t.Fatalf("result = %p for cached ID %#x, want match %t", got, tc.id, tc.want)
			}
		})
	}
}

func TestPlayerObserverFindGameBall4E6230ScansCachedZero(t *testing.T) {
	obj := &server.Object{TypeInd: 0}
	got := playerObserverFindGameBall_4E6230(
		func() {},
		func() uint32 { return 0 },
		func() *server.Object { return obj },
		func(*server.Object) *server.Object {
			t.Fatal("advanced past matching zero type")
			return nil
		},
	)
	if got != obj {
		t.Fatalf("result = %p, want zero-type object %p", got, obj)
	}
}
