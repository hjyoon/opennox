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

func meteorCastAlloc52D9D0[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("native allocation %p must be above 4 GiB", ptr)
	}
	return ptr
}

func meteorCastTestDeps52D9D0(t *testing.T, meteor *Object, cache *uint32) meteorCastDeps52D9D0 {
	t.Helper()
	return meteorCastDeps52D9D0{
		loadCache: func() uint32 { return *cache }, storeCache: func(kind uint32) { *cache = kind },
		lookupType: func(name string) uint32 {
			if name != "Meteor" {
				t.Fatalf("lookup=%q", name)
			}
			return 7
		},
		priorityMsg: func(*Object, string, byte) { t.Fatal("unexpected duplicate message") },
		hasGameFlag: func(uint32) bool { return false },
		traceRay:    func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		newObject:   func(uint32) *Object { return meteor },
		balance:     func(string, int32) float64 { return 2.5 },
		createAt:    func(*Object, *Object, types.Pointf) {},
		raise:       func(got *Object, height float32) { got.ZVal = height },
		speed:       func(string) float64 { return 0 },
		castSound:   func(int32) sound.ID { return 123 },
		audio:       func(sound.ID, types.Pointf, int, uint32) {},
		inform:      func(uint8, byte, int32) { t.Fatal("unexpected blocked-player message") },
	}
}

func TestMeteorCastNativePointersAndLiveOrder52D9D0(t *testing.T) {
	last := meteorCastAlloc52D9D0(t, Object{TypeInd: 102})
	first := meteorCastAlloc52D9D0(t, Object{TypeInd: 101, Field128: last})
	owner := meteorCastAlloc52D9D0(t, Object{Field129: first, PosVec: types.Ptf(-1, -2)})
	caster := meteorCastAlloc52D9D0(t, Object{PosVec: types.Ptf(10, 20)})
	arg := meteorCastAlloc52D9D0(t, SpellAcceptArg{Obj: last, Pos: types.Ptf(30, 40)})
	data := meteorCastAlloc52D9D0(t, MeteorUpdateData{Damage: -123})
	replacement := meteorCastAlloc52D9D0(t, MeteorUpdateData{Damage: -456})
	meteor := meteorCastAlloc52D9D0(t, Object{UpdateData: unsafe.Pointer(data), Field5: 0x1234,
		ObjFlags: 0x100, Field27: 99, Field29: 0x12345678})
	var cache uint32
	var events []string
	h := meteorCastTestDeps52D9D0(t, meteor, &cache)
	h.loadCache = func() uint32 { events = append(events, "load"); return cache }
	h.storeCache = func(kind uint32) { events = append(events, "store"); cache = kind }
	h.lookupType = func(name string) uint32 {
		events = append(events, "type:"+name)
		return 11
	}
	h.hasGameFlag = func(flag uint32) bool {
		events = append(events, "mode")
		if flag != 2048 {
			t.Fatalf("mode flag=%#x", flag)
		}
		cache = 77
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
		if kind != 77 {
			t.Fatalf("live allocation cache=%d", kind)
		}
		return meteor
	}
	h.balance = func(key string, index int32) float64 {
		events = append(events, "damage")
		if key != "MeteorDamage" || index != 2 {
			t.Fatalf("balance=%q/%d", key, index)
		}
		meteor.UpdateData = unsafe.Pointer(replacement)
		arg.Pos = types.Ptf(90, 100)
		return 2.50000001 // Spills to binary32 2.5 before ties-to-even conversion.
	}
	h.createAt = func(got, by *Object, pos types.Pointf) {
		events = append(events, "create")
		if got != meteor || by != owner || by == caster || pos != types.Ptf(90, 100) ||
			data.Damage != 2 || replacement.Damage != -456 {
			t.Fatalf("create=%p/%p/%v damage=%d/%d", got, by, pos, data.Damage, replacement.Damage)
		}
		meteor.Field5 = 0x80000001
	}
	h.raise = func(got *Object, height float32) {
		events = append(events, "raise")
		if got != meteor || height != 255 || meteor.Field5 != 0x80000021 || meteor.Field27 != 99 {
			t.Fatalf("raise=%p/%g flags=%#x velocity=%g", got, height, meteor.Field5, meteor.Field27)
		}
		meteor.ZVal = height
	}
	h.speed = func(key string) float64 {
		events = append(events, "speed")
		if key != "MeteorSpeed" || meteor.ZVal != 255 {
			t.Fatalf("speed=%q height=%g", key, meteor.ZVal)
		}
		meteor.ObjFlags = 0x40000001
		arg.Pos = types.Ptf(110, 120)
		return 0.123456789
	}
	h.castSound = func(id int32) sound.ID {
		events = append(events, "cast-sound")
		if id != 52 || meteor.ObjFlags != 0x40000001 || meteor.Field29 != 0x12345678 ||
			math.Float32bits(meteor.Field27) != math.Float32bits(float32(-0.123456789)) {
			t.Fatalf("audio prefix=%d/%#x/%#x/%g", id, meteor.ObjFlags, meteor.Field29, meteor.Field27)
		}
		arg.Pos = types.Ptf(130, 140)
		return 321
	}
	h.audio = func(id sound.ID, pos types.Pointf, kind int, code uint32) {
		events = append(events, "audio")
		if id != 321 || pos != types.Ptf(130, 140) || kind != 0 || code != 0 {
			t.Fatalf("positional audio=%d/%v/%d/%d", id, pos, kind, code)
		}
	}
	got := meteorCast52D9D0(52, owner, caster, arg, 3, h)
	want := []string{"load", "type:Meteor", "store", "mode", "trace", "load", "new", "damage", "create", "raise", "speed", "cast-sound", "audio"}
	if got != 1 || !reflect.DeepEqual(events, want) || owner.Field129 != first || first.Field128 != last || arg.Obj != last {
		t.Fatalf("result/events/list/target=%d/%v/%p/%p/%p", got, events, owner.Field129, first.Field128, arg.Obj)
	}
}

func TestMeteorCastLevelsCacheAndAllocationFailure52D9D0(t *testing.T) {
	for _, level := range []int32{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		for _, allocated := range []bool{false, true} {
			t.Run(fmt.Sprintf("level-%d/allocated-%t", level, allocated), func(t *testing.T) {
				cache := uint32(76543) // Full DWORD reaches allocation, not a WORD.
				data := &MeteorUpdateData{}
				meteor := &Object{UpdateData: unsafe.Pointer(data), Field5: 0x80, ObjFlags: 0x100, Field29: 17}
				owner, caster := &Object{}, &Object{Field129: &Object{TypeInd: 765}}
				arg := &SpellAcceptArg{Obj: caster, Pos: types.Ptf(1, 2)}
				before := *arg
				h := meteorCastTestDeps52D9D0(t, meteor, &cache)
				h.lookupType = func(string) uint32 { t.Fatal("initialized cache was looked up again"); return 0 }
				h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
					if from != caster.PosVec || to != arg.Pos || flags != 73 {
						t.Fatalf("trace=%v/%v/%d", from, to, flags)
					}
					return true
				}
				created, audio := 0, 0
				h.newObject = func(kind uint32) *Object {
					if kind != cache {
						t.Fatalf("allocation kind=%d", kind)
					}
					if !allocated {
						return nil
					}
					return meteor
				}
				h.balance = func(key string, index int32) float64 {
					if key != "MeteorDamage" || index != int32(uint32(level)-1) || !allocated {
						t.Fatalf("damage=%q/%d level=%d", key, index, level)
					}
					return 3.5
				}
				h.createAt = func(got, by *Object, pos types.Pointf) {
					created++
					if got != meteor || by != owner || pos != arg.Pos {
						t.Fatal("wrong placement or ownership")
					}
				}
				h.speed = func(string) float64 { return 4.5 }
				h.audio = func(sound.ID, types.Pointf, int, uint32) { audio++ }
				got := meteorCast52D9D0(52, owner, caster, arg, level, h)
				wantCalls := 0
				if allocated {
					wantCalls = 1
				}
				if got != 1 || created != wantCalls || audio != wantCalls || *arg != before || cache != 76543 {
					t.Fatalf("result/create/audio/arg/cache=%d/%d/%d/%+v/%d", got, created, audio, *arg, cache)
				}
				if allocated && (data.Damage != 4 || meteor.ZVal != 255 || meteor.Field27 != -4.5 ||
					meteor.Field5 != 0xa0 || meteor.ObjFlags != 0x100 || meteor.Field29 != 17) {
					t.Fatalf("unexpected meteor state=%+v damage=%d", meteor, data.Damage)
				}
			})
		}
	}
}

func TestMeteorCastOwnedListAndCachedDWORD52D9D0(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind uint32
		pos  int
	}{
		{"first", 11, 0}, {"middle", 11, 1}, {"last", 11, 2},
		{"different-WORD", 12, -1}, {"wide-cache-does-not-match-low-WORD", 0x1000b, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			third := meteorCastAlloc52D9D0(t, Object{TypeInd: 11})
			second := meteorCastAlloc52D9D0(t, Object{TypeInd: 11, Field128: third})
			first := meteorCastAlloc52D9D0(t, Object{TypeInd: 11, Field128: second})
			nodes := []*Object{first, second, third}
			if tc.pos >= 0 {
				for i := 0; i < tc.pos; i++ {
					nodes[i].TypeInd = 99
				}
			}
			owner := meteorCastAlloc52D9D0(t, Object{Field129: first})
			cache, loads, messages := tc.kind, 0, 0
			h := meteorCastTestDeps52D9D0(t, nil, &cache)
			h.loadCache = func() uint32 { loads++; return cache }
			h.priorityMsg = func(by *Object, text string, value byte) {
				messages++
				if by != owner || text != "ExecSpel.c:TooManyMeteors" || value != 0 {
					t.Fatal("wrong duplicate message")
				}
			}
			// A duplicate must return before touching either nil argument.
			var caster *Object
			var arg *SpellAcceptArg
			want := int32(0)
			if tc.pos < 0 {
				caster, arg, want = &Object{}, &SpellAcceptArg{}, 1
			}
			got := meteorCast52D9D0(52, owner, caster, arg, 3, h)
			wantLoads, wantMessages := 1, 1
			if tc.pos < 0 {
				wantLoads, wantMessages = 2, 0
			}
			if got != want || loads != wantLoads || messages != wantMessages {
				t.Fatalf("result/loads/messages=%d/%d/%d", got, loads, messages)
			}
		})
	}
	for _, kind := range []uint32{0, 11} {
		t.Run(fmt.Sprintf("lookup-result-%d-is-kept-across-store", kind), func(t *testing.T) {
			owner := &Object{Field129: &Object{TypeInd: uint16(kind)}}
			cache := uint32(0)
			h := meteorCastTestDeps52D9D0(t, nil, &cache)
			h.lookupType = func(string) uint32 { return kind }
			h.storeCache = func(got uint32) {
				if got != kind {
					t.Fatalf("stored kind=%d", got)
				}
				cache = 99
			}
			messages := 0
			h.priorityMsg = func(*Object, string, byte) { messages++ }
			if got := meteorCast52D9D0(52, owner, nil, nil, 3, h); got != 0 || messages != 1 || cache != 99 {
				t.Fatalf("result/message/cache=%d/%d/%d", got, messages, cache)
			}
		})
	}
}

func TestMeteorCastBlockedTraceLivePlayer52D9D0(t *testing.T) {
	for _, beforePlayer := range []bool{false, true} {
		for _, afterPlayer := range []bool{false, true} {
			for _, coop := range []bool{false, true} {
				t.Run(fmt.Sprintf("class-%t-to-%t/coop-%t", beforePlayer, afterPlayer, coop), func(t *testing.T) {
					player := meteorCastAlloc52D9D0(t, Player{PlayerInd: 31})
					replacement := meteorCastAlloc52D9D0(t, Player{PlayerInd: 255})
					update := meteorCastAlloc52D9D0(t, PlayerUpdateData{Player: player})
					caster := meteorCastAlloc52D9D0(t, Object{PosVec: types.Ptf(10, 20), UpdateData: unsafe.Pointer(update)})
					if beforePlayer {
						caster.ObjClass = object.ClassPlayer
					}
					arg := meteorCastAlloc52D9D0(t, SpellAcceptArg{Pos: types.Ptf(30, 40)})
					before := *arg
					cache, messages := uint32(11), 0
					h := meteorCastTestDeps52D9D0(t, nil, &cache)
					h.hasGameFlag = func(uint32) bool { return coop }
					h.traceRay = func(from, to types.Pointf, flags MapTraceFlags) bool {
						wantFlags := MapTraceFlags(73)
						if coop {
							wantFlags = 9
						}
						if from != caster.PosVec || to != arg.Pos || flags != wantFlags {
							t.Fatalf("trace=%v/%v/%d", from, to, flags)
						}
						caster.ObjClass = 0
						if afterPlayer {
							caster.ObjClass = object.ClassPlayer
						}
						update.Player = replacement
						return false
					}
					h.inform = func(index uint8, code byte, value int32) {
						messages++
						if index != 255 || code != 0 || value != 2 {
							t.Fatalf("inform=%d/%d/%d", index, code, value)
						}
					}
					got := meteorCast52D9D0(52, &Object{}, caster, arg, 3, h)
					wantMessages := 0
					if afterPlayer {
						wantMessages = 1
					}
					if got != 0 || messages != wantMessages || *arg != before {
						t.Fatalf("result/messages/arg=%d/%d/%+v", got, messages, *arg)
					}
				})
			}
		}
	}
}

func TestMeteorCastDamageConversion52D9D0(t *testing.T) {
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
			meteor := &Object{UpdateData: unsafe.Pointer(data)}
			cache := uint32(11)
			h := meteorCastTestDeps52D9D0(t, meteor, &cache)
			h.balance = func(string, int32) float64 { return tc.value }
			if got := meteorCast52D9D0(52, &Object{}, &Object{}, &SpellAcceptArg{}, 3, h); got != 1 ||
				data.Damage != tc.want || math.Float32bits(meteor.Field27) != 0x80000000 {
				t.Fatalf("result/damage/velocity=%d/%d/%#x", got, data.Damage, math.Float32bits(meteor.Field27))
			}
		})
	}
}

func TestMeteorCastMissingBindingFaults52D9D0(t *testing.T) {
	for _, kind := range []string{"owner", "caster", "arg", "damage", "player-update", "player", "create"} {
		t.Run(kind, func(t *testing.T) {
			owner, caster, arg := &Object{}, &Object{}, &SpellAcceptArg{}
			meteor := &Object{UpdateData: unsafe.Pointer(&MeteorUpdateData{})}
			cache := uint32(11)
			h := meteorCastTestDeps52D9D0(t, meteor, &cache)
			switch kind {
			case "owner":
				owner = nil
			case "caster":
				caster = nil
			case "arg":
				arg = nil
			case "damage":
				meteor.UpdateData = nil
			case "player-update", "player":
				caster.ObjClass = object.ClassPlayer
				h.traceRay = func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false }
				if kind == "player" {
					caster.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
				}
			case "create":
				h.createAt = nil
			}
			defer func() {
				if recover() == nil {
					t.Fatalf("missing %s binding silently reported success", kind)
				}
			}()
			meteorCast52D9D0(52, owner, caster, arg, 3, h)
		})
	}
}

func TestMeteorCastNativeLayout52D9D0(t *testing.T) {
	want := []uintptr{4, 20, 56, 104, 108, 512, 516, 748, 276, 2064, 4, 12}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = []uintptr{8, 24, 60, 108, 112, 560, 568, 872, 336, 2068, 8, 16}
	}
	got := []uintptr{unsafe.Offsetof(Object{}.TypeInd), unsafe.Offsetof(Object{}.Field5), unsafe.Offsetof(Object{}.PosVec),
		unsafe.Offsetof(Object{}.ZVal), unsafe.Offsetof(Object{}.Field27), unsafe.Offsetof(Object{}.Field128),
		unsafe.Offsetof(Object{}.Field129), unsafe.Offsetof(Object{}.UpdateData), unsafe.Offsetof(PlayerUpdateData{}.Player),
		unsafe.Offsetof(Player{}.PlayerInd), unsafe.Offsetof(SpellAcceptArg{}.Pos), unsafe.Sizeof(SpellAcceptArg{})}
	if !reflect.DeepEqual(got, want) || unsafe.Sizeof(MeteorUpdateData{}) != 4 || unsafe.Offsetof(MeteorUpdateData{}.Damage) != 0 {
		t.Fatalf("native layout=%v want %v", got, want)
	}
}
