package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestInversionCast52BEB0NativePointersLiveOriginAndOrder(t *testing.T) {
	owner, freeOwner := alloc.New(Object{})
	t.Cleanup(freeOwner)
	caster, freeCaster := alloc.New(Object{})
	t.Cleanup(freeCaster)
	*owner, *caster = Object{PosVec: types.Ptf(7, 8)}, Object{PosVec: types.Ptf(9, 10)}
	first, second := new(Object), new(Object)
	var events []string
	got := inversionCast52BEB0(38, owner, caster, inversionCastDeps52BEB0{
		balance: func(key string) float64 {
			events = append(events, "balance:"+key)
			caster.PosVec = types.Ptf(123.5, 456.25)
			return 100.000004
		},
		eachMissile: func(origin types.Pointf, radius float32, visit func(*Object) bool) {
			events = append(events, "query")
			if origin != caster.PosVec || math.Float32bits(radius) != 0x42c80001 {
				t.Fatal("query did not use live native caster position and final binary32 range")
			}
			if !visit(first) || !visit(second) {
				t.Fatal("ownership callback stopped full missile enumeration")
			}
		},
		changeOwner: func(missile, gotOwner *Object) {
			if gotOwner != owner {
				t.Fatal("callback owner was replaced by caster")
			}
			if missile == first {
				events = append(events, "first")
			} else if missile == second {
				events = append(events, "second")
			} else {
				t.Fatal("wrong native missile")
			}
			caster.PosVec = types.Ptf(11, 12)
		},
		castAudio: func(id int32, unit *Object) {
			events = append(events, "audio")
			if id != 38 || unit != caster || unit.PosVec != types.Ptf(11, 12) {
				t.Fatal("cast audio did not retain live caster")
			}
		},
	})
	if got != 1 || !reflect.DeepEqual(events, []string{"balance:InversionRange", "query", "first", "second", "audio"}) {
		t.Fatalf("result/events=%d/%q", got, events)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(caster)) <= math.MaxUint32) {
		t.Fatal("fixture is not above 4 GiB")
	}
}

func TestInversionCast52BEB0Binary32RangeAndEmptyEnumeration(t *testing.T) {
	for _, radius := range []float64{100.000004, -100.000004, 0, math.Copysign(0, -1), math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1), math.NaN()} {
		t.Run(fmt.Sprintf("%g/%x", radius, math.Float64bits(radius)), func(t *testing.T) {
			queries, audio := 0, 0
			if inversionCast52BEB0(-1, nil, &Object{}, inversionCastDeps52BEB0{
				balance: func(string) float64 { return radius },
				eachMissile: func(_ types.Pointf, got float32, _ func(*Object) bool) {
					queries++
					if !math.IsNaN(radius) && math.Float32bits(got) != math.Float32bits(float32(radius)) || math.IsNaN(radius) && !math.IsNaN(float64(got)) {
						t.Fatal("range spill was clamped or rounded differently")
					}
				},
				castAudio: func(id int32, unit *Object) {
					audio++
					if id != -1 || unit == nil {
						t.Fatal("empty query omitted cast audio")
					}
				},
			}) != 1 || queries != 1 || audio != 1 {
				t.Fatal("empty query changed original result or audio")
			}
		})
	}
}

func TestInversionCast52BEB0ServerRangeMapBoundaryAndCasterAudio(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	for _, solo := range []bool{false, true} {
		t.Run(fmt.Sprint(solo), func(t *testing.T) {
			noxflags.ResetGame()
			if solo {
				noxflags.SetGame(noxflags.GameModeCoop)
			}
			s := &Server{}
			s.Map.Init()
			s.Balance.file = &balance.File{Global: balance.Config{"inversionrange": balance.Float(50)}, Tags: map[balance.Tag]balance.Config{balance.TagSolo: {"inversionrange": balance.Float(100)}}}
			s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_INVERSION: {CastSound: sound.SoundInversionCast, OnSound: sound.SoundPullCast}}
			owner, caster, second := new(Object), &Object{PosVec: types.Ptf(200, 200)}, new(Object)
			radius := float32(50)
			if solo {
				radius = 100
			}
			missiles := []*Object{
				{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, PosVec: types.Ptf(200+radius, 200)},
				{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, PosVec: types.Ptf(200+radius+1, 200)},
				{ObjClass: object.ClassMonster, ObjFlags: object.FlagActive, PosVec: caster.PosVec},
			}
			for _, missile := range missiles {
				missile.NewPos = missile.PosVec
				s.Map.AddObjectToIndex(missile)
			}
			arg := &SpellAcceptArg{Obj: second, Pos: types.Ptf(-1000, -2000)}
			before, changes := *arg, 0
			got := s.CastInversion52BEB0(38, second, owner, caster, arg, -3, InversionCastRuntime52BEB0{ChangeOwner: func(missile, gotOwner *Object) {
				changes++
				if missile != missiles[0] || gotOwner != owner || len(s.Audio.delayedObj) != 0 {
					t.Fatal("native map boundary/class gate, owner or audio order changed")
				}
			}})
			if got != 1 || changes != 1 || *arg != before || len(s.Audio.delayedObj) != 1 {
				t.Fatalf("result/changes/arg/audio=%d/%d/%+v/%v", got, changes, *arg, s.Audio.delayedObj)
			}
			if event := s.Audio.delayedObj[0]; event.ID != sound.SoundInversionCast || event.Obj != caster || event.Kind != 0 || event.Code != 0 {
				t.Fatalf("audio=%+v", event)
			}
		})
	}
}

func TestInversionCast52BEB0MissingCasterFaultAfterBalance(t *testing.T) {
	balanced := 0
	defer func() {
		if recover() == nil || balanced != 1 {
			t.Fatal("missing caster was swallowed or balance order changed")
		}
	}()
	inversionCast52BEB0(38, nil, nil, inversionCastDeps52BEB0{
		balance:     func(string) float64 { balanced++; return 50 },
		eachMissile: func(types.Pointf, float32, func(*Object) bool) { t.Fatal("missing caster reached query") },
		castAudio:   func(int32, *Object) { t.Fatal("missing caster reached audio") },
	})
}
