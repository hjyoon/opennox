package legacy

import (
	"reflect"
	"testing"
)

type objectNormalizeDirectionFixture4E7680 struct {
	direction uint16
}

func objectNormalizeDirectionTestHooks4E7680(
	events *[]int16,
	writes *int,
) objectNormalizeDirectionHooks4E7680[*objectNormalizeDirectionFixture4E7680] {
	return objectNormalizeDirectionHooks4E7680[*objectNormalizeDirectionFixture4E7680]{
		direction: func(obj *objectNormalizeDirectionFixture4E7680) int16 {
			return int16(obj.direction)
		},
		addDirection: func(obj *objectNormalizeDirectionFixture4E7680, delta int16) {
			*events = append(*events, delta)
			*writes++
			obj.direction += uint16(delta)
		},
	}
}

func TestObjectNormalizeDirection4E7680SignedRangeAndWriteCount(t *testing.T) {
	tests := []struct {
		name       string
		initial    uint16
		want       uint16
		wantWrites int
		wantDelta  int16
	}{
		{name: "zero", initial: 0, want: 0},
		{name: "one", initial: 1, want: 1},
		{name: "255", initial: 255, want: 255},
		{name: "256", initial: 256, want: 0, wantWrites: 1, wantDelta: -256},
		{name: "257", initial: 257, want: 1, wantWrites: 1, wantDelta: -256},
		{name: "signed maximum", initial: 0x7fff, want: 255, wantWrites: 127, wantDelta: -256},
		{name: "signed minimum", initial: 0x8000, want: 0, wantWrites: 128, wantDelta: 256},
		{name: "signed minimum plus one", initial: 0x8001, want: 1, wantWrites: 128, wantDelta: 256},
		{name: "minus 257", initial: 0xfeff, want: 255, wantWrites: 2, wantDelta: 256},
		{name: "minus 256", initial: 0xff00, want: 0, wantWrites: 1, wantDelta: 256},
		{name: "minus 255", initial: 0xff01, want: 1, wantWrites: 1, wantDelta: 256},
		{name: "minus one", initial: 0xffff, want: 255, wantWrites: 1, wantDelta: 256},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &objectNormalizeDirectionFixture4E7680{direction: tc.initial}
			var events []int16
			writes := 0
			got := objectNormalizeDirection4E7680(obj, objectNormalizeDirectionTestHooks4E7680(&events, &writes))
			if got != obj {
				t.Fatalf("return object = %p, want %p", got, obj)
			}
			if obj.direction != tc.want {
				t.Fatalf("direction = %#04x, want %#04x", obj.direction, tc.want)
			}
			if writes != tc.wantWrites {
				t.Fatalf("writes = %d, want %d", writes, tc.wantWrites)
			}
			for i, delta := range events {
				if delta != tc.wantDelta {
					t.Fatalf("write %d delta = %d, want %d", i, delta, tc.wantDelta)
				}
			}
		})
	}
}

func TestObjectNormalizeDirection4E7680PreservesLoopOrder(t *testing.T) {
	obj := &objectNormalizeDirectionFixture4E7680{direction: 0xfeff}
	var events []string
	objectNormalizeDirection4E7680(obj, objectNormalizeDirectionHooks4E7680[*objectNormalizeDirectionFixture4E7680]{
		direction: func(obj *objectNormalizeDirectionFixture4E7680) int16 {
			events = append(events, "load")
			return int16(obj.direction)
		},
		addDirection: func(obj *objectNormalizeDirectionFixture4E7680, delta int16) {
			events = append(events, "add")
			obj.direction += uint16(delta)
		},
	})
	want := []string{"load", "add", "load", "add", "load", "load"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectNormalizeDirection4E7680NilFaultsBeforeWrite(t *testing.T) {
	var obj *objectNormalizeDirectionFixture4E7680
	writes := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault")
		}
		if writes != 0 {
			t.Fatalf("writes before fault = %d, want 0", writes)
		}
	}()
	objectNormalizeDirection4E7680(obj, objectNormalizeDirectionHooks4E7680[*objectNormalizeDirectionFixture4E7680]{
		direction: func(obj *objectNormalizeDirectionFixture4E7680) int16 {
			return int16(obj.direction)
		},
		addDirection: func(*objectNormalizeDirectionFixture4E7680, int16) {
			writes++
		},
	})
}
