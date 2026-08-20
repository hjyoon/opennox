package server

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"unsafe"
)

func TestBoltDamageNative4EF1E0BindsNamedModifierFields(t *testing.T) {
	cache := uint32(0)
	var events []string
	modifier := &Modifier{
		TypeInd:              73,
		ReqStrength60:        12,
		DamageCoeffOrArmor64: math.Float32frombits(0x3fa00000),
		DamageMin72:          0xffff,
	}
	got := boltDamageNative4EF1E0(37, modifier, boltDamageNativeDeps4EF1E0{
		loadCachedArcherBoltType: func() uint32 {
			events = append(events, "cache")
			return cache
		},
		lookupType: func(name string) uint32 {
			events = append(events, "lookup:"+name)
			return 73
		},
		storeCachedArcherBoltType: func(value uint32) {
			events = append(events, "store")
			cache = value
		},
		gameFlagsCheck: func(mask uint32) int32 {
			events = append(events, "flag")
			if mask != 0x800 {
				t.Fatalf("flag mask = %#x, want 0x800", mask)
			}
			return 1
		},
		balanceFloat: func(key string) float64 {
			events = append(events, "balance:"+key)
			return 5.5
		},
	})
	if got != 36.75 {
		t.Fatalf("damage = %v, want 36.75", got)
	}
	wantEvents := []string{
		"cache", "lookup:ArcherBolt", "store", "flag", "cache", "balance:BoltSoloDamageMin",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("dependency events = %q, want %q", events, wantEvents)
	}
	if cache != 73 {
		t.Fatalf("cache = %d, want 73", cache)
	}
}

func TestBoltDamageNative4EF1E0NormalPathSkipsTypeServices(t *testing.T) {
	modifier := &Modifier{
		TypeInd:              73,
		ReqStrength60:        40,
		DamageCoeffOrArmor64: 0.75,
		DamageMin72:          9,
	}
	got := boltDamageNative4EF1E0(20, modifier, boltDamageNativeDeps4EF1E0{
		loadCachedArcherBoltType: func() uint32 { return 73 },
		lookupType: func(string) uint32 {
			t.Fatal("nonzero cache performed lookup")
			return 0
		},
		storeCachedArcherBoltType: func(uint32) {
			t.Fatal("nonzero cache performed store")
		},
		gameFlagsCheck: func(uint32) int32 { return 0 },
		balanceFloat: func(string) float64 {
			t.Fatal("normal path loaded balance minimum")
			return 0
		},
	})
	if got != -6 {
		t.Fatalf("damage = %v, want -6", got)
	}
}

func TestBoltDamageModifierLayout4EF1E0(t *testing.T) {
	wantSize := uintptr(88)
	wantType := uintptr(4)
	wantRequired := uintptr(60)
	wantCoefficient := uintptr(64)
	wantMinimum := uintptr(72)
	if strconv.IntSize == 64 {
		wantSize = 112
		wantType = 8
		wantRequired = 72
		wantCoefficient = 76
		wantMinimum = 84
	}
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "size", got: unsafe.Sizeof(Modifier{}), want: wantSize},
		{name: "TypeInd", got: unsafe.Offsetof(Modifier{}.TypeInd), want: wantType},
		{name: "ReqStrength60", got: unsafe.Offsetof(Modifier{}.ReqStrength60), want: wantRequired},
		{name: "DamageCoeffOrArmor64", got: unsafe.Offsetof(Modifier{}.DamageCoeffOrArmor64), want: wantCoefficient},
		{name: "DamageMin72", got: unsafe.Offsetof(Modifier{}.DamageMin72), want: wantMinimum},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("layout = %d, want %d", test.got, test.want)
			}
		})
	}
}
