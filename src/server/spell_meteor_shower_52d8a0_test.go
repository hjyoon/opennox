package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func meteorShowerCastAlloc52D8A0[T comparable](t *testing.T, value T) *T {
	t.Helper()
	p, free := alloc.New(value)
	t.Cleanup(free)
	*p = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatalf("native MeteorShower pointer=%p, want >4 GiB", p)
	}
	return p
}

func meteorShowerCastTestDeps52D8A0(t *testing.T, shower *Object, cache *uint32) meteorShowerCastDeps52D8A0 {
	t.Helper()
	return meteorShowerCastDeps52D8A0{
		loadCache: func() uint32 { return *cache }, storeCache: func(kind uint32) { *cache = kind },
		lookupType: func(name string) uint32 {
			if name != "MeteorShower" {
				t.Fatalf("type lookup=%q", name)
			}
			return 11
		},
		hasGameFlag: func(uint32) bool { return false },
		traceRay:    func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		newObject:   func(uint32) *Object { return shower },
		balance:     func(string, int32) float64 { return 2.5 },
		createAt:    func(*Object, *Object, types.Pointf) {},
		castSound:   func(int32) sound.ID { return 123 },
		audio:       func(sound.ID, *Object, int, uint32) {},
		inform:      func(uint8, byte, int32) { t.Fatal("unexpected failure message") },
	}
}

func TestMeteorShowerCastNativePointersAndLiveOrder52D8A0(t *testing.T) {
	owned := meteorShowerCastAlloc52D8A0(t, Object{TypeInd: 11})
	owner := meteorShowerCastAlloc52D8A0(t, Object{Field129: owned, PosVec: types.Ptf(-1, -2)})
	caster := meteorShowerCastAlloc52D8A0(t, Object{PosVec: types.Ptf(10, 20)})
	arg := meteorShowerCastAlloc52D8A0(t, SpellAcceptArg{Obj: owned, Pos: types.Ptf(30, 40)})
	data := meteorShowerCastAlloc52D8A0(t, MeteorUpdateData{Damage: -123})
	replacement := meteorShowerCastAlloc52D8A0(t, MeteorUpdateData{Damage: -456})
	shower := meteorShowerCastAlloc52D8A0(t, Object{UpdateData: unsafe.Pointer(data), Field5: 0x1234,
		ObjFlags: 0x100, Field27: 99, ZVal: 17, Field29: 0x12345678})
	cache := uint32(0)
	var events []string
	h := meteorShowerCastTestDeps52D8A0(t, shower, &cache)
	h.loadCache = func() uint32 { events = append(events, "load"); return cache }
	h.storeCache = func(kind uint32) { events = append(events, "store"); cache = kind }
	h.lookupType = func(name string) uint32 { events = append(events, "type:"+name); return 11 }
	h.hasGameFlag = func(flag uint32) bool {
		events = append(events, "mode")
		if flag != 2048 {
			t.Fatalf("mode flag=%#x", flag)
		}
		cache = 0x1234abcd
		caster.PosVec, arg.Pos = types.Ptf(50, 60), types.Ptf(70, 80)
		return true
	}
	h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
		events = append(events, "trace")
		if from != types.Ptf(10, 20) || to != types.Ptf(30, 40) || flags != 9 {
			t.Fatalf("cached trace=%v/%v/%d", from, to, flags)
		}
		return true
	}
	h.newObject = func(kind uint32) *Object {
		events = append(events, "new")
		if kind != 0x1234abcd {
			t.Fatalf("live DWORD type cache=%#x", kind)
		}
		return shower
	}
	h.balance = func(key string, index int32) float64 {
		events = append(events, "damage")
		if key != "MeteorDamage" || index != 2 {
			t.Fatalf("damage table=%q/%d", key, index)
		}
		shower.UpdateData = unsafe.Pointer(replacement)
		arg.Pos = types.Ptf(90, 100)
		return 2.50000001
	}
	h.createAt = func(got, by *Object, position types.Pointf) {
		events = append(events, "create")
		if got != shower || by != owner || by == caster || position != types.Ptf(90, 100) ||
			data.Damage != 2 || replacement.Damage != -456 {
			t.Fatalf("create=%p/%p/%v damage=%d/%d", got, by, position, data.Damage, replacement.Damage)
		}
	}
	h.castSound = func(id int32) sound.ID {
		events = append(events, "cast-sound")
		if id != 53 || shower.Field5 != 0x1234 || shower.ObjFlags != 0x100 || shower.Field27 != 99 ||
			shower.ZVal != 17 || shower.Field29 != 0x12345678 {
			t.Fatalf("shower changed falling-Meteor fields or wrong spell=%d", id)
		}
		owner.PosVec, arg.Pos = types.Ptf(110, 120), types.Ptf(130, 140)
		return 321
	}
	h.audio = func(id sound.ID, by *Object, kind int, code uint32) {
		events = append(events, "audio")
		if id != 321 || by != owner || by.PosVec != types.Ptf(110, 120) || kind != 0 || code != 0 {
			t.Fatalf("owner-attached audio=%d/%p/%d/%d", id, by, kind, code)
		}
	}
	want := []string{"load", "type:MeteorShower", "store", "mode", "trace", "load", "new", "damage", "create", "cast-sound", "audio"}
	if got := meteorShowerCast52D8A0(53, owner, caster, arg, 3, h); got != 1 || !reflect.DeepEqual(events, want) ||
		owner.Field129 != owned || arg.Obj != owned {
		t.Fatalf("result/events/owned/arg=%d/%v/%p/%p", got, events, owner.Field129, arg.Obj)
	}
}

func TestMeteorShowerCastLevelsAndAllocationFailure52D8A0(t *testing.T) {
	for _, level := range []int32{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		for _, allocated := range []bool{false, true} {
			t.Run(fmt.Sprintf("level-%d/allocated-%t", level, allocated), func(t *testing.T) {
				cache := uint32(0xfedc0123)
				data := &MeteorUpdateData{Damage: 99}
				shower := &Object{UpdateData: unsafe.Pointer(data), Field5: 0x20, ZVal: 19, Field27: -3}
				owner, caster := &Object{Field129: &Object{TypeInd: 0x123}}, &Object{}
				arg := &SpellAcceptArg{Obj: caster, Pos: types.Ptf(1, 2)}
				before := *arg
				h := meteorShowerCastTestDeps52D8A0(t, shower, &cache)
				h.lookupType = func(string) uint32 { t.Fatal("initialized cache looked up again"); return 0 }
				h.newObject = func(kind uint32) *Object {
					if kind != cache {
						t.Fatalf("type cache=%#x", kind)
					}
					if !allocated {
						return nil
					}
					return shower
				}
				created, audio := 0, 0
				h.balance = func(key string, index int32) float64 {
					if key != "MeteorDamage" || index != int32(uint32(level)-1) || !allocated {
						t.Fatalf("balance=%q/%d level=%d", key, index, level)
					}
					return 3.5
				}
				h.createAt = func(got, by *Object, position types.Pointf) {
					created++
					if got != shower || by != owner || position != arg.Pos || data.Damage != 4 {
						t.Fatal("wrong shower placement, ownership or damage")
					}
				}
				h.audio = func(id sound.ID, by *Object, kind int, code uint32) {
					audio++
					if id != 123 || by != owner || kind != 0 || code != 0 {
						t.Fatal("wrong owner audio")
					}
				}
				wantCalls := 0
				if allocated {
					wantCalls = 1
				}
				if got := meteorShowerCast52D8A0(53, owner, caster, arg, level, h); got != 1 ||
					created != wantCalls || audio != wantCalls || *arg != before || cache != 0xfedc0123 ||
					shower.Field5 != 0x20 || shower.ZVal != 19 || shower.Field27 != -3 || (!allocated && data.Damage != 99) {
					t.Fatalf("result/create/audio=%d/%d/%d", got, created, audio)
				}
			})
		}
	}
}

func TestMeteorShowerCastBlockedTraceLivePlayer52D8A0(t *testing.T) {
	for _, beforePlayer := range []bool{false, true} {
		for _, afterPlayer := range []bool{false, true} {
			for _, coop := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-%t-to-%t/coop-%t", beforePlayer, afterPlayer, coop), func(t *testing.T) {
					player := meteorShowerCastAlloc52D8A0(t, Player{PlayerInd: 31})
					replacementPlayer := meteorShowerCastAlloc52D8A0(t, Player{PlayerInd: 255})
					update := meteorShowerCastAlloc52D8A0(t, PlayerUpdateData{Player: player})
					replacement := meteorShowerCastAlloc52D8A0(t, PlayerUpdateData{Player: replacementPlayer})
					caster := meteorShowerCastAlloc52D8A0(t, Object{UpdateData: unsafe.Pointer(update), PosVec: types.Ptf(10, 20)})
					if beforePlayer {
						caster.ObjClass = object.ClassPlayer
					}
					arg := meteorShowerCastAlloc52D8A0(t, SpellAcceptArg{Pos: types.Ptf(30, 40)})
					before := *arg
					cache, messages := uint32(11), 0
					h := meteorShowerCastTestDeps52D8A0(t, nil, &cache)
					h.hasGameFlag = func(uint32) bool { return coop }
					h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
						want := MapTraceFlags(73)
						if coop {
							want = 9
						}
						if from != caster.PosVec || to != arg.Pos || flags != want {
							t.Fatalf("trace=%v/%v/%d", from, to, flags)
						}
						caster.ObjClass = 0
						if afterPlayer {
							caster.ObjClass = object.ClassPlayer
						}
						caster.UpdateData = unsafe.Pointer(replacement)
						return false
					}
					h.newObject = func(uint32) *Object { t.Fatal("blocked cast allocated"); return nil }
					h.inform = func(index uint8, code byte, value int32) {
						messages++
						if index != 255 || code != 0 || value != 2 {
							t.Fatalf("live failure message=%d/%d/%d", index, code, value)
						}
					}
					wantMessages := 0
					if afterPlayer {
						wantMessages = 1
					}
					if got := meteorShowerCast52D8A0(53, nil, caster, arg, 3, h); got != 0 || messages != wantMessages || *arg != before {
						t.Fatalf("result/messages/arg=%d/%d/%+v", got, messages, *arg)
					}
				})
			}
		}
	}
}

func TestMeteorShowerCastDamageConversion52D8A0(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  int32
	}{
		{0, 0}, {2.5, 2}, {3.5, 4}, {-2.5, -2}, {-3.5, -4}, {2.50000001, 2},
		{math.NaN(), math.MinInt32}, {math.Inf(1), math.MinInt32}, {math.Inf(-1), math.MinInt32},
		{2147483647, math.MinInt32}, {-2147483648, math.MinInt32}, {2147483520, 2147483520}, {-2147483904, math.MinInt32},
	} {
		t.Run(fmt.Sprint(tc.value), func(t *testing.T) {
			data := &MeteorUpdateData{}
			shower := &Object{UpdateData: unsafe.Pointer(data)}
			cache := uint32(11)
			h := meteorShowerCastTestDeps52D8A0(t, shower, &cache)
			h.balance = func(string, int32) float64 { return tc.value }
			if got := meteorShowerCast52D8A0(53, nil, &Object{}, &SpellAcceptArg{}, 3, h); got != 1 || data.Damage != tc.want {
				t.Fatalf("result/damage=%d/%d", got, data.Damage)
			}
		})
	}
}

func TestMeteorShowerCastZeroLookupAndRepeatedCast52D8A0(t *testing.T) {
	cache, lookups, allocations := uint32(0), 0, 0
	shower := &Object{UpdateData: unsafe.Pointer(&MeteorUpdateData{})}
	owner := &Object{Field129: &Object{TypeInd: 11}}
	h := meteorShowerCastTestDeps52D8A0(t, shower, &cache)
	h.lookupType = func(name string) uint32 {
		lookups++
		if name != "MeteorShower" {
			t.Fatalf("lookup=%q", name)
		}
		return 0
	}
	h.newObject = func(kind uint32) *Object {
		allocations++
		if kind != 0 {
			t.Fatalf("zero lookup must reach allocation unchanged, got=%d", kind)
		}
		return shower
	}
	for i := 0; i < 2; i++ {
		if got := meteorShowerCast52D8A0(53, owner, &Object{}, &SpellAcceptArg{}, 3, h); got != 1 {
			t.Fatalf("repeated cast %d=%d", i, got)
		}
	}
	if cache != 0 || lookups != 2 || allocations != 2 {
		t.Fatalf("zero lookup/repeated casts=%d/%d/%d", cache, lookups, allocations)
	}
}

func TestMeteorShowerCastMissingBindingFaults52D8A0(t *testing.T) {
	for _, kind := range []string{"caster", "arg", "damage", "player-update", "player", "create"} {
		t.Run(kind, func(t *testing.T) {
			caster, arg := &Object{}, &SpellAcceptArg{}
			shower := &Object{UpdateData: unsafe.Pointer(&MeteorUpdateData{})}
			cache := uint32(11)
			h := meteorShowerCastTestDeps52D8A0(t, shower, &cache)
			switch kind {
			case "caster":
				caster = nil
			case "arg":
				arg = nil
			case "damage":
				shower.UpdateData = nil
			case "player-update", "player":
				caster.ObjClass = object.ClassPlayer
				if kind == "player" {
					caster.UpdateData = unsafe.Pointer(new(PlayerUpdateData))
				}
				h.traceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false }
			case "create":
				h.createAt = nil
			}
			defer func() {
				if recover() == nil {
					t.Fatal("missing binding must not silently report success")
				}
			}()
			meteorShowerCast52D8A0(53, nil, caster, arg, 3, h)
		})
	}
}
