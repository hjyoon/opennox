package server

import (
	"slices"
	"testing"

	"github.com/opennox/libs/object"
)

func TestImpEggDieNative54CAE0AudioThenNoCollide(t *testing.T) {
	obj := &Object{ObjFlags: object.FlagActive | object.FlagMarked}
	var events []string
	result := ImpEggDieNative54CAE0(obj, SimpleObjectDeathRuntime54CAE0{
		Audio: func(id uint32, got *Object) {
			events = append(events, "audio")
			if id != 764 || got != obj {
				t.Fatalf("audio = %d/%p, want 764/%p", id, got, obj)
			}
			if got.ObjFlags.Has(object.FlagNoCollide) {
				t.Fatal("NO_COLLIDE was set before the audio event")
			}
		},
	})
	wantFlags := object.FlagActive | object.FlagMarked | object.FlagNoCollide
	if obj.ObjFlags != wantFlags || result != uint32(wantFlags) {
		t.Fatalf("flags/result = %#x/%#x, want %#x", obj.ObjFlags, result, wantFlags)
	}
	if !slices.Equal(events, []string{"audio"}) {
		t.Fatalf("events = %v, want [audio]", events)
	}
}

func TestPotionDieNative54CBB0AudioThenDelete(t *testing.T) {
	obj := &Object{}
	var events []string
	PotionDieNative54CBB0(obj, SimpleObjectDeathRuntime54CAE0{
		Audio: func(id uint32, got *Object) {
			events = append(events, "audio")
			if id != 753 || got != obj {
				t.Fatalf("audio = %d/%p, want 753/%p", id, got, obj)
			}
		},
		DelayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != obj {
				t.Fatalf("deleted object = %p, want %p", got, obj)
			}
		},
	})
	if !slices.Equal(events, []string{"audio", "delete"}) {
		t.Fatalf("events = %v, want [audio delete]", events)
	}
}
