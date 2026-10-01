package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFistCastNativePointersAndLiveOrder52D3C0(t *testing.T) {
	owner, freeOwner := alloc.New(Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(Object{})
	defer freeCaster()
	ownA, freeA := alloc.New(Object{})
	defer freeA()
	ownB, freeB := alloc.New(Object{})
	defer freeB()
	fist, freeFist := alloc.New(Object{})
	defer freeFist()
	data, freeData := alloc.New(FistUpdateData{})
	defer freeData()
	replacement, freeReplacement := alloc.New(FistUpdateData{})
	defer freeReplacement()
	arg, freeArg := alloc.New(SpellAcceptArg{})
	defer freeArg()
	*owner = Object{Field129: ownA}
	*ownA = Object{TypeInd: 101, Field128: ownB}
	*ownB = Object{TypeInd: 102}
	*caster = Object{PosVec: types.Ptf(10, 20)}
	*data = FistUpdateData{Damage: -123}
	*replacement = FistUpdateData{Damage: -456}
	*fist = Object{UpdateData: unsafe.Pointer(data), Field5: 0x1234, ObjFlags: 0x100, Field27: 99, Field29: 0x12345678}
	*arg = SpellAcceptArg{Obj: ownB, Pos: types.Ptf(30, 40)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"owner": unsafe.Pointer(owner), "caster": unsafe.Pointer(caster), "owned-first": unsafe.Pointer(ownA),
			"owned-next": unsafe.Pointer(ownB), "fist": unsafe.Pointer(fist), "damage": unsafe.Pointer(data), "arg": unsafe.Pointer(arg),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s = %p, want native pointer above 4 GiB", name, pointer)
			}
		}
	}
	var cache [6]uint32
	var events []string
	lookups := 0
	got := fistCast52D3C0(29, owner, caster, arg, 3, fistCastDeps52D3C0{
		loadCache: func(slot int32) uint32 {
			events = append(events, fmt.Sprintf("load:%d", slot))
			return cache[slot]
		},
		storeCache: func(slot int32, value uint32) {
			events = append(events, fmt.Sprintf("store:%d:%d", slot, value))
			cache[slot] = value
		},
		lookupType: func(name string) uint32 {
			events = append(events, "type:"+name)
			lookups++
			return uint32(lookups + 10)
		},
		hasGameFlag: func(flag uint32) bool {
			events = append(events, "mode")
			if flag != 2048 {
				t.Fatalf("mode flag = %#x", flag)
			}
			caster.PosVec = types.Ptf(50, 60)
			arg.Pos = types.Ptf(70, 80)
			return true
		},
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			events = append(events, "trace")
			if from != types.Ptf(10, 20) || to != types.Ptf(30, 40) || flags != 9 {
				t.Fatalf("trace = %v/%v/%d", from, to, flags)
			}
			return true
		},
		newObject: func(kind uint32) *Object {
			events = append(events, "new")
			if kind != 13 {
				t.Fatalf("type = %d, want cache slot 3", kind)
			}
			return fist
		},
		balance: func(key string, index int32) float64 {
			events = append(events, "damage")
			if key != "FistOfVengeanceDamage" || index != 2 {
				t.Fatalf("balance = %s/%d", key, index)
			}
			// The data pointer is cached before balance; creation position is not.
			fist.UpdateData = unsafe.Pointer(replacement)
			arg.Pos = types.Ptf(90, 100)
			return 2.50000001
		},
		createAt: func(got, by *Object, position types.Pointf) {
			events = append(events, "create")
			if got != fist || by != caster || position != types.Ptf(90, 100) || data.Damage != 2 || replacement.Damage != -456 {
				t.Fatalf("create = %p/%p/%v damage=%d/%d", got, by, position, data.Damage, replacement.Damage)
			}
			fist.PosVec = position
			fist.Field5 = 0x80000001
		},
		raise: func(got *Object, z float32) {
			events = append(events, "raise")
			if got != fist || z != 255 || fist.Field5 != 0x80000021 || fist.Field27 != 99 {
				t.Fatalf("raise = %p/%g fields=%#x/%g", got, z, fist.Field5, fist.Field27)
			}
			fist.ZVal = z
		},
		speed: func(key string) float64 {
			events = append(events, "speed")
			if key != "FistSpeed" || fist.ZVal != 255 {
				t.Fatalf("speed = %s z=%g", key, fist.ZVal)
			}
			fist.ObjFlags = 0x40000001
			return 0.123456789
		},
		castAudio: func(id int32, by *Object) {
			events = append(events, "audio")
			if id != 29 || by != caster || fist.ObjFlags != 0x40800001 || fist.Field29 != 0x41100000 ||
				math.Float32bits(fist.Field27) != math.Float32bits(float32(-0.123456789)) {
				t.Fatalf("audio state = %d/%p flags=%#x field29=%#x velocity=%g", id, by, fist.ObjFlags, fist.Field29, fist.Field27)
			}
		},
	})
	want := []string{"load:0", "type:SmallFist", "store:1:11", "type:MediumFist", "store:2:12",
		"type:LargeFist", "store:3:13", "type:LargeFist", "store:4:14", "type:LargeFist", "store:5:15", "store:0:1",
		"load:1", "load:2", "load:3", "load:4", "load:5", "load:1", "load:2", "load:3", "load:4", "load:5",
		"mode", "trace", "load:3", "new", "damage", "create", "raise", "speed", "audio"}
	if got != 1 || !reflect.DeepEqual(events, want) || arg.Obj != ownB || owner.Field129 != ownA || ownA.Field128 != ownB {
		t.Fatalf("result=%d events=%v, want %v", got, events, want)
	}
}

func TestFistCastLevelsCacheAndAllocationFailure52D3C0(t *testing.T) {
	for level := int32(1); level <= 5; level++ {
		for _, allocated := range []bool{true, false} {
			t.Run(fmt.Sprintf("level-%d/allocated-%t", level, allocated), func(t *testing.T) {
				owner, caster := &Object{}, &Object{}
				arg := &SpellAcceptArg{Pos: types.Ptf(100, 200)}
				data := &FistUpdateData{}
				fist := &Object{UpdateData: unsafe.Pointer(data), ObjFlags: 0x20, Field5: 0x80}
				cache := [6]uint32{7, 11, 12, 13, 13, 13}
				created, audio := false, false
				got := fistCast52D3C0(29, owner, caster, arg, level, fistCastDeps52D3C0{
					loadCache:   func(slot int32) uint32 { return cache[slot] },
					hasGameFlag: func(uint32) bool { return false },
					traceRay: func(_, to types.Pointf, flags MapTraceFlags) bool {
						if flags != 73 || to != arg.Pos {
							t.Fatalf("trace flags=%d pos=%v", flags, to)
						}
						return true
					},
					newObject: func(kind uint32) *Object {
						if kind != cache[level] {
							t.Fatalf("kind=%d want %d", kind, cache[level])
						}
						if !allocated {
							return nil
						}
						return fist
					},
					balance: func(_ string, index int32) float64 {
						if !allocated || index != level-1 {
							t.Fatalf("damage index=%d", index)
						}
						return float64(level) * 10.5
					},
					createAt: func(got, by *Object, pos types.Pointf) {
						created = true
						if got != fist || by != caster || pos != arg.Pos {
							t.Fatal("wrong creation")
						}
					},
					raise:     func(got *Object, z float32) { got.ZVal = z },
					speed:     func(string) float64 { return 4.5 },
					castAudio: func(int32, *Object) { audio = true },
				})
				if got != 1 || created != allocated || audio != allocated {
					t.Fatalf("result/create/audio=%d/%t/%t", got, created, audio)
				}
				if allocated && (data.Damage != int32(math.RoundToEven(float64(level)*10.5)) || fist.ZVal != 255 ||
					fist.Field27 != -4.5 || fist.Field29 != 0x41100000 || fist.Field5 != 0xa0 || fist.ObjFlags != 0x800020) {
					t.Fatalf("damage/height/velocity/fields=%d/%g/%g/%#x/%#x/%#x", data.Damage, fist.ZVal, fist.Field27, fist.Field29, fist.Field5, fist.ObjFlags)
				}
			})
		}
	}
}

func TestFistCastOwnedListDuplicate52D3C0(t *testing.T) {
	for slot := int32(1); slot <= 5; slot++ {
		t.Run(fmt.Sprintf("slot-%d", slot), func(t *testing.T) {
			cache := [6]uint32{1, 11, 12, 13, 14, 15}
			match := &Object{TypeInd: uint16(cache[slot])}
			first := &Object{TypeInd: 99, Field128: match}
			owner := &Object{Field129: first}
			var loads []int32
			messages := 0
			got := fistCast52D3C0(29, owner, nil, nil, 3, fistCastDeps52D3C0{
				loadCache: func(index int32) uint32 { loads = append(loads, index); return cache[index] },
				priorityMsg: func(unit *Object, text string, arg byte) {
					messages++
					if unit != owner || text != "ExecSpel.c:TooManyFists" || arg != 0 {
						t.Fatal("wrong duplicate message")
					}
				},
			})
			want := []int32{0, 1, 2, 3, 4, 5}
			for i := int32(1); i <= slot; i++ {
				want = append(want, i)
			}
			if got != 0 || messages != 1 || !reflect.DeepEqual(loads, want) {
				t.Fatalf("result/messages/loads=%d/%d/%v want %v", got, messages, loads, want)
			}
		})
	}
}

func TestFistCastCacheDWORDAndLiveOwnedNext52D3C0(t *testing.T) {
	first := &Object{TypeInd: 11}
	matched := &Object{TypeInd: 22}
	owner := &Object{Field129: first}
	cache := [6]uint32{1, 0x1000b, 22, 33, 33, 33}
	loads, messages := 0, 0
	got := fistCast52D3C0(29, owner, nil, nil, 3, fistCastDeps52D3C0{
		loadCache: func(slot int32) uint32 {
			loads++
			if loads == 2 {
				first.TypeInd = 22
				first.Field128 = matched
			}
			return cache[slot]
		},
		priorityMsg: func(unit *Object, _ string, _ byte) {
			if unit != owner {
				t.Fatal("wrong owner")
			}
			messages++
		},
	})
	// TypeInd is cached once per node, full cache DWORDs do not truncate to a
	// WORD, and the owned-next pointer is loaded after the five comparisons.
	if got != 0 || loads != 8 || messages != 1 {
		t.Fatalf("result/loads/messages=%d/%d/%d", got, loads, messages)
	}
}

func TestFistCastBlockedTracePlayerBinding52D3C0(t *testing.T) {
	for _, playerClass := range []bool{false, true} {
		for _, coop := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-%t/coop-%t", playerClass, coop), func(t *testing.T) {
				unit := &Object{PosVec: types.Ptf(10, 20)}
				if playerClass {
					unit.ObjClass = object.ClassPlayer
				}
				player, replacement := &Player{PlayerInd: 31}, &Player{PlayerInd: 255}
				update := &PlayerUpdateData{Player: player}
				unit.UpdateData = unsafe.Pointer(update)
				arg := &SpellAcceptArg{Pos: types.Ptf(30, 40)}
				before := *arg
				messages := 0
				got := fistCast52D3C0(29, unit, unit, arg, 3, fistCastDeps52D3C0{
					loadCache:   func(int32) uint32 { return 1 },
					hasGameFlag: func(uint32) bool { return coop },
					traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
						wantFlags := MapTraceFlags(73)
						if coop {
							wantFlags = 9
						}
						if from != unit.PosVec || to != arg.Pos || flags != wantFlags {
							t.Fatalf("trace = %v/%v/%d", from, to, flags)
						}
						update.Player = replacement
						return false
					},
					inform: func(index uint8, code byte, value int32) {
						messages++
						if index != 255 || code != 0 || value != 2 {
							t.Fatalf("inform = %d/%d/%d", index, code, value)
						}
					},
				})
				wantMessages := 0
				if playerClass {
					wantMessages = 1
				}
				if got != 0 || messages != wantMessages || *arg != before {
					t.Fatalf("result/messages/arg=%d/%d/%+v", got, messages, *arg)
				}
			})
		}
	}
}

func TestFistCastDamageConversion52D3C0(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  int32
	}{
		{2.5, 2}, {3.5, 4}, {-2.5, -2}, {2.50000001, 2},
		{math.NaN(), math.MinInt32}, {math.Inf(1), math.MinInt32}, {math.Inf(-1), math.MinInt32},
		{2147483647, math.MinInt32}, {-2147483648, math.MinInt32}, {2147483520, 2147483520},
	} {
		t.Run(fmt.Sprint(tc.value), func(t *testing.T) {
			data := &FistUpdateData{}
			fist := &Object{UpdateData: unsafe.Pointer(data)}
			got := fistCast52D3C0(29, &Object{}, &Object{}, &SpellAcceptArg{}, 3, fistCastDeps52D3C0{
				loadCache: func(int32) uint32 { return 1 }, hasGameFlag: func(uint32) bool { return false },
				traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true }, newObject: func(uint32) *Object { return fist },
				balance: func(string, int32) float64 { return tc.value }, createAt: func(*Object, *Object, types.Pointf) {},
				raise: func(*Object, float32) {}, speed: func(string) float64 { return 0 }, castAudio: func(int32, *Object) {},
			})
			if got != 1 || data.Damage != tc.want || math.Float32bits(fist.Field27) != 0x80000000 {
				t.Fatalf("result/damage/velocity=%d/%d/%#x want %d", got, data.Damage, math.Float32bits(fist.Field27), tc.want)
			}
		})
	}
}

func TestFistCastNativeLayout52D3C0(t *testing.T) {
	want := []uintptr{4, 16, 20, 56, 104, 108, 116, 512, 516, 748, 276, 2064, 4, 12}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = []uintptr{8, 20, 24, 60, 108, 112, 120, 560, 568, 872, 336, 2068, 8, 16}
	}
	got := []uintptr{unsafe.Offsetof(Object{}.TypeInd), unsafe.Offsetof(Object{}.ObjFlags), unsafe.Offsetof(Object{}.Field5),
		unsafe.Offsetof(Object{}.PosVec), unsafe.Offsetof(Object{}.ZVal), unsafe.Offsetof(Object{}.Field27), unsafe.Offsetof(Object{}.Field29),
		unsafe.Offsetof(Object{}.Field128), unsafe.Offsetof(Object{}.Field129), unsafe.Offsetof(Object{}.UpdateData),
		unsafe.Offsetof(PlayerUpdateData{}.Player), unsafe.Offsetof(Player{}.PlayerInd), unsafe.Offsetof(SpellAcceptArg{}.Pos), unsafe.Sizeof(SpellAcceptArg{})}
	if !reflect.DeepEqual(got, want) || unsafe.Sizeof(FistUpdateData{}.Damage) != 4 {
		t.Fatalf("layout=%v want %v", got, want)
	}
	for slot := int32(0); slot <= 5; slot++ {
		if got := fistCastCacheOffset52D3C0(slot); got != uintptr(2487736+4*slot) {
			t.Fatalf("slot=%d offset=%d", slot, got)
		}
	}
}

func TestFistCastMissingBindingFaults52D3C0(t *testing.T) {
	for _, kind := range []string{"owner", "caster", "arg", "damage", "player-update", "player"} {
		t.Run(kind, func(t *testing.T) {
			owner, caster, arg := &Object{}, &Object{}, &SpellAcceptArg{}
			data := &FistUpdateData{}
			fist := &Object{UpdateData: unsafe.Pointer(data)}
			blocked := false
			switch kind {
			case "owner":
				owner = nil
			case "caster":
				caster = nil
			case "arg":
				arg = nil
			case "damage":
				fist.UpdateData = nil
			case "player-update":
				caster.ObjClass = object.ClassPlayer
				blocked = true
			case "player":
				caster.ObjClass = object.ClassPlayer
				caster.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
				blocked = true
			}
			defer func() {
				if recover() == nil {
					t.Fatalf("missing %s binding was silently ignored", kind)
				}
			}()
			fistCast52D3C0(29, owner, caster, arg, 3, fistCastDeps52D3C0{
				loadCache: func(int32) uint32 { return 1 }, hasGameFlag: func(uint32) bool { return false },
				traceRay:  func(types.Pointf, types.Pointf, MapTraceFlags) bool { return !blocked },
				newObject: func(uint32) *Object { return fist }, balance: func(string, int32) float64 { return 123 },
			})
		})
	}
}
