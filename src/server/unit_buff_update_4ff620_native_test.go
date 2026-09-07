package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnitBuffUpdateNative4FF620ObjectLayouts(t *testing.T) {
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantBuffs := uintptr(340)
	wantDuration := uintptr(344)
	wantPower := uintptr(408)
	wantObj130 := uintptr(520)
	wantDamageType := uintptr(524)
	wantSpeed := uintptr(544)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFlags = 20
		wantBuffs = 344
		wantDuration = 348
		wantPower = 412
		wantObj130 = 576
		wantDamageType = 584
		wantSpeed = 604
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ObjClass offset", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"ObjClass width", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"ObjFlags offset", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"ObjFlags width", unsafe.Sizeof(Object{}.ObjFlags), 4},
		{"Buffs offset", unsafe.Offsetof(Object{}.Buffs), wantBuffs},
		{"Buffs width", unsafe.Sizeof(Object{}.Buffs), 4},
		{"BuffsDur offset", unsafe.Offsetof(Object{}.BuffsDur), wantDuration},
		{"BuffsDur element width", unsafe.Sizeof(Object{}.BuffsDur[0]), 2},
		{"BuffsPower offset", unsafe.Offsetof(Object{}.BuffsPower), wantPower},
		{"BuffsPower element width", unsafe.Sizeof(Object{}.BuffsPower[0]), 1},
		{"Obj130 offset", unsafe.Offsetof(Object{}.Obj130), wantObj130},
		{"Obj130 width", unsafe.Sizeof(Object{}.Obj130), unsafe.Sizeof(uintptr(0))},
		{"Field131 offset", unsafe.Offsetof(Object{}.Field131), wantDamageType},
		{"Field131 width", unsafe.Sizeof(Object{}.Field131), 4},
		{"SpeedCur offset", unsafe.Offsetof(Object{}.SpeedCur), wantSpeed},
		{"SpeedCur width", unsafe.Sizeof(Object{}.SpeedCur), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Object %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestUnitBuffUpdateNative4FF620PreservesPointerAndAccessorOrder(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	marker, freeMarker := alloc.New(Object{})
	t.Cleanup(freeMarker)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(marker)) <= math.MaxUint32) {
		t.Fatalf("native pointers = %p/%p, want both above 4 GiB", unit, marker)
	}
	unit.Buffs = uint32(1) << 16
	unit.BuffsDur[16] = 1
	unit.BuffsPower[16] = 0xff
	unit.Obj130 = marker
	unit.Field131 = math.MaxUint32
	unit.ObjClass = object.ClassPlayer | object.Class(0x80000000)
	unit.SpeedCur = -3.5

	var events []string
	buffLoads := 0
	checkUnit := func(label string, got *Object) {
		if got != unit {
			t.Fatalf("%s unit = %p, want full native identity %p", label, got, unit)
		}
	}
	deps := unitBuffUpdateNativeDeps4FF620{
		loadUnitArg: func(got *Object) *Object {
			events = append(events, "unit")
			checkUnit("load arg", got)
			return got
		},
		loadBuffs: func(got *Object) uint32 {
			buffLoads++
			checkUnit("load buffs", got)
			return got.Buffs
		},
		loadFPS: func() uint32 {
			events = append(events, "fps")
			return 60
		},
		loadDuration: func(got *Object, buff int32) uint16 {
			events = append(events, fmt.Sprintf("duration-%d", buff))
			checkUnit("load duration", got)
			return got.BuffsDur[int(buff)]
		},
		storeDuration: func(got *Object, buff int32, value uint16) {
			events = append(events, fmt.Sprintf("store-duration-%d", buff))
			checkUnit("store duration", got)
			got.BuffsDur[int(buff)] = value
		},
		audio: func(id int32, got *Object, kind, code int32) {
			events = append(events, fmt.Sprintf("audio-%d", id))
			checkUnit("audio", got)
			if kind != 0 || code != 0 {
				t.Fatalf("audio scalars = %d/%d, want zero/zero", kind, code)
			}
		},
		loadFlags: func(*Object) uint32 {
			t.Fatal("death slot loaded flags")
			return 0
		},
		storeFlags: func(*Object, uint32) { t.Fatal("death slot stored flags") },
		storeObj130: func(got, value *Object) {
			events = append(events, "obj130")
			checkUnit("store Obj130", got)
			if value != nil {
				t.Fatalf("Obj130 value = %p, want native nil", value)
			}
			got.Obj130 = value
		},
		storeDamageType: func(got *Object, value uint32) {
			events = append(events, "damage-type")
			checkUnit("store damage type", got)
			if value != 13 {
				t.Fatalf("damage type = %d, want 13", value)
			}
			got.Field131 = value
		},
		damageClear: func(got *Object, damage int32) {
			events = append(events, "damage")
			checkUnit("damage", got)
			if damage != 9999999 {
				t.Fatalf("damage = %d, want 9999999", damage)
			}
		},
		loadClassLow: func(got *Object) uint8 {
			events = append(events, "class")
			checkUnit("load class", got)
			return uint8(got.ObjClass)
		},
		incrementElimDeath: func(got *Object) {
			events = append(events, "increment")
			checkUnit("increment", got)
		},
		reportLesson: func(got *Object) {
			events = append(events, "report")
			checkUnit("report", got)
		},
		buffOff: func(got *Object, buff int32) int32 {
			events = append(events, fmt.Sprintf("buff-off-%d", buff))
			checkUnit("buff off", got)
			return math.MinInt32
		},
		storePower: func(got *Object, buff int32, value uint8) {
			events = append(events, fmt.Sprintf("power-%d", buff))
			checkUnit("store power", got)
			got.BuffsPower[int(buff)] = value
		},
		testBuff: func(got *Object, buff int32) int32 {
			events = append(events, fmt.Sprintf("test-%d", buff))
			checkUnit("test buff", got)
			return math.MinInt32
		},
		loadSpeed: func(got *Object) float32 {
			events = append(events, "speed")
			checkUnit("load speed", got)
			return got.SpeedCur
		},
		storeSpeed: func(got *Object, value float32) {
			events = append(events, "store-speed")
			checkUnit("store speed", got)
			got.SpeedCur = value
		},
	}

	unitBuffUpdateNative4FF620(unit, deps)
	want := []string{
		"unit", "fps", "duration-16", "duration-16", "store-duration-16",
		"obj130", "damage-type", "damage", "audio-779", "class", "increment",
		"report", "buff-off-16", "power-16", "test-9", "speed", "store-speed",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want exact native accessor order %q", events, want)
	}
	if buffLoads != 33 {
		t.Fatalf("Buffs loads = %d, want initial gate plus 32 live loop loads", buffLoads)
	}
	if unit.Obj130 != nil || unit.Field131 != 13 || unit.BuffsDur[16] != 0 || unit.BuffsPower[16] != 0 || unit.SpeedCur != -4.375 {
		t.Fatalf("native state = obj130:%p type:%d duration:%d power:%d speed:%g", unit.Obj130, unit.Field131, unit.BuffsDur[16], unit.BuffsPower[16], unit.SpeedCur)
	}
	runtime.KeepAlive(marker)
	runtime.KeepAlive(unit)
}

func TestUnitBuffUpdate4FF620ServerBindingFlagAndSlowPaths(t *testing.T) {
	var s Server
	s.SetTickRate(60)
	unit := &Object{
		ObjClass: object.ClassClientPersist,
		ObjFlags: object.Flags(0xa5b6c7ff),
		Buffs:    uint32(1)<<7 | uint32(1)<<9,
		SpeedCur: 8,
	}
	unit.BuffsDur[7] = 1
	unit.BuffsPower[7] = 0xee
	unit.BuffsDur[9] = 0

	var buffOff []EnchantID
	s.UnitBuffUpdate4FF620(unit, UnitBuffUpdateRuntime4FF620{
		DamageClear:        func(*Object, int32) { t.Fatal("flag expiration applied death damage") },
		IncrementElimDeath: func(*Object) { t.Fatal("flag expiration incremented a death") },
		BuffOff: func(got *Object, buff EnchantID) {
			if got != unit {
				t.Fatalf("BuffOff unit = %p, want %p", got, unit)
			}
			buffOff = append(buffOff, buff)
			got.Buffs &^= uint32(1) << uint32(buff)
			got.BuffsDur[int(buff)] = 0
			got.BuffsPower[int(buff)] = 0xdd
		},
	})
	if !reflect.DeepEqual(buffOff, []EnchantID{7}) {
		t.Fatalf("BuffOff calls = %v, want [7]", buffOff)
	}
	if unit.ObjFlags != object.Flags(0xa5b6c7bf) || unit.BuffsPower[7] != 0 || unit.BuffsDur[7] != 0 {
		t.Fatalf("flag state = flags:%08x duration:%04x power:%02x", uint32(unit.ObjFlags), unit.BuffsDur[7], unit.BuffsPower[7])
	}
	if unit.SpeedCur != 10 {
		t.Fatalf("buff-9 speed = %g, want 10", unit.SpeedCur)
	}
}

func TestUnitBuffUpdate4FF620ServerBindingDeathMetadataAndAudio(t *testing.T) {
	var s Server
	s.SetTickRate(60)
	marker, freeMarker := alloc.New(Object{})
	t.Cleanup(freeMarker)
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	unit.ObjClass = object.ClassClientPersist
	unit.Buffs = uint32(1) << 16
	unit.Obj130 = marker
	unit.Field131 = math.MaxUint32
	unit.BuffsDur[16] = 1
	unit.BuffsPower[16] = 0xff

	damageCalls := 0
	s.UnitBuffUpdate4FF620(unit, UnitBuffUpdateRuntime4FF620{
		DamageClear: func(got *Object, damage int32) {
			damageCalls++
			if got != unit || damage != 9999999 || unit.Obj130 != nil || unit.Field131 != 13 {
				t.Fatalf("damage callback = unit:%p damage:%d metadata:%p/%d", got, damage, unit.Obj130, unit.Field131)
			}
		},
		IncrementElimDeath: func(*Object) { t.Fatal("non-player incremented elimination death") },
		BuffOff: func(got *Object, buff EnchantID) {
			if got != unit || buff != 16 {
				t.Fatalf("BuffOff = %p/%d, want %p/16", got, buff, unit)
			}
			got.Buffs = 0
			got.BuffsPower[16] = 0xee
		},
	})
	if damageCalls != 1 || unit.Obj130 != nil || unit.Field131 != 13 || unit.BuffsPower[16] != 0 {
		t.Fatalf("death state = calls:%d obj130:%p type:%d power:%d", damageCalls, unit.Obj130, unit.Field131, unit.BuffsPower[16])
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio events = %d, want expiration sound", len(s.Audio.delayedObj))
	}
	event := s.Audio.delayedObj[0]
	if event.ID != 779 || event.Obj != unit || event.Kind != 0 || event.Code != 0 {
		t.Fatalf("queued audio = %#v, want sound 779 for %p", event, unit)
	}
	runtime.KeepAlive(marker)
	runtime.KeepAlive(unit)
}

func TestUnitBuffUpdate4FF620NativeNilObjectFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil native object returned instead of faulting at initial Buffs load")
		}
	}()
	(&Server{}).UnitBuffUpdate4FF620(nil, UnitBuffUpdateRuntime4FF620{})
}
