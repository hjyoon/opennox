package server

import (
	"reflect"
	"testing"
)

type ownCollideTestObject4EA2C0 struct {
	name   string
	class  uint32
	owner  *ownCollideTestObject4EA2C0
	frame  uint32
	marker uint32
}

func defaultOwnCollideHooks4EA2C0() ownCollideHooks4EA2C0[*ownCollideTestObject4EA2C0] {
	return ownCollideHooks4EA2C0[*ownCollideTestObject4EA2C0]{
		loadTargetClass: func(obj *ownCollideTestObject4EA2C0) uint32 {
			return obj.class
		},
		loadSourceOwner: func(obj *ownCollideTestObject4EA2C0) *ownCollideTestObject4EA2C0 {
			return obj.owner
		},
		loadFrame: func() uint32 { return 0 },
		storeSourceFrame: func(obj *ownCollideTestObject4EA2C0, frame uint32) {
			obj.frame = frame
		},
		setOwner: func(owner, obj *ownCollideTestObject4EA2C0) {
			obj.owner = owner
		},
	}
}

func TestOwnCollide4EA2C0SuccessOrderAndCachedOwner(t *testing.T) {
	source := &ownCollideTestObject4EA2C0{name: "source", frame: 0x11111111}
	target := &ownCollideTestObject4EA2C0{name: "target", class: 0x80000104}
	lateOwner := &ownCollideTestObject4EA2C0{name: "late-owner"}
	events := make([]string, 0, 5)
	hooks := defaultOwnCollideHooks4EA2C0()
	hooks.loadTargetClass = func(got *ownCollideTestObject4EA2C0) uint32 {
		events = append(events, "target-class")
		if got != target {
			t.Fatalf("class target = %p", got)
		}
		return got.class
	}
	hooks.loadSourceOwner = func(got *ownCollideTestObject4EA2C0) *ownCollideTestObject4EA2C0 {
		events = append(events, "source-owner")
		if got != source {
			t.Fatalf("owner source = %p", got)
		}
		return got.owner
	}
	hooks.loadFrame = func() uint32 {
		events = append(events, "frame")
		source.owner = lateOwner
		return 0xfedcba98
	}
	hooks.storeSourceFrame = func(got *ownCollideTestObject4EA2C0, frame uint32) {
		events = append(events, "store-frame")
		if got != source || frame != 0xfedcba98 {
			t.Fatalf("frame store = %p/%#x", got, frame)
		}
		got.frame = frame
	}
	hooks.setOwner = func(owner, obj *ownCollideTestObject4EA2C0) {
		events = append(events, "set-owner")
		if owner != target || obj != source {
			t.Fatalf("set owner = %p/%p", owner, obj)
		}
		if obj.frame != 0xfedcba98 {
			t.Fatalf("set owner observed frame = %#x", obj.frame)
		}
		if obj.owner != lateOwner {
			t.Fatal("source owner was reloaded after the frame callback")
		}
		obj.owner = owner
	}

	ownCollide4EA2C0(source, target, hooks)
	want := []string{"target-class", "source-owner", "frame", "store-frame", "set-owner"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.owner != target || source.frame != 0xfedcba98 {
		t.Fatalf("source = owner %p frame %#x", source.owner, source.frame)
	}
}

func TestOwnCollide4EA2C0EarlyReturns(t *testing.T) {
	source := &ownCollideTestObject4EA2C0{name: "source", frame: 0x11223344}
	player := &ownCollideTestObject4EA2C0{name: "player", class: 4}
	nonPlayer := &ownCollideTestObject4EA2C0{name: "non-player", class: 0x80000100}
	other := &ownCollideTestObject4EA2C0{name: "other"}
	tests := []struct {
		name   string
		target *ownCollideTestObject4EA2C0
		owner  *ownCollideTestObject4EA2C0
		want   []string
	}{
		{name: "nil target", target: nil, want: []string{}},
		{name: "non-player target", target: nonPlayer, want: []string{"class"}},
		{name: "already target-owned", target: player, owner: player, want: []string{"class", "owner"}},
		{name: "already other-owned", target: player, owner: other, want: []string{"class", "owner"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source.owner = tc.owner
			source.frame = 0x11223344
			events := make([]string, 0, 2)
			hooks := defaultOwnCollideHooks4EA2C0()
			hooks.loadTargetClass = func(obj *ownCollideTestObject4EA2C0) uint32 {
				events = append(events, "class")
				return obj.class
			}
			hooks.loadSourceOwner = func(obj *ownCollideTestObject4EA2C0) *ownCollideTestObject4EA2C0 {
				events = append(events, "owner")
				return obj.owner
			}
			hooks.loadFrame = func() uint32 {
				t.Fatal("early return loaded frame")
				return 0
			}
			hooks.storeSourceFrame = func(*ownCollideTestObject4EA2C0, uint32) {
				t.Fatal("early return stored frame")
			}
			hooks.setOwner = func(*ownCollideTestObject4EA2C0, *ownCollideTestObject4EA2C0) {
				t.Fatal("early return set owner")
			}

			ownCollide4EA2C0(source, tc.target, hooks)
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
			if source.owner != tc.owner || source.frame != 0x11223344 {
				t.Fatalf("source changed = owner %p frame %#x", source.owner, source.frame)
			}
		})
	}
}

func TestOwnCollide4EA2C0SourceFaultIsDelayedUntilPlayerTarget(t *testing.T) {
	hooks := defaultOwnCollideHooks4EA2C0()
	hooks.loadSourceOwner = func(obj *ownCollideTestObject4EA2C0) *ownCollideTestObject4EA2C0 {
		return obj.owner
	}
	ownCollide4EA2C0(nil, nil, hooks)
	ownCollide4EA2C0(nil, &ownCollideTestObject4EA2C0{class: 0x100}, hooks)

	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault after a Player target")
		}
	}()
	ownCollide4EA2C0(nil, &ownCollideTestObject4EA2C0{class: 4}, hooks)
}
