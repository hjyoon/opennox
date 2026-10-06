package server

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func toxicCloudCastAlloc52DB60[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("native allocation %p must be above 4 GiB", ptr)
	}
	return ptr
}

func toxicCloudCastTestDeps52DB60(t *testing.T, cloud *Object, cache *uint32) toxicCloudCastDeps52DB60 {
	t.Helper()
	return toxicCloudCastDeps52DB60{
		loadCache:  func() uint32 { return *cache },
		storeCache: func(kind uint32) { *cache = kind },
		lookupType: func(key string) uint32 {
			if key != "ToxicCloud" {
				t.Fatalf("type key=%q", key)
			}
			return 11
		},
		traceRay:  func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		newObject: func(uint32) *Object { return cloud },
		createAt:  func(*Object, *Object, types.Pointf) {},
		lifetime: func(key string) float32 {
			if key != "ToxicCloudLifetime" {
				t.Fatalf("lifetime key=%q", key)
			}
			return 1.5
		},
		fps:       func() uint32 { return 31 },
		castSound: func(int32) sound.ID { return 123 },
		audio:     func(sound.ID, *Object, int, uint32) {},
		inform:    func(uint8, byte, int32) { t.Fatal("unexpected player notification") },
	}
}

func TestToxicCloudCastNativePointersAndLiveOrder52DB60(t *testing.T) {
	second := toxicCloudCastAlloc52DB60(t, Object{Field29: 123})
	owner := toxicCloudCastAlloc52DB60(t, Object{Field129: second, PosVec: types.Ptf(-1, -2)})
	caster := toxicCloudCastAlloc52DB60(t, Object{PosVec: types.Ptf(10, 20)})
	arg := toxicCloudCastAlloc52DB60(t, SpellAcceptArg{Obj: second, Pos: types.Ptf(30, 40)})
	data := toxicCloudCastAlloc52DB60(t, ToxicCloudUpdateData{Duration: -123})
	replacement := toxicCloudCastAlloc52DB60(t, ToxicCloudUpdateData{Duration: -456})
	cloud := toxicCloudCastAlloc52DB60(t, Object{UpdateData: unsafe.Pointer(data), Field5: 0x1234, ObjFlags: 0x100, Field29: 0x12345678})
	beforeOwner, beforeSecond := *owner, *second
	var cache uint32
	var events []string
	h := toxicCloudCastTestDeps52DB60(t, cloud, &cache)
	h.loadCache = func() uint32 { events = append(events, "load"); return cache }
	h.storeCache = func(kind uint32) { events = append(events, "store"); cache = kind }
	h.lookupType = func(key string) uint32 { events = append(events, "type:"+key); return 11 }
	h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
		events = append(events, "trace")
		if from != types.Ptf(10, 20) || to != types.Ptf(30, 40) || flags != 9 || cache != 11 {
			t.Fatalf("trace=%v/%v/%d cache=%d", from, to, flags, cache)
		}
		cache = 0x8001000b
		caster.PosVec, arg.Pos = types.Ptf(50, 60), types.Ptf(70, 80)
		return true
	}
	h.newObject = func(kind uint32) *Object {
		events = append(events, "new")
		if kind != 0x8001000b {
			t.Fatalf("live DWORD allocation=%#x", kind)
		}
		arg.Pos = types.Ptf(90, 100)
		return cloud
	}
	h.createAt = func(got, by *Object, position types.Pointf) {
		events = append(events, "create")
		if got != cloud || by != owner || by == caster || position != types.Ptf(90, 100) || data.Duration != -123 {
			t.Fatalf("placement=%p/%p/%v duration=%d", got, by, position, data.Duration)
		}
		cloud.UpdateData = unsafe.Pointer(replacement)
	}
	h.lifetime = func(key string) float32 {
		events = append(events, "lifetime:"+key)
		if data.Duration != -123 || replacement.Duration != -456 {
			t.Fatal("duration written before balance")
		}
		return 1.5
	}
	h.fps = func() uint32 { events = append(events, "fps"); return 31 }
	h.castSound = func(id int32) sound.ID {
		events = append(events, "sound")
		if id != 128 || data.Duration != 46 || replacement.Duration != -456 {
			t.Fatalf("sound prefix=%d/%d/%d", id, data.Duration, replacement.Duration)
		}
		return 321
	}
	h.audio = func(id sound.ID, got *Object, kind int, code uint32) {
		events = append(events, "audio")
		if id != 321 || got != cloud || kind != 0 || code != 0 {
			t.Fatalf("object audio=%d/%p/%d/%d", id, got, kind, code)
		}
	}
	got := toxicCloudCast52DB60(128, owner, caster, arg, h)
	want := []string{"load", "type:ToxicCloud", "store", "trace", "load", "new", "create", "lifetime:ToxicCloudLifetime", "fps", "sound", "audio"}
	if got != 1 || !reflect.DeepEqual(events, want) || *owner != beforeOwner || *second != beforeSecond || arg.Obj != second ||
		cloud.Field5 != 0x1234 || cloud.ObjFlags != 0x100 || cloud.Field29 != 0x12345678 {
		t.Fatalf("result/order/untouched=%d/%v/%+v", got, events, cloud)
	}
}

func TestToxicCloudCastAllocationAndCache52DB60(t *testing.T) {
	for _, cache := range []uint32{11, 0x1000b, 0x7fffffff, 0x80000000, 0xffffffff} {
		for _, allocated := range []bool{false, true} {
			t.Run(fmt.Sprintf("cache-%08x/allocated-%t", cache, allocated), func(t *testing.T) {
				data := &ToxicCloudUpdateData{Duration: -12}
				cloud := &Object{UpdateData: unsafe.Pointer(data)}
				arg := &SpellAcceptArg{Pos: types.Ptf(1, 2)}
				before := *arg
				h := toxicCloudCastTestDeps52DB60(t, cloud, &cache)
				h.lookupType = func(string) uint32 { t.Fatal("unexpected type lookup"); return 0 }
				h.newObject = func(got uint32) *Object {
					if got != cache {
						t.Fatalf("cache truncated=%#x/%#x", got, cache)
					}
					if !allocated {
						return nil
					}
					return cloud
				}
				create, lifetime, fps, soundCalls, audio := 0, 0, 0, 0, 0
				h.createAt = func(got, owner *Object, position types.Pointf) {
					create++
					if got != cloud || owner != nil || position != arg.Pos {
						t.Fatal("wrong placement")
					}
				}
				h.lifetime = func(string) float32 { lifetime++; return 1.5 }
				h.fps = func() uint32 { fps++; return 31 }
				h.castSound = func(id int32) sound.ID {
					soundCalls++
					if id != 128 {
						t.Fatal("wrong spell")
					}
					return 321
				}
				h.audio = func(id sound.ID, got *Object, kind int, code uint32) {
					audio++
					if id != 321 || kind != 0 || code != 0 || allocated && got != cloud || !allocated && got != nil {
						t.Fatal("wrong allocation-failure audio")
					}
				}
				got := toxicCloudCast52DB60(128, nil, &Object{}, arg, h)
				wantCalls, wantDuration := 0, int32(-12)
				if allocated {
					wantCalls, wantDuration = 1, 46
				}
				if got != 1 || create != wantCalls || lifetime != wantCalls || fps != wantCalls || soundCalls != 1 || audio != 1 || data.Duration != wantDuration || *arg != before {
					t.Fatalf("result/create/lifetime/fps/sound/audio/duration=%d/%d/%d/%d/%d/%d/%d", got, create, lifetime, fps, soundCalls, audio, data.Duration)
				}
			})
		}
	}
}

func TestToxicCloudCastBlockedLivePlayer52DB60(t *testing.T) {
	for _, initialPlayer := range []bool{false, true} {
		for _, livePlayer := range []bool{false, true} {
			for _, index := range []uint8{0, 31, 255} {
				t.Run(fmt.Sprintf("initial-%t/live-%t/index-%d", initialPlayer, livePlayer, index), func(t *testing.T) {
					player := toxicCloudCastAlloc52DB60(t, Player{PlayerInd: index})
					data := toxicCloudCastAlloc52DB60(t, PlayerUpdateData{Player: player})
					caster := toxicCloudCastAlloc52DB60(t, Object{ObjClass: object.ClassMonster})
					if initialPlayer {
						caster.ObjClass = object.ClassPlayer
					}
					arg := toxicCloudCastAlloc52DB60(t, SpellAcceptArg{Pos: types.Ptf(30, 40)})
					cache := uint32(11)
					h := toxicCloudCastTestDeps52DB60(t, nil, &cache)
					h.traceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool {
						caster.ObjClass = object.ClassMonster
						if livePlayer {
							caster.ObjClass = object.ClassPlayer
						}
						caster.UpdateData = unsafe.Pointer(data)
						return false
					}
					h.newObject = func(uint32) *Object { t.Fatal("blocked allocation"); return nil }
					h.castSound = func(int32) sound.ID { t.Fatal("blocked sound"); return 0 }
					calls := 0
					h.inform = func(got uint8, code byte, value int32) {
						calls++
						if got != index || code != 0 || value != 2 {
							t.Fatalf("inform=%d/%d/%d", got, code, value)
						}
					}
					beforePlayer, beforeData, beforeArg := *player, *data, *arg
					got := toxicCloudCast52DB60(128, nil, caster, arg, h)
					wantCalls := 0
					if livePlayer {
						wantCalls = 1
					}
					if got != 0 || calls != wantCalls || cache != 11 || *player != beforePlayer || *data != beforeData || *arg != beforeArg {
						t.Fatalf("blocked result=%d calls=%d", got, calls)
					}
				})
			}
		}
	}
}

func TestToxicCloudCastZeroCacheRetry52DB60(t *testing.T) {
	var cache uint32
	h := toxicCloudCastTestDeps52DB60(t, nil, &cache)
	lookups, stores, traces := 0, 0, 0
	h.lookupType = func(string) uint32 { lookups++; return 0 }
	h.storeCache = func(value uint32) { stores++; cache = value }
	h.traceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool { traces++; return false }
	for range 2 {
		if got := toxicCloudCast52DB60(128, nil, &Object{}, &SpellAcceptArg{}, h); got != 0 {
			t.Fatal(got)
		}
	}
	if cache != 0 || lookups != 2 || stores != 2 || traces != 2 {
		t.Fatalf("retry=%d/%d/%d/%d", cache, lookups, stores, traces)
	}
}

func TestToxicCloudCastFaultPrefix52DB60(t *testing.T) {
	for _, fault := range []string{"caster", "argument", "cloud-update", "player-update", "player"} {
		t.Run(fault, func(t *testing.T) {
			var events []string
			var cache uint32
			cloud := &Object{}
			caster, arg := &Object{}, &SpellAcceptArg{}
			h := toxicCloudCastTestDeps52DB60(t, cloud, &cache)
			h.loadCache = func() uint32 { events = append(events, "load"); return cache }
			h.lookupType = func(string) uint32 { events = append(events, "lookup"); return 11 }
			h.storeCache = func(value uint32) { events = append(events, "store"); cache = value }
			h.traceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool {
				events = append(events, "trace")
				return fault == "cloud-update"
			}
			h.newObject = func(uint32) *Object { events = append(events, "new"); return cloud }
			h.createAt = func(*Object, *Object, types.Pointf) { events = append(events, "create") }
			h.lifetime = func(string) float32 { events = append(events, "lifetime"); return 1 }
			h.fps = func() uint32 { events = append(events, "fps"); return 30 }
			h.castSound = func(int32) sound.ID { t.Fatal("sound after required pointer fault"); return 0 }
			want := []string{"load", "lookup", "store"}
			switch fault {
			case "caster":
				caster = nil
			case "argument":
				arg = nil
			case "cloud-update":
				want = append(want, "trace", "load", "new", "create", "lifetime", "fps")
			case "player-update":
				caster.ObjClass = object.ClassPlayer
				want = append(want, "trace")
			case "player":
				caster.ObjClass = object.ClassPlayer
				caster.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
				want = append(want, "trace")
			}
			panicked := false
			func() { defer func() { panicked = recover() != nil }(); toxicCloudCast52DB60(128, nil, caster, arg, h) }()
			if !panicked || cache != 11 || !reflect.DeepEqual(events, want) {
				t.Fatalf("fault=%t cache=%d events=%v want=%v", panicked, cache, events, want)
			}
		})
	}
}

// Independent precision-53/ToZero multiplication followed by precision-24
// binary32 rounding and integer truncation; do not reuse production helpers.
func toxicCloudLifetimeReference52DB60(lifetime float32, fps uint32) int32 {
	if math.IsNaN(float64(lifetime)) || math.IsInf(float64(lifetime), 0) {
		return math.MinInt32
	}
	a := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetFloat64(float64(lifetime))
	b := new(big.Float).SetPrec(53).SetMode(big.ToZero).SetInt64(int64(int32(fps)))
	product := new(big.Float).SetPrec(53).SetMode(big.ToZero).Mul(a, b)
	// Conversion is always out of int32 range before binary32 overflow matters.
	spill := new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(product)
	value, _ := spill.Int(nil)
	if !value.IsInt64() || value.Int64() < math.MinInt32 || value.Int64() > math.MaxInt32 {
		return math.MinInt32
	}
	return int32(value.Int64())
}

func TestToxicCloudLifetimeOriginalFrameArithmetic52DB60(t *testing.T) {
	words := []uint32{0, 0x80000000, 1, 0x80000001, 0x007fffff, 0x00800000, 0x3f000000, 0x3f800000,
		0x3fbfffff, 0x3fc00000, 0x3fc00001, 0x40200000, 0x40600000, 0xbfc00000, 0xc0200000, 0xc0600000,
		0x4affffff, 0x4b000001, 0x4b7fffff, 0x4b800001, 0x4effffff, 0x4f000000, 0xcf000000, 0xcf000001,
		0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000, 0x7fc12345}
	for _, bits := range words {
		for _, fps := range []uint32{0, 1, 2, 3, 30, 31, 60, 0x7fffffff, 0x80000000, 0x80000001, 0xfffffffe, 0xffffffff} {
			t.Run(fmt.Sprintf("value-%08x/fps-%08x", bits, fps), func(t *testing.T) {
				value := math.Float32frombits(bits)
				got, want := toxicCloudLifetime52DB60(value, fps), toxicCloudLifetimeReference52DB60(value, fps)
				if got != want {
					t.Fatalf("lifetime bits=%08x fps=%08x got=%d want=%d", bits, fps, got, want)
				}
			})
		}
	}
	for _, tc := range []struct {
		value float32
		fps   uint32
		want  int32
	}{{1.5, 1, 1}, {2.5, 1, 2}, {3.5, 1, 3}, {-1.5, 1, -1}, {1.5, 31, 46}, {1, 0x80000000, math.MinInt32}} {
		if got := toxicCloudLifetime52DB60(tc.value, tc.fps); got != tc.want {
			t.Fatalf("literal %g * %08x = %d, want %d", tc.value, tc.fps, got, tc.want)
		}
	}
}
