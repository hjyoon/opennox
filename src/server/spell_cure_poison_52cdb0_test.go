package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const curePoisonTestTarget52CDB0 = uint64(0x7fb555dd5b10)

type curePoisonCastTestWorld52CDB0 struct {
	events                      []string
	live                        uint64
	poison                      uint8
	effectTarget, messageTarget uint64
	audioTarget, refundTarget   uint64
	power                       int32
	message                     string
	cost                        int32
	refundAmount                int16
	after                       func(string)
}

func (w *curePoisonCastTestWorld52CDB0) record(stage string) {
	w.events = append(w.events, stage)
	if w.after != nil {
		w.after(stage)
	}
}

func (w *curePoisonCastTestWorld52CDB0) hooks(t *testing.T) curePoisonCastHooks52CDB0[uint64] {
	return curePoisonCastHooks52CDB0[uint64]{
		target: func() uint64 { w.record("target"); return w.live },
		poison: func(target uint64) uint8 {
			if target != curePoisonTestTarget52CDB0 {
				t.Fatalf("poison load target=%#x", target)
			}
			w.record("poison")
			return w.poison
		},
		update: func(target uint64, power int32) {
			w.effectTarget, w.power = target, power
			w.record("update")
			w.poison -= uint8(power)
		},
		remove: func(target uint64) {
			w.effectTarget = target
			w.record("remove")
			w.poison = 0
		},
		message: func(target uint64, message string, value uint8) {
			if value != 0 {
				t.Fatalf("message arg=%d", value)
			}
			w.messageTarget, w.message = target, message
			w.record("message")
		},
		onSound: func(id int32) sound.ID {
			if id != -17 {
				t.Fatalf("sound spell ID=%d", id)
			}
			w.record("sound")
			return sound.SoundCurePoisonEffect
		},
		audio: func(id sound.ID, target uint64, kind int, code uint32) {
			if id != sound.SoundCurePoisonEffect || kind != 0 || code != 0 {
				t.Fatalf("audio id/kind/code=%d/%d/%d", id, kind, code)
			}
			w.audioTarget = target
			w.record("audio")
		},
		manaCost: func(id, kind int32) int32 {
			if id != -17 || kind != 1 {
				t.Fatalf("mana cost id/kind=%d/%d", id, kind)
			}
			w.record("cost")
			return w.cost
		},
		refundMana: func(target uint64, amount int16) uint16 {
			w.refundTarget, w.refundAmount = target, amount
			w.record("refund")
			return 0xbeef // the original discards this service's return value
		},
	}
}

func TestCurePoisonCast52CDB0SignedPowerAndOriginalBranches(t *testing.T) {
	for _, poison := range []uint8{0, 1, 2, 3, 127, 128, 254, 255} {
		for _, power := range []int32{math.MinInt32, -3, -1, 0, 1, 2, 3, 127, 255, math.MaxInt32} {
			for _, self := range []bool{false, true} {
				t.Run(fmt.Sprintf("poison-%d/power-%d/self-%t", poison, power, self), func(t *testing.T) {
					w := &curePoisonCastTestWorld52CDB0{live: curePoisonTestTarget52CDB0, poison: poison, cost: 17}
					second := curePoisonTestTarget52CDB0 + 0x100000000
					if self {
						second = curePoisonTestTarget52CDB0
					}
					got := curePoisonCast52CDB0(-17, power, second, w.hooks(t))
					want := []string{"target", "poison"}
					wantPoison := poison
					if poison != 0 {
						operation, message := "remove", "ExecSpel.c:PoisonClean"
						wantPoison = 0
						if int32(poison) > power {
							operation, message = "update", "ExecSpel.c:PoisonCure"
							wantPoison = poison - uint8(power)
							if w.power != power {
								t.Fatalf("signed power=%d, want %d", w.power, power)
							}
						}
						want = append(want, operation, "target", "message")
						if w.effectTarget != curePoisonTestTarget52CDB0 || w.messageTarget != curePoisonTestTarget52CDB0 || w.message != message {
							t.Fatalf("treatment/message=%#x/%#x/%s", w.effectTarget, w.messageTarget, w.message)
						}
					}
					if poison == 0 && self {
						want = append(want, "cost", "target", "refund")
						if w.refundTarget != second || w.refundAmount != 17 || w.audioTarget != 0 {
							t.Fatalf("self refund/audio=%#x/%d/%#x", w.refundTarget, w.refundAmount, w.audioTarget)
						}
					} else {
						want = append(want, "target", "sound", "audio")
						if w.audioTarget != curePoisonTestTarget52CDB0 || w.refundTarget != 0 {
							t.Fatal("audio/refund branch changed")
						}
					}
					if got != 1 || w.poison != wantPoison || !reflect.DeepEqual(w.events, want) {
						t.Fatalf("result/poison/events=%d/%d/%v, want 1/%d/%v", got, w.poison, w.events, wantPoison, want)
					}
				})
			}
		}
	}
}

func TestCurePoisonCast52CDB0CallbackReloadsAndSoundSnapshot(t *testing.T) {
	for _, operation := range []string{"update", "remove", "unpoisoned-other"} {
		for _, clear := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/nil-%t", operation, clear), func(t *testing.T) {
				w := &curePoisonCastTestWorld52CDB0{live: curePoisonTestTarget52CDB0, poison: 3}
				power := int32(3)
				if operation == "update" {
					power = -1
				}
				if operation == "unpoisoned-other" {
					w.poison = 0
				}
				messageTarget, audioTarget := curePoisonTestTarget52CDB0+0x100000000, curePoisonTestTarget52CDB0+0x200000000
				if clear {
					messageTarget, audioTarget = 0, 0
				}
				w.after = func(stage string) {
					switch stage {
					case "poison":
						w.live = messageTarget
					case "message":
						w.live = audioTarget
					case "sound":
						w.live = curePoisonTestTarget52CDB0 + 0x300000000
					}
				}
				if got := curePoisonCast52CDB0(-17, power, curePoisonTestTarget52CDB0+1, w.hooks(t)); got != 1 {
					t.Fatalf("result=%d", got)
				}
				if operation == "unpoisoned-other" {
					audioTarget = messageTarget
				} else if w.effectTarget != curePoisonTestTarget52CDB0 || w.messageTarget != messageTarget {
					t.Fatalf("entry effect/live message=%#x/%#x", w.effectTarget, w.messageTarget)
				}
				if w.audioTarget != audioTarget {
					t.Fatalf("cached audio=%#x, want %#x", w.audioTarget, audioTarget)
				}
			})
		}
	}
}

func TestCurePoisonCast52CDB0RefundCostThenLiveTargetAndLowWord(t *testing.T) {
	for _, cost := range []int32{math.MinInt32, -32769, -1, 0, 32767, 32768, math.MaxInt32} {
		for _, clear := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/nil-%t", cost, clear), func(t *testing.T) {
				w := &curePoisonCastTestWorld52CDB0{live: curePoisonTestTarget52CDB0, cost: cost}
				wantTarget := curePoisonTestTarget52CDB0 + 0x100000000
				if clear {
					wantTarget = 0
				}
				w.after = func(stage string) {
					if stage == "cost" {
						w.live = wantTarget
					}
				}
				got := curePoisonCast52CDB0(-17, math.MinInt32, curePoisonTestTarget52CDB0, w.hooks(t))
				if got != 1 || w.refundTarget != wantTarget || w.refundAmount != int16(cost) ||
					!reflect.DeepEqual(w.events, []string{"target", "poison", "cost", "target", "refund"}) {
					t.Fatalf("refund result/state/order=%d/%+v", got, w)
				}
			})
		}
	}
}

func TestCurePoisonCast52CDB0OriginalFaultPrefixes(t *testing.T) {
	for _, branch := range []string{"update", "remove", "other", "self"} {
		want := []string{"target", "poison"}
		poison, power, second := uint8(0), int32(3), curePoisonTestTarget52CDB0+1
		switch branch {
		case "update", "remove":
			poison = 3
			if branch == "update" {
				power = 2
			}
			want = append(want, branch, "target", "message", "target", "sound", "audio")
		case "other":
			want = append(want, "target", "sound", "audio")
		case "self":
			second = curePoisonTestTarget52CDB0
			want = append(want, "cost", "target", "refund")
		}
		for fault := 0; fault <= len(want); fault++ {
			t.Run(fmt.Sprintf("%s/fault-%d", branch, fault), func(t *testing.T) {
				w := &curePoisonCastTestWorld52CDB0{live: curePoisonTestTarget52CDB0, poison: poison}
				w.after = func(string) {
					if fault != 0 && len(w.events) == fault {
						panic("helper fault")
					}
				}
				got := int32(-99)
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					got = curePoisonCast52CDB0(-17, power, second, w.hooks(t))
				}()
				if fault == 0 {
					if recovered != nil || got != 1 || !reflect.DeepEqual(w.events, want) {
						t.Fatalf("result/fault/order=%d/%v/%v", got, recovered, w.events)
					}
				} else if recovered == nil || got != -99 || !reflect.DeepEqual(w.events, want[:fault]) {
					t.Fatalf("fault prefix result/fault/order=%d/%v/%v", got, recovered, w.events)
				}
			})
		}
	}
}

func TestCurePoisonCast52CDB0NilTargetAndRequiredArgument(t *testing.T) {
	w := new(curePoisonCastTestWorld52CDB0)
	if got := curePoisonCast52CDB0(-17, math.MinInt32, curePoisonTestTarget52CDB0, w.hooks(t)); got != 0 || !reflect.DeepEqual(w.events, []string{"target"}) {
		t.Fatalf("nil result/order=%d/%v", got, w.events)
	}
	s := new(Server)
	if got := s.CastCurePoison52CDB0(-17, nil, nil, nil, new(SpellAcceptArg), 3, CurePoisonCastRuntime52CDB0{}); got != 0 {
		t.Fatalf("native nil target=%d", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("required acceptance pointer did not fault at the original first load")
		}
	}()
	s.CastCurePoison52CDB0(-17, nil, nil, nil, nil, 3, CurePoisonCastRuntime52CDB0{})
}

func TestCurePoisonCast52CDB0NativeStateMessageAndActualAudio(t *testing.T) {
	for _, power := range []int32{-3, 2, 3, math.MaxInt32} {
		t.Run(fmt.Sprint(power), func(t *testing.T) {
			target, freeTarget := alloc.New(Object{})
			defer freeTarget()
			replacement, freeReplacement := alloc.New(Object{})
			defer freeReplacement()
			health, freeHealth := alloc.New(HealthData{})
			defer freeHealth()
			arg, freeArg := alloc.New(SpellAcceptArg{})
			defer freeArg()
			*health = HealthData{Cur: 77, Max: 80, Field16: 91}
			*target = Object{HealthData: health, Poison540: 3, Field542: 1234, ObjFlags: object.FlagDestroyed}
			*arg = SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, 2000)}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(replacement), unsafe.Pointer(health), unsafe.Pointer(arg)} {
					if uintptr(ptr) <= math.MaxUint32 {
						t.Fatalf("actual native allocation=%p, want >4 GiB", ptr)
					}
				}
			}
			s := new(Server)
			s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_CURE_POISON: {OnSound: sound.SoundCurePoisonEffect}}
			messages := 0
			got := s.CastCurePoison52CDB0(int32(spell.SPELL_CURE_POISON), target, replacement, nil, arg, power, CurePoisonCastRuntime52CDB0{
				PriorityMessage: func(unit *Object, message string, value uint8) {
					messages++
					want := "ExecSpel.c:PoisonClean"
					if power < 3 {
						want = "ExecSpel.c:PoisonCure"
					}
					if unit != target || message != want || value != 0 || len(s.Audio.delayedObj) != 0 {
						t.Fatalf("message before sound=%p/%s/%d", unit, message, value)
					}
					arg.Obj = replacement
				},
			})
			wantPoison, wantFrame := uint8(0), uint32(0)
			if power < 3 {
				wantPoison, wantFrame = uint8(3)-uint8(power), 91
			}
			if got != 1 || messages != 1 || target.Poison540 != wantPoison || health.Field16 != wantFrame || health.Cur != 77 || health.Max != 80 || target.Field542 != 1234 || arg.Pos != types.Ptf(-1000, 2000) || len(s.Audio.delayedObj) != 1 {
				t.Fatalf("native state/result=%d target=%+v health=%+v arg=%+v audio=%+v", got, target, health, arg, s.Audio.delayedObj)
			}
			if event := s.Audio.delayedObj[0]; event.ID != sound.SoundCurePoisonEffect || event.Obj != replacement || event.Kind != 0 || event.Code != 0 {
				t.Fatalf("actual native audio=%+v", event)
			}
		})
	}
}

func TestCurePoisonCast52CDB0RealManaCostAndIgnoredRefundResult(t *testing.T) {
	s := new(Server)
	s.Spells.s = s
	def := &SpellDef{}
	def.Def.ManaCost = 32768
	s.Spells.byID = map[spell.ID]*SpellDef{spell.SPELL_CURE_POISON: def}
	target := new(Object)
	arg := &SpellAcceptArg{Obj: target, Pos: types.Ptf(-1, 2)}
	calls := 0
	if got := s.CastCurePoison52CDB0(int32(spell.SPELL_CURE_POISON), target, nil, nil, arg, 3, CurePoisonCastRuntime52CDB0{
		RefundMana: func(unit *Object, amount int16) uint16 {
			calls++
			if unit != target || amount != math.MinInt16 {
				t.Fatalf("native mana target/amount=%p/%d", unit, amount)
			}
			return 0
		},
	}); got != 1 || calls != 1 || len(s.Audio.delayedObj) != 0 || arg.Obj != target || arg.Pos != types.Ptf(-1, 2) {
		t.Fatalf("native refund result/calls/audio=%d/%d/%v", got, calls, s.Audio.delayedObj)
	}
}
