package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestAudioEventZoneNativeLayout501C00(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantPlayerUpdatePlayer := uintptr(276)
	wantPlayerZone := uintptr(3668)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantPlayerUpdatePlayer = 336
		wantPlayerZone = 4964
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerUpdatePlayer},
		{"MonsterUpdateData.Field0", unsafe.Offsetof(MonsterUpdateData{}.Field0), 0},
		{"Player.field3668", unsafe.Offsetof(Player{}.field3668), wantPlayerZone},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAudioEventZoneNative501C00PreservesPointersAndUnsignedZone(t *testing.T) {
	player := &Player{field3668: 0xffffffe7}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer | object.ClassMonster,
		UpdateData: unsafe.Pointer(update),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit":   unsafe.Pointer(unit),
			"update": unsafe.Pointer(update),
			"player": unsafe.Pointer(player),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}
	runtimeHooks := AudioEventZoneRuntime501C00{
		PolygonByID: func(uint32) unsafe.Pointer {
			t.Fatal("Player branch called PolygonByID")
			return nil
		},
		PolygonAtPoint: func([2]int32, uint32) unsafe.Pointer {
			t.Fatal("nonzero player zone accessed nil position")
			return nil
		},
		PolygonZone: func(unsafe.Pointer) uint8 {
			t.Fatal("Player branch called PolygonZone")
			return 0
		},
	}
	if got := audioEventZoneNative501C00(nil, unit, runtimeHooks); got != 0xe7 {
		t.Fatalf("zone = %#x, want 0xe7", got)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestAudioEventZoneNative501C00MonsterAndFallback(t *testing.T) {
	monsterUpdate := &MonsterUpdateData{Field0: 0xfedcba98}
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(monsterUpdate)}
	monsterPolygon := new(uintptr)
	fallbackPolygon := new(uintptr)
	position := types.Ptf(12.5, -7.5)
	var byID uint32
	var point [2]int32
	var previous uint32
	var zones []unsafe.Pointer
	runtimeHooks := AudioEventZoneRuntime501C00{
		PolygonByID: func(id uint32) unsafe.Pointer {
			byID = id
			return unsafe.Pointer(monsterPolygon)
		},
		PolygonAtPoint: func(got [2]int32, gotPrevious uint32) unsafe.Pointer {
			point, previous = got, gotPrevious
			return unsafe.Pointer(fallbackPolygon)
		},
		PolygonZone: func(polygon unsafe.Pointer) uint8 {
			zones = append(zones, polygon)
			if polygon == unsafe.Pointer(monsterPolygon) {
				return 0
			}
			if polygon == unsafe.Pointer(fallbackPolygon) {
				return 0xc3
			}
			t.Fatalf("unexpected polygon = %p", polygon)
			return 0
		},
	}
	got := audioEventZoneNative501C00(&position, unit, runtimeHooks)
	if got != 0xc3 || byID != monsterUpdate.Field0 || point != [2]int32{12, -8} || previous != 0 {
		t.Fatalf("result/byID/point/previous = %#x/%#x/%v/%#x", got, byID, point, previous)
	}
	wantZones := []unsafe.Pointer{unsafe.Pointer(monsterPolygon), unsafe.Pointer(fallbackPolygon)}
	if len(zones) != len(wantZones) || zones[0] != wantZones[0] || zones[1] != wantZones[1] {
		t.Fatalf("zones = %v, want %v", zones, wantZones)
	}
}

func TestAudioEventZoneNative501C00MissingMonsterPolygonSkipsZone(t *testing.T) {
	update := &MonsterUpdateData{Field0: 0xdeadface}
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	position := types.Ptf(1, 2)
	zoneCalls := 0
	got := audioEventZoneNative501C00(&position, unit, AudioEventZoneRuntime501C00{
		PolygonByID: func(id uint32) unsafe.Pointer {
			if id != update.Field0 {
				t.Fatalf("id = %#x, want %#x", id, update.Field0)
			}
			return nil
		},
		PolygonAtPoint: func(point [2]int32, previous uint32) unsafe.Pointer {
			if point != [2]int32{1, 2} || previous != 0 {
				t.Fatalf("point/previous = %v/%#x", point, previous)
			}
			return nil
		},
		PolygonZone: func(unsafe.Pointer) uint8 {
			zoneCalls++
			return 0
		},
	})
	if got != 0 || zoneCalls != 0 {
		t.Fatalf("result/zone calls = %#x/%d, want 0/0", got, zoneCalls)
	}
}

func TestAudioEventZoneNative501C00RetainsOracleFaults(t *testing.T) {
	assertPanic := func(t *testing.T, call func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("expected native layout fault")
			}
		}()
		call()
	}
	position := types.Ptf(0, 0)
	emptyRuntime := AudioEventZoneRuntime501C00{
		PolygonByID:    func(uint32) unsafe.Pointer { return nil },
		PolygonAtPoint: func([2]int32, uint32) unsafe.Pointer { return nil },
		PolygonZone:    func(unsafe.Pointer) uint8 { return 0 },
	}
	t.Run("nil player update", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		assertPanic(t, func() { audioEventZoneNative501C00(&position, unit, emptyRuntime) })
	})
	t.Run("nil player", func(t *testing.T) {
		update := &PlayerUpdateData{}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		assertPanic(t, func() { audioEventZoneNative501C00(&position, unit, emptyRuntime) })
	})
	t.Run("nil monster update", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassMonster}
		assertPanic(t, func() { audioEventZoneNative501C00(&position, unit, emptyRuntime) })
	})
	t.Run("nil fallback position", func(t *testing.T) {
		assertPanic(t, func() { audioEventZoneNative501C00(nil, nil, emptyRuntime) })
	})
}

func TestAudioEventZoneFloatToInt501C00(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0, 0},
		{0.5, 0},
		{1.5, 2},
		{2.5, 2},
		{-1.5, -2},
		{float32(math.NaN()), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{2147483648, math.MinInt32},
	}
	for _, test := range tests {
		if got := audioEventZoneFloatToInt501C00(test.value); got != test.want {
			t.Errorf("convert(%v) = %d, want %d", test.value, got, test.want)
		}
	}
}
