package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestPoisonCastSignedLevelsAndIgnoredActivationResult52C720(t *testing.T) {
	for _, level := range []int32{math.MinInt32, -3, -1, 0, 1, 2, 3, 4, 5, math.MaxInt32} {
		for _, result := range []int32{-17, 0, 1, 7} {
			t.Run(fmt.Sprintf("%d/%d", level, result), func(t *testing.T) {
				var events []string
				got := castPoison52C720("owner", level, poisonCastHooks52C720[string]{
					loadTarget: func() string { events = append(events, "target"); return "target" },
					activatePoison: func(target string, increment, maximum int32) int32 {
						events = append(events, "activate")
						if target != "target" || increment != level || maximum != level {
							t.Fatalf("activation=%s/%d/%d", target, increment, maximum)
						}
						return result
					},
					recordPlayerAttribution: func(owner, target string) {
						events = append(events, "attribution")
						if owner != "owner" || target != "target" {
							t.Fatalf("attribution=%s/%s", owner, target)
						}
					},
				})
				if got != 1 || !reflect.DeepEqual(events, []string{"target", "activate", "target", "attribution"}) {
					t.Fatalf("result/order=%d/%v", got, events)
				}
			})
		}
	}
}

func TestPoisonCastTargetReloadAndFaultPrefixes52C720(t *testing.T) {
	want := []string{"target", "activate", "target", "attribution"}
	for fault := 0; fault <= len(want); fault++ {
		t.Run(fmt.Sprintf("fault-%d", fault), func(t *testing.T) {
			var events []string
			live := "entry"
			event := func(name string) {
				events = append(events, name)
				if fault != 0 && len(events) == fault {
					panic("injected helper fault")
				}
			}
			result := int32(-1)
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				result = castPoison52C720("owner", 3, poisonCastHooks52C720[string]{
					loadTarget: func() string { event("target"); return live },
					activatePoison: func(target string, increment, maximum int32) int32 {
						if target != "entry" || increment != 3 || maximum != 3 {
							t.Fatal("entry target or signed level changed")
						}
						event("activate")
						live = "replacement"
						return 0
					},
					recordPlayerAttribution: func(owner, target string) {
						if owner != "owner" || target != "replacement" {
							t.Fatalf("live attribution=%s/%s", owner, target)
						}
						event("attribution")
					},
				})
			}()
			if fault == 0 {
				if recovered != nil || result != 1 || !reflect.DeepEqual(events, want) {
					t.Fatalf("result/fault/events=%d/%v/%v", result, recovered, events)
				}
			} else if recovered == nil || result != -1 || !reflect.DeepEqual(events, want[:fault]) {
				t.Fatalf("fault prefix=%v result=%d events=%v", recovered, result, events)
			}
		})
	}
}

func TestPoisonCastNativePointersAndLiveArgument52C720(t *testing.T) {
	owner, freeOwner := alloc.New(Object{})
	defer freeOwner()
	entry, freeEntry := alloc.New(Object{})
	defer freeEntry()
	replacement, freeReplacement := alloc.New(Object{})
	defer freeReplacement()
	arg, freeArg := alloc.New(SpellAcceptArg{})
	defer freeArg()
	*entry = Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDestroyed, Poison540: 7}
	*arg = SpellAcceptArg{Obj: entry, Pos: types.Ptf(-1000, 2000)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(entry), unsafe.Pointer(replacement), unsafe.Pointer(arg)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Poison pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	var events []string
	s := new(Server)
	got := s.CastPoison52C720(int32(spell.SPELL_POISON), nil, owner, nil, arg, -3, PoisonCastRuntime52C720{
		ActivatePoison: func(target *Object, increment, maximum int32) int32 {
			events = append(events, "activate")
			if target != entry || increment != -3 || maximum != -3 {
				t.Fatalf("native activation=%p/%d/%d", target, increment, maximum)
			}
			arg.Obj = replacement
			return 0
		},
		RecordPlayerAttribution: func(source, target *Object) {
			events = append(events, "attribution")
			if source != owner || target != replacement {
				t.Fatalf("native attribution=%p/%p", source, target)
			}
		},
	})
	if got != 1 || !reflect.DeepEqual(events, []string{"activate", "attribution"}) || arg.Obj != replacement || arg.Pos != types.Ptf(-1000, 2000) || entry.Poison540 != 7 {
		t.Fatalf("result/events/arg/poison=%d/%v/%+v/%d", got, events, *arg, entry.Poison540)
	}
}

func TestPoisonCastNilTargetsAndReloadedNil52C720(t *testing.T) {
	s := new(Server)
	if got := s.CastPoison52C720(-17, nil, nil, nil, new(SpellAcceptArg), math.MinInt32, PoisonCastRuntime52C720{}); got != 0 {
		t.Fatalf("nil target result=%d", got)
	}
	arg := &SpellAcceptArg{Obj: new(Object)}
	records := 0
	if got := s.CastPoison52C720(-17, nil, nil, nil, arg, -3, PoisonCastRuntime52C720{
		ActivatePoison: func(*Object, int32, int32) int32 { arg.Obj = nil; return -17 },
		RecordPlayerAttribution: func(owner, target *Object) {
			records++
			if owner != nil || target != nil {
				t.Fatalf("reloaded nil attribution=%p/%p", owner, target)
			}
		},
	}); got != 1 || records != 1 {
		t.Fatalf("reloaded nil result/records=%d/%d", got, records)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("missing spell argument did not fault at the original target load")
		}
	}()
	s.CastPoison52C720(-17, nil, nil, nil, nil, 0, PoisonCastRuntime52C720{})
}

func TestPoisonCastServerUsesExistingStateAndAudio52C720(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(0)
	s.SetFrame(91)
	health := new(HealthData)
	target := &Object{HealthData: health}
	arg := &SpellAcceptArg{Obj: target}
	before := *arg
	records := 0
	got := s.CastPoison52C720(int32(spell.SPELL_POISON), nil, nil, nil, arg, 3, PoisonCastRuntime52C720{
		ActivatePoison: func(target *Object, increment, maximum int32) int32 {
			return s.ActivatePoison4EE7E0(target, increment, maximum, ActivatePoisonRuntime4EE7E0{})
		},
		RecordPlayerAttribution: func(owner, got *Object) {
			records++
			if owner != nil || got != target || target.Poison540 != 3 || target.Field542 != 1000 || health.Field16 != 91 || len(s.Audio.delayedObj) != 1 {
				t.Fatal("attribution did not follow the real native poison services")
			}
		},
	})
	if got != 1 || records != 1 || *arg != before || s.Rand.Logic.Index() != 1 || len(s.Audio.delayedObj) != 1 {
		t.Fatalf("result/records/RNG/audio=%d/%d/%d/%v", got, records, s.Rand.Logic.Index(), s.Audio.delayedObj)
	}
	if event := s.Audio.delayedObj[0]; event.ID != 100 || event.Obj != target || event.Kind != 0 || event.Code != 0 {
		t.Fatalf("Poison audio=%+v", event)
	}
}

func TestPoisonCastNativeArgumentLayout52C720(t *testing.T) {
	wantSize, wantPos := uintptr(12), uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize, wantPos = 16, 8
	}
	if unsafe.Sizeof(SpellAcceptArg{}) != wantSize || unsafe.Offsetof(SpellAcceptArg{}.Obj) != 0 || unsafe.Offsetof(SpellAcceptArg{}.Pos) != wantPos {
		t.Fatalf("native spell argument size/target/position=%d/%d/%d", unsafe.Sizeof(SpellAcceptArg{}), unsafe.Offsetof(SpellAcceptArg{}.Obj), unsafe.Offsetof(SpellAcceptArg{}.Pos))
	}
}
