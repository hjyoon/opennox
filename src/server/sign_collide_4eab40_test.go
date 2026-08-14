package server

import (
	"reflect"
	"testing"
)

type signCollideTestObject4EAB40 struct {
	class uint8
	use   *signCollideTestUse4EAB40
	guard uint32
}

type signCollideTestUse4EAB40 struct {
	id uint32
}

func defaultSignCollideHooks4EAB40() signCollideHooks4EAB40[*signCollideTestObject4EAB40, *signCollideTestUse4EAB40] {
	return signCollideHooks4EAB40[*signCollideTestObject4EAB40, *signCollideTestUse4EAB40]{
		classLow: func(obj *signCollideTestObject4EAB40) uint8 {
			return obj.class
		},
		loadUse: func(obj *signCollideTestObject4EAB40) *signCollideTestUse4EAB40 {
			return obj.use
		},
		callUse: func(*signCollideTestUse4EAB40, *signCollideTestObject4EAB40, *signCollideTestObject4EAB40) int32 {
			return 0
		},
	}
}

func TestSignCollide4EAB40TargetGuardsBeforeSource(t *testing.T) {
	tests := []struct {
		name       string
		target     *signCollideTestObject4EAB40
		wantEvents []string
	}{
		{name: "nil target"},
		{name: "non player", target: &signCollideTestObject4EAB40{class: 0x80}, wantEvents: []string{"class"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := defaultSignCollideHooks4EAB40()
			hooks.classLow = func(obj *signCollideTestObject4EAB40) uint8 {
				events = append(events, "class")
				return obj.class
			}
			hooks.loadUse = func(*signCollideTestObject4EAB40) *signCollideTestUse4EAB40 {
				t.Fatal("source path reached")
				return nil
			}
			signCollide4EAB40[*signCollideTestObject4EAB40](nil, tc.target, func() { t.Fatal("collision observed") }, hooks)
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestSignCollide4EAB40CachesUseAndCallsTargetSource(t *testing.T) {
	first := &signCollideTestUse4EAB40{id: 1}
	second := &signCollideTestUse4EAB40{id: 2}
	source := &signCollideTestObject4EAB40{use: first, guard: 0x11223344}
	target := &signCollideTestObject4EAB40{class: 0x84, guard: 0x55667788}
	collision := &struct{ guard uint32 }{guard: 0x31415926}
	events := make([]string, 0, 3)
	hooks := defaultSignCollideHooks4EAB40()
	hooks.classLow = func(obj *signCollideTestObject4EAB40) uint8 {
		events = append(events, "class")
		if obj != target {
			t.Fatalf("class object = %p", obj)
		}
		return obj.class
	}
	hooks.loadUse = func(obj *signCollideTestObject4EAB40) *signCollideTestUse4EAB40 {
		events = append(events, "use")
		if obj != source {
			t.Fatalf("use object = %p", obj)
		}
		source.use = second
		return first
	}
	hooks.callUse = func(use *signCollideTestUse4EAB40, gotTarget, gotSource *signCollideTestObject4EAB40) int32 {
		events = append(events, "call")
		if use != first || gotTarget != target || gotSource != source {
			t.Fatalf("call = %p/%p/%p", use, gotTarget, gotSource)
		}
		target.guard++
		return -0x1234567
	}

	signCollide4EAB40(source, target, collision, hooks)
	if want := []string{"class", "use", "call"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.use != second || source.guard != 0x11223344 || target.guard != 0x55667789 || collision.guard != 0x31415926 {
		t.Fatalf("state = source %+v target %+v collision %+v", source, target, collision)
	}
}

func TestSignCollide4EAB40NilSourceFaultsAfterTargetClass(t *testing.T) {
	target := &signCollideTestObject4EAB40{class: signCollidePlayerClass4EAB40}
	events := make([]string, 0, 2)
	hooks := defaultSignCollideHooks4EAB40()
	hooks.classLow = func(obj *signCollideTestObject4EAB40) uint8 {
		events = append(events, "class")
		return obj.class
	}
	hooks.loadUse = func(obj *signCollideTestObject4EAB40) *signCollideTestUse4EAB40 {
		events = append(events, "use")
		return obj.use
	}
	hooks.callUse = func(*signCollideTestUse4EAB40, *signCollideTestObject4EAB40, *signCollideTestObject4EAB40) int32 {
		t.Fatal("call reached")
		return 0
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault")
		}
		if want := []string{"class", "use"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	signCollide4EAB40[*signCollideTestObject4EAB40](nil, target, 0, hooks)
}

func TestSignCollide4EAB40NilUseFaultsAtCall(t *testing.T) {
	source := &signCollideTestObject4EAB40{}
	target := &signCollideTestObject4EAB40{class: signCollidePlayerClass4EAB40}
	events := make([]string, 0, 3)
	hooks := defaultSignCollideHooks4EAB40()
	hooks.classLow = func(obj *signCollideTestObject4EAB40) uint8 {
		events = append(events, "class")
		return obj.class
	}
	hooks.loadUse = func(obj *signCollideTestObject4EAB40) *signCollideTestUse4EAB40 {
		events = append(events, "use")
		return obj.use
	}
	hooks.callUse = func(use *signCollideTestUse4EAB40, target, source *signCollideTestObject4EAB40) int32 {
		events = append(events, "call")
		return int32(use.id)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Use callback did not fault")
		}
		if want := []string{"class", "use", "call"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	signCollide4EAB40(source, target, 0, hooks)
}
