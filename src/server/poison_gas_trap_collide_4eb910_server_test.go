package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestPoisonGasTrapCollide4EB910NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantPos := uintptr(56)
	wantOwner := uintptr(508)
	wantUpdateData := uintptr(748)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantPos = 60
		wantOwner = 552
		wantUpdateData = 872
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.PosVec.Y", unsafe.Offsetof(Object{}.PosVec) + unsafe.Offsetof(types.Pointf{}.Y), wantPos + 4},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"ToxicCloudUpdateData size", unsafe.Sizeof(ToxicCloudUpdateData{}), 4},
		{"ToxicCloudUpdateData.Duration", unsafe.Offsetof(ToxicCloudUpdateData{}.Duration), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPoisonGasTrapCollideNative4EB910UsesLiveNativeFieldsAndEffects(t *testing.T) {
	owner := &Object{}
	source := &Object{PosVec: types.Pointf{X: 12.5, Y: -4.25}, ObjOwner: owner}
	target := &Object{}
	entryData := &ToxicCloudUpdateData{Duration: -1}
	liveData := &ToxicCloudUpdateData{Duration: -2}
	cloud := &Object{UpdateData: unsafe.Pointer(entryData)}
	events := make([]string, 0, 8)

	poisonGasTrapCollideNative4EB910(source, target, nil, poisonGasTrapCollideNativeDeps4EB910{
		allowed: func(gotSource, gotTarget *Object) int32 {
			if gotSource != source || gotTarget != target {
				t.Fatalf("gate objects = (%p,%p), want (%p,%p)", gotSource, gotTarget, source, target)
			}
			events = append(events, "allowed")
			return 1
		},
		newObject: func(name string) *Object {
			if name != poisonGasTrapCloudType4EB910 {
				t.Fatalf("object type = %q, want %q", name, poisonGasTrapCloudType4EB910)
			}
			events = append(events, "new")
			return cloud
		},
		createAt: func(gotCloud, gotOwner *Object, pos types.Pointf, reserved uint32) {
			if gotCloud != cloud || gotOwner != owner || pos != source.PosVec || reserved != 0 {
				t.Fatalf("create = (%p,%p,%+v,%d), want (%p,%p,%+v,0)", gotCloud, gotOwner, pos, reserved, cloud, owner, source.PosVec)
			}
			events = append(events, "create")
			cloud.UpdateData = unsafe.Pointer(liveData)
		},
		loadLifetime: func(key string) float32 {
			if key != poisonGasTrapLifetime4EB910 {
				t.Fatalf("balance key = %q, want %q", key, poisonGasTrapLifetime4EB910)
			}
			events = append(events, "lifetime")
			return 2.5
		},
		loadFPS: func() uint32 {
			events = append(events, "fps")
			return 30
		},
		multiply: poisonGasTrapMultiply4EB910,
		floatToInt: func(value float32) int32 {
			events = append(events, "round")
			return poisonGasTrapRound4EB910(value)
		},
		audio: func(id uint32, got *Object, kind int32, code uint32) {
			events = append(events, "audio")
			if id != poisonGasTrapTriggered4EB910 || got != source || kind != 0 || code != 0 {
				t.Fatalf("audio = (%d,%p,%d,%d)", id, got, kind, code)
			}
			if liveData.Duration != 75 {
				t.Fatalf("duration at audio = %d, want 75", liveData.Duration)
			}
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("deleted object = %p, want %p", got, source)
			}
		},
	})

	want := []string{"allowed", "new", "create", "lifetime", "fps", "round", "audio", "delete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if entryData.Duration != -1 || liveData.Duration != 75 {
		t.Fatalf("duration stores = entry %d, live %d", entryData.Duration, liveData.Duration)
	}
}

func TestPoisonGasTrapCollide4EB910RecoveredSound(t *testing.T) {
	if sound.ID(poisonGasTrapTriggered4EB910) != sound.SoundPoisonTrapTriggered {
		t.Fatalf("sound = %d, want %d", poisonGasTrapTriggered4EB910, sound.SoundPoisonTrapTriggered)
	}
}
