package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestPullCastNativePointersLiveOriginAndOrder52BFA0(t *testing.T) {
	owner, freeOwner := alloc.New(Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(Object{})
	defer freeCaster()
	*owner, *caster = Object{PosVec: types.Ptf(7, 8)}, Object{PosVec: types.Ptf(9, 10)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(caster)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Pull pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	var events []string
	got := pullCast52BFA0(int32(spell.SPELL_PULL), owner, caster, 3, pullCastDeps52BFA0{
		balance: func(key string) float64 {
			events = append(events, "balance")
			if key != "PullPowerCoeff" {
				t.Fatalf("balance key=%q", key)
			}
			caster.PosVec = types.Ptf(123.5, -456.25)
			return 1.00000004
		},
		pushUnits: func(origin types.Pointf, outer, inner, force float32, source *Object, callback, callbackArg int) {
			events = append(events, "pull")
			if origin != types.Ptf(123.5, -456.25) || outer != 600 || inner != 10 || math.Float32bits(force) != 0xc0400001 ||
				source != nil || callback != 0 || callbackArg != 0 {
				t.Fatalf("pull=%v/%g/%g/%#x/%p/%d/%d", origin, outer, inner, math.Float32bits(force), source, callback, callbackArg)
			}
			owner.PosVec = types.Ptf(11, 12)
		},
		castAudio: func(id int32, unit *Object) {
			events = append(events, "audio")
			if id != int32(spell.SPELL_PULL) || unit != owner || unit.PosVec != types.Ptf(11, 12) {
				t.Fatalf("audio=%d/%p", id, unit)
			}
		},
	})
	if got != 1 || !reflect.DeepEqual(events, []string{"balance", "pull", "audio"}) {
		t.Fatalf("result/events=%d/%v", got, events)
	}
}

func TestPullCastSignedLevelNegationAndBinary32Spill52BFA0(t *testing.T) {
	coefficients := []float64{1.00000004, -1.00000004, 0, math.Copysign(0, -1), math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1), math.NaN()}
	for _, coefficient := range coefficients {
		for _, level := range []int32{math.MinInt32, -1, 0, 1, 2, 3, 4, 5, math.MaxInt32} {
			t.Run(fmt.Sprintf("%g/%d", coefficient, level), func(t *testing.T) {
				var gotForce float32
				pulls, audio := 0, 0
				got := pullCast52BFA0(int32(spell.SPELL_PULL), nil, &Object{}, level, pullCastDeps52BFA0{
					balance: func(string) float64 { return coefficient },
					pushUnits: func(_ types.Pointf, _, _, force float32, _ *Object, _, _ int) {
						pulls++
						gotForce = force
					},
					castAudio: func(id int32, owner *Object) {
						audio++
						if id != int32(spell.SPELL_PULL) || owner != nil {
							t.Fatalf("nil-owner audio=%d/%p", id, owner)
						}
					},
				})
				want := float32(-(coefficient * float64(level)))
				if got != 1 || pulls != 1 || audio != 1 ||
					(!math.IsNaN(float64(want)) && math.Float32bits(gotForce) != math.Float32bits(want)) ||
					(math.IsNaN(float64(want)) && !math.IsNaN(float64(gotForce))) {
					t.Fatalf("result/pull/audio/force=%d/%d/%d/%#x want %#x", got, pulls, audio, math.Float32bits(gotForce), math.Float32bits(want))
				}
			})
		}
	}
}

func TestPullCastServerBalanceAndOwnerAudio52BFA0(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	for _, solo := range []bool{false, true} {
		t.Run(fmt.Sprintf("solo-%t", solo), func(t *testing.T) {
			noxflags.ResetGame()
			wantCoefficient := float64(2.50000001)
			if solo {
				noxflags.SetGame(noxflags.GameModeCoop)
				wantCoefficient = 4.50000001
			}
			s := &Server{}
			s.Balance.file = &balance.File{
				Global: balance.Config{"pullpowercoeff": balance.Float(2.50000001)},
				Tags:   map[balance.Tag]balance.Config{balance.TagSolo: {"pullpowercoeff": balance.Float(4.50000001)}},
			}
			s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_PULL: {CastSound: sound.SoundPullCast, OnSound: sound.SoundPushCast}}
			owner, caster, second := &Object{PosVec: types.Ptf(1, 2)}, &Object{PosVec: types.Ptf(100, 200)}, &Object{}
			arg := &SpellAcceptArg{Obj: second, Pos: types.Ptf(-1000, -2000)}
			before := *arg
			pulls := 0
			got := s.CastPull52BFA0(int32(spell.SPELL_PULL), second, owner, caster, arg, 3, PullCastRuntime52BFA0{
				PushUnits: func(origin types.Pointf, outer, inner, force float32, source *Object, callback, callbackArg int) {
					pulls++
					if len(s.Audio.delayedObj) != 0 || origin != caster.PosVec || outer != 600 || inner != 10 ||
						force != float32(-(wantCoefficient*3)) || source != nil || callback != 0 || callbackArg != 0 {
						t.Fatal("wrong server Pull binding or audio order")
					}
				},
			})
			if got != 1 || pulls != 1 || *arg != before || len(s.Audio.delayedObj) != 1 {
				t.Fatalf("result/pull/arg/audio=%d/%d/%+v/%v", got, pulls, *arg, s.Audio.delayedObj)
			}
			if event := s.Audio.delayedObj[0]; event.ID != sound.SoundPullCast || event.Obj != owner || event.Kind != 0 || event.Code != 0 {
				t.Fatalf("audio event=%+v", event)
			}
		})
	}
}

func TestPullCastUnusedArgumentsAndMissingCaster52BFA0(t *testing.T) {
	pulls := 0
	s := &Server{}
	if got := s.CastPull52BFA0(int32(spell.SPELL_PULL), nil, nil, &Object{}, nil, 0, PullCastRuntime52BFA0{
		PushUnits: func(_ types.Pointf, _, _, force float32, _ *Object, _, _ int) {
			pulls++
			if math.Float32bits(force) != 0x80000000 {
				t.Fatalf("zero force=%#x, want original negative zero", math.Float32bits(force))
			}
		},
	}); got != 1 || pulls != 1 {
		t.Fatalf("unused nil arguments result/pull=%d/%d", got, pulls)
	}
	balanceCalls := 0
	defer func() {
		if recover() == nil || balanceCalls != 1 {
			t.Fatal("missing caster was silently ignored or balance order changed")
		}
	}()
	pullCast52BFA0(int32(spell.SPELL_PULL), nil, nil, 1, pullCastDeps52BFA0{
		balance: func(string) float64 { balanceCalls++; return 5 },
		pushUnits: func(types.Pointf, float32, float32, float32, *Object, int, int) {
			t.Fatal("missing caster reached radial service")
		},
		castAudio: func(int32, *Object) { t.Fatal("missing caster reached audio") },
	})
}

func TestPullCastNativePositionLayout52BFA0(t *testing.T) {
	want := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = 60
	}
	if got := unsafe.Offsetof(Object{}.PosVec); got != want {
		t.Fatalf("native position offset=%d, want %d", got, want)
	}
}
