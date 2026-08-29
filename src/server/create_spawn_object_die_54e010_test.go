package server

import (
	"math"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestCreateSpawnObjectDeathNativeLayouts54E010(t *testing.T) {
	wantFlags := uintptr(16)
	wantPosition := uintptr(56)
	wantDeath := uintptr(724)
	wantDeathData := uintptr(728)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags = 20
		wantPosition = 60
		wantDeath = 824
		wantDeathData = 832
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"death data size", unsafe.Sizeof(CreateSpawnObjectDeathData54E010{}), 132},
		{"death data TypeID", unsafe.Offsetof(CreateSpawnObjectDeathData54E010{}.TypeID), 0},
		{"death data Sound", unsafe.Offsetof(CreateSpawnObjectDeathData54E010{}.Sound), 128},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.Death", unsafe.Offsetof(Object{}.Death), wantDeath},
		{"Object.DeathData", unsafe.Offsetof(Object{}.DeathData), wantDeathData},
		{"Object.Death width", unsafe.Sizeof(Object{}.Death), unsafe.Sizeof(uintptr(0))},
		{"Object.DeathData width", unsafe.Sizeof(Object{}.DeathData), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestCreateObjectDieNative54E010PreservesCachedDataAndLiveLoads(t *testing.T) {
	data := new(CreateSpawnObjectDeathData54E010)
	copy(data.TypeID[:], "SmallFist")
	data.Sound = 611
	replacement := new(CreateSpawnObjectDeathData54E010)
	replacement.Sound = 999
	created := new(Object)
	source := &Object{
		ObjFlags:  object.FlagActive,
		PosVec:    types.Pointf{X: 11.5, Y: -7.25},
		DeathData: unsafe.Pointer(data),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(source), unsafe.Pointer(data), unsafe.Pointer(replacement), unsafe.Pointer(created), source.DeathData,
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want a native high address", index, pointer)
			}
		}
	}

	var events []string
	CreateObjectDieNative54E010(source, CreateSpawnObjectDeathRuntime54E010{
		NewObjectByTypeID: func(id string) *Object {
			events = append(events, "new:"+id)
			data.Sound = 612
			source.DeathData = unsafe.Pointer(replacement)
			source.PosVec = types.Pointf{X: 31.25, Y: 42.5}
			return created
		},
		CreateAt: func(obj *Object, pos types.Pointf) {
			events = append(events, "create")
			if obj != created || pos != (types.Pointf{X: 31.25, Y: 42.5}) {
				t.Fatalf("created object/position = %p/%+v", obj, pos)
			}
		},
		Audio: func(id uint32, obj *Object) {
			events = append(events, "audio")
			if id != 612 || obj != source {
				t.Fatalf("audio = %d/%p, want 612/%p", id, obj, source)
			}
		},
		DelayedDelete: func(obj *Object) {
			events = append(events, "delete")
			if obj != source {
				t.Fatalf("deleted object = %p, want %p", obj, source)
			}
		},
	})
	wantEvents := []string{"new:SmallFist", "create", "audio", "delete"}
	if !slices.Equal(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if source.ObjFlags != object.FlagActive {
		t.Fatalf("CreateObjectDie flags = %#x, want ACTIVE", source.ObjFlags)
	}
	runtime.KeepAlive(data)
	runtime.KeepAlive(replacement)
	runtime.KeepAlive(created)
}

func TestCreateObjectDieNative54E010NilCreationStillDeletes(t *testing.T) {
	data := new(CreateSpawnObjectDeathData54E010)
	copy(data.TypeID[:], "MissingObject")
	source := &Object{DeathData: unsafe.Pointer(data)}
	var events []string
	CreateObjectDieNative54E010(source, CreateSpawnObjectDeathRuntime54E010{
		NewObjectByTypeID: func(id string) *Object {
			events = append(events, "new:"+id)
			return nil
		},
		CreateAt: func(*Object, types.Pointf) {
			t.Fatal("CreateAt called for a nil factory result")
		},
		Audio: func(uint32, *Object) {
			t.Fatal("zero sound produced an audio event")
		},
		DelayedDelete: func(obj *Object) {
			events = append(events, "delete")
			if obj != source {
				t.Fatalf("deleted object = %p, want %p", obj, source)
			}
		},
	})
	if want := []string{"new:MissingObject", "delete"}; !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpawnObjectDieNative54E070SetsOnlyDeadAndReturnsLowWord(t *testing.T) {
	data := new(CreateSpawnObjectDeathData54E010)
	copy(data.TypeID[:], "Smoke")
	source := &Object{
		ObjFlags:  object.FlagActive | object.FlagMarked,
		DeathData: unsafe.Pointer(data),
	}
	var events []string
	got := SpawnObjectDieNative54E070(source, CreateSpawnObjectDeathRuntime54E010{
		NewObjectByTypeID: func(id string) *Object {
			events = append(events, "new:"+id)
			return nil
		},
		CreateAt: func(*Object, types.Pointf) {
			t.Fatal("CreateAt called for a nil factory result")
		},
		Audio: func(uint32, *Object) {
			t.Fatal("zero sound produced an audio event")
		},
		DelayedDelete: func(*Object) {
			t.Fatal("SpawnObjectDie scheduled deletion")
		},
	})
	wantFlags := object.FlagActive | object.FlagMarked | object.FlagDead
	if source.ObjFlags != wantFlags || source.ObjFlags.Has(object.FlagDestroyed) {
		t.Fatalf("SpawnObjectDie flags = %#x, want %#x without DESTROYED", source.ObjFlags, wantFlags)
	}
	if want := int16(uint16(wantFlags)); got != want {
		t.Fatalf("SpawnObjectDie result = %d, want low word %d", got, want)
	}
	if want := []string{"new:Smoke"}; !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
