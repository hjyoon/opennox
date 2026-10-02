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

func TestPushCastNativePointersLiveOriginAndOrder52C000(t *testing.T) {
	owner, freeOwner := alloc.New(Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(Object{})
	defer freeCaster()
	*owner, *caster = Object{PosVec: types.Ptf(7, 8)}, Object{PosVec: types.Ptf(9, 10)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(caster)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Push pointer=%p, want >4 GiB", pointer)
			}
		}
	}
	var events []string
	got := pushCast52C000(int32(spell.SPELL_PUSH), owner, caster, 3, pushCastDeps52C000{
		balance: func(key string) float64 {
			events = append(events, "balance")
			if key != "PushPowerCoeff" {
				t.Fatalf("balance key=%q", key)
			}
			caster.PosVec = types.Ptf(123.5, -456.25)
			return 1.00000004
		},
		pushUnits: func(origin types.Pointf, outer, inner, force float32, source *Object, callback, callbackArg int) {
			events = append(events, "push")
			if origin != types.Ptf(123.5, -456.25) || outer != 600 || inner != 10 || math.Float32bits(force) != 0x40400001 ||
				source != nil || callback != 0 || callbackArg != 0 {
				t.Fatalf("push=%v/%g/%g/%#x/%p/%d/%d", origin, outer, inner, math.Float32bits(force), source, callback, callbackArg)
			}
			owner.PosVec = types.Ptf(11, 12)
		},
		castAudio: func(id int32, unit *Object) {
			events = append(events, "audio")
			if id != int32(spell.SPELL_PUSH) || unit != owner || unit.PosVec != types.Ptf(11, 12) {
				t.Fatalf("audio=%d/%p", id, unit)
			}
		},
	})
	if got != 1 || !reflect.DeepEqual(events, []string{"balance", "push", "audio"}) {
		t.Fatalf("result/events=%d/%v", got, events)
	}
}

func TestPushCastSignedLevelAndBinary32Spill52C000(t *testing.T) {
	for _, coefficient := range []float64{1.00000004, -1.00000004, 0, math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1), math.NaN()} {
		for _, level := range []int32{math.MinInt32, -1, 0, 1, 2, 3, 4, 5, math.MaxInt32} {
			t.Run(fmt.Sprintf("%g/%d", coefficient, level), func(t *testing.T) {
				var gotForce float32
				pushes, audio := 0, 0
				got := pushCast52C000(66, nil, &Object{}, level, pushCastDeps52C000{
					balance: func(string) float64 { return coefficient },
					pushUnits: func(_ types.Pointf, _, _, force float32, _ *Object, _, _ int) {
						pushes++
						gotForce = force
					},
					castAudio: func(id int32, owner *Object) {
						audio++
						if id != 66 || owner != nil {
							t.Fatalf("nil-owner audio=%d/%p", id, owner)
						}
					},
				})
				want := float32(coefficient * float64(level))
				if got != 1 || pushes != 1 || audio != 1 ||
					(!math.IsNaN(float64(want)) && math.Float32bits(gotForce) != math.Float32bits(want)) ||
					(math.IsNaN(float64(want)) && !math.IsNaN(float64(gotForce))) {
					t.Fatalf("result/push/audio/force=%d/%d/%d/%#x want %#x", got, pushes, audio, math.Float32bits(gotForce), math.Float32bits(want))
				}
			})
		}
	}
}

func TestPushCastServerBalanceAndOwnerAudio52C000(t *testing.T) {
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
				Global: balance.Config{"pushpowercoeff": balance.Float(2.50000001)},
				Tags:   map[balance.Tag]balance.Config{balance.TagSolo: {"pushpowercoeff": balance.Float(4.50000001)}},
			}
			s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_PUSH: {CastSound: sound.SoundPushCast, OnSound: sound.SoundPullCast}}
			owner, caster, second := &Object{PosVec: types.Ptf(1, 2)}, &Object{PosVec: types.Ptf(100, 200)}, &Object{}
			arg := &SpellAcceptArg{Obj: second, Pos: types.Ptf(-1000, -2000)}
			before := *arg
			pushes := 0
			got := s.CastPush52C000(int32(spell.SPELL_PUSH), second, owner, caster, arg, 3, PushCastRuntime52C000{
				PushUnits: func(origin types.Pointf, outer, inner, force float32, source *Object, callback, callbackArg int) {
					pushes++
					if len(s.Audio.delayedObj) != 0 || origin != caster.PosVec || outer != 600 || inner != 10 ||
						force != float32(wantCoefficient*3) || source != nil || callback != 0 || callbackArg != 0 {
						t.Fatal("wrong server push binding or audio order")
					}
				},
			})
			if got != 1 || pushes != 1 || *arg != before || len(s.Audio.delayedObj) != 1 {
				t.Fatalf("result/push/arg/audio=%d/%d/%+v/%v", got, pushes, *arg, s.Audio.delayedObj)
			}
			if event := s.Audio.delayedObj[0]; event.ID != sound.SoundPushCast || event.Obj != owner || event.Kind != 0 || event.Code != 0 {
				t.Fatalf("audio event=%+v", event)
			}
		})
	}
}

func TestPushCastUnusedArgumentsAndMissingCaster52C000(t *testing.T) {
	pushes := 0
	s := &Server{}
	if got := s.CastPush52C000(66, nil, nil, &Object{}, nil, 0, PushCastRuntime52C000{
		PushUnits: func(types.Pointf, float32, float32, float32, *Object, int, int) { pushes++ },
	}); got != 1 || pushes != 1 {
		t.Fatalf("unused nil arguments result/push=%d/%d", got, pushes)
	}
	balanceCalls := 0
	defer func() {
		if recover() == nil || balanceCalls != 1 {
			t.Fatal("missing caster was silently ignored or balance order changed")
		}
	}()
	pushCast52C000(66, nil, nil, 1, pushCastDeps52C000{
		balance: func(string) float64 { balanceCalls++; return 5 },
		pushUnits: func(types.Pointf, float32, float32, float32, *Object, int, int) {
			t.Fatal("missing caster reached push")
		},
		castAudio: func(int32, *Object) { t.Fatal("missing caster reached audio") },
	})
}

func TestPushCastNativePositionLayout52C000(t *testing.T) {
	want := uintptr(56)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		want = 60
	}
	if got := unsafe.Offsetof(Object{}.PosVec); got != want {
		t.Fatalf("native position offset=%d, want %d", got, want)
	}
}
