package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// GAME.EXE 00545497 pushes TIME before reading FPS or consuming RNG; 005454D8
// pushes FLEE before reloading the enemy from the entry-cached update record.
func TestMonsterActionRetreat545440OriginalFallbackOrder(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	unit.UpdateDataMonster().CurrentEnemy = &Object{PosVec: types.Ptf(300, 400)}
	var events []string
	var time, flee *AIStackItem
	frame := uint32(1000)
	monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
		castRelated: func(*Object) bool { events = append(events, "related"); return false },
		tickRate: func() uint32 {
			events = append(events, "tick")
			return 30
		},
		random: func(low, high int) int {
			events = append(events, fmt.Sprintf("random %d..%d", low, high))
			frame = 1200
			return 151
		},
		frame: func() uint32 { events = append(events, "frame"); return frame },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			events = append(events, fmt.Sprintf("push %d args=%d", action, len(args)))
			item := &AIStackItem{Action: uint32(action), Args: [4]uintptr{71, 72, 73, 74}}
			item.SetArgs(args...)
			if action == ai.DEPENDENCY_TIME {
				time = item
			} else {
				flee = item
			}
			return item
		},
		pop: func() int { t.Fatal("injured retreat popped"); return 0 },
	})
	want := []string{"related", "push 41 args=0", "tick", "random 120..180", "frame", "push 24 args=0"}
	if !reflect.DeepEqual(events, want) || time == nil || flee == nil ||
		time.Args != [4]uintptr{1351, 72, 73, 74} ||
		flee.Args != [4]uintptr{uintptr(math.Float32bits(300)), uintptr(math.Float32bits(400)), 0, 74} {
		t.Fatalf("original fallback order/payload: events=%v TIME=%+v FLEE=%+v", events, time, flee)
	}
}

func TestMonsterActionRetreat545440RejectedPushDoesNotConsumeServices(t *testing.T) {
	for _, acceptTime := range []bool{false, true} {
		for _, acceptFlee := range []bool{false, true} {
			t.Run(fmt.Sprintf("TIME=%t/FLEE=%t", acceptTime, acceptFlee), func(t *testing.T) {
				unit := retreatMonsterTestObject545440(t)
				unit.UpdateDataMonster().CurrentEnemy = &Object{PosVec: types.Ptf(30, 40)}
				var actions []ai.ActionType
				var tickCalls, randomCalls, frameCalls, payloads int
				time := &AIStackItem{Action: uint32(ai.DEPENDENCY_TIME)}
				flee := &AIStackItem{Action: uint32(ai.ACTION_FLEE)}
				monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
					tickRate: func() uint32 { tickCalls++; return 30 },
					random:   func(int, int) int { randomCalls++; return 151 },
					frame:    func() uint32 { frameCalls++; return 1000 },
					push: func(action ai.ActionType, args ...any) *AIStackItem {
						actions = append(actions, action)
						payloads += len(args)
						var item *AIStackItem
						if action == ai.DEPENDENCY_TIME && acceptTime {
							item = time
						} else if action == ai.ACTION_FLEE && acceptFlee {
							item = flee
						}
						item.SetArgs(args...)
						return item
					},
					pop: func() int { t.Fatal("injured retreat popped"); return 0 },
				})
				wantCalls := 0
				if acceptTime {
					wantCalls = 1
				}
				if !reflect.DeepEqual(actions, []ai.ActionType{ai.DEPENDENCY_TIME, ai.ACTION_FLEE}) ||
					tickCalls != wantCalls || randomCalls != wantCalls || frameCalls != wantCalls || payloads != 0 ||
					acceptTime && time.ArgU32(0) != 1151 ||
					acceptFlee && (flee.ArgPos(0) != unit.UpdateDataMonster().CurrentEnemy.PosVec || flee.ArgU32(2) != 0) {
					t.Fatalf("rejected push services: actions=%v tick/RNG/frame=%d/%d/%d payloads=%d TIME=%+v FLEE=%+v", actions, tickCalls, randomCalls, frameCalls, payloads, time, flee)
				}
			})
		}
	}
}

func TestMonsterActionRetreat545440SingleFPSReadAndDWORDWrap(t *testing.T) {
	for _, fps := range []uint32{0, 30, 0x2aaaaaab, 0x40000000, 0x40000001, 0x80000000, math.MaxUint32} {
		t.Run(fmt.Sprintf("fps_%08x", fps), func(t *testing.T) {
			unit := retreatMonsterTestObject545440(t)
			unit.UpdateDataMonster().CurrentEnemy = &Object{}
			var tickCalls, frameCalls int
			var low, high int
			var time *AIStackItem
			frame := uint32(1)
			monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
				tickRate: func() uint32 { tickCalls++; return fps },
				random:   func(a, b int) int { low, high = a, b; frame = math.MaxUint32 - 2; return -7 },
				frame:    func() uint32 { frameCalls++; return frame },
				push: func(action ai.ActionType, args ...any) *AIStackItem {
					item := &AIStackItem{Action: uint32(action)}
					item.SetArgs(args...)
					if action == ai.DEPENDENCY_TIME {
						time = item
					}
					return item
				},
				pop: func() int { t.Fatal("injured retreat popped"); return 0 },
			})
			wantLow, wantHigh := int(int32(4*fps)), int(int32(6*fps))
			if tickCalls != 1 || frameCalls != 1 || low != wantLow || high != wantHigh ||
				time == nil || time.Args[0] != uintptr(uint32(math.MaxUint32-9)) {
				t.Fatalf("PE32 FPS/deadline: fps=%x calls=%d/%d bounds=%d..%d want=%d..%d TIME=%+v", fps, tickCalls, frameCalls, low, high, wantLow, wantHigh, time)
			}
		})
	}
}

func TestMonsterActionRetreat545440CachedEnemyReloadAfterCallbacks(t *testing.T) {
	for _, stage := range []string{"related", "TIME", "FLEE"} {
		for _, replaceUpdate := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/replace-update=%t", stage, replaceUpdate), func(t *testing.T) {
				unit := retreatMonsterTestObject545440(t)
				entry := unit.UpdateDataMonster()
				first := &Object{PosVec: types.Ptf(30, 40)}
				second := &Object{PosVec: types.Ptf(math.Float32frombits(0x80000000), math.Float32frombits(0x7fc54321))}
				entry.CurrentEnemy = first
				replacement := &MonsterUpdateData{CurrentEnemy: &Object{PosVec: types.Ptf(50, 60)}}
				mutate := func(at string) {
					if at != stage {
						return
					}
					entry.CurrentEnemy = second
					if replaceUpdate {
						unit.UpdateData = unsafe.Pointer(replacement)
					}
					first.PosVec = types.Ptf(70, 80)
				}
				var flee *AIStackItem
				monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
					castRelated: func(*Object) bool { mutate("related"); return false },
					tickRate:    func() uint32 { return 30 },
					random:      func(int, int) int { return 151 },
					frame:       func() uint32 { return 1000 },
					push: func(action ai.ActionType, args ...any) *AIStackItem {
						if action == ai.DEPENDENCY_TIME {
							mutate("TIME")
						} else {
							mutate("FLEE")
						}
						item := &AIStackItem{Action: uint32(action)}
						item.SetArgs(args...)
						if action == ai.ACTION_FLEE {
							flee = item
						}
						return item
					},
					pop: func() int { t.Fatal("injured retreat popped"); return 0 },
				})
				if flee == nil || flee.Args != [4]uintptr{0x80000000, 0x7fc54321, 0, 0} || entry.CurrentEnemy != second ||
					replaceUpdate && unit.UpdateDataMonster() != replacement {
					t.Fatalf("FLEE lost entry-cached update/live raw enemy position: item=%+v enemy=%p/%p", flee, entry.CurrentEnemy, second)
				}
				runtime.KeepAlive(entry)
				runtime.KeepAlive(replacement)
				runtime.KeepAlive(first)
				runtime.KeepAlive(second)
			})
		}
	}
}

func TestMonsterActionRetreat545440RejectedFleeSkipsEnemyRead(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{PosVec: types.Ptf(30, 40)}
	var payloads int
	var time *AIStackItem
	monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
		tickRate: func() uint32 { return 30 },
		random:   func(int, int) int { return 151 },
		frame:    func() uint32 { return 1000 },
		push: func(action ai.ActionType, args ...any) *AIStackItem {
			payloads += len(args)
			if action == ai.ACTION_FLEE {
				update.CurrentEnemy = nil
				return nil
			}
			time = &AIStackItem{Action: uint32(action)}
			time.SetArgs(args...)
			return time
		},
		pop: func() int { t.Fatal("injured retreat popped"); return 0 },
	})
	if payloads != 0 || time == nil || time.ArgU32(0) != 1151 || update.CurrentEnemy != nil {
		t.Fatalf("rejected FLEE read/payload: payloads=%d TIME=%+v enemy=%p", payloads, time, update.CurrentEnemy)
	}
}

func TestMonsterActionRetreat545440NilEnemyFaultPreservesCompletedPrefix(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{PosVec: types.Ptf(30, 40)}
	var actions []ai.ActionType
	var time, flee *AIStackItem
	var fault any
	func() {
		defer func() { fault = recover() }()
		monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
			tickRate: func() uint32 { return 30 },
			random:   func(int, int) int { return 151 },
			frame:    func() uint32 { return 1000 },
			push: func(action ai.ActionType, args ...any) *AIStackItem {
				actions = append(actions, action)
				item := &AIStackItem{Action: uint32(action), Args: [4]uintptr{71, 72, 73, 74}}
				item.SetArgs(args...)
				if action == ai.DEPENDENCY_TIME {
					time = item
				} else {
					flee, update.CurrentEnemy = item, nil
				}
				return item
			},
			pop: func() int { t.Fatal("injured retreat popped"); return 0 },
		})
	}()
	if fault == nil || !reflect.DeepEqual(actions, []ai.ActionType{ai.DEPENDENCY_TIME, ai.ACTION_FLEE}) ||
		time == nil || time.Args != [4]uintptr{1151, 72, 73, 74} ||
		flee == nil || flee.Args != [4]uintptr{71, 72, 73, 74} {
		t.Fatalf("nil live enemy fault lost original completed prefix: fault=%v actions=%v TIME=%+v FLEE=%+v", fault, actions, time, flee)
	}
}

func TestMonsterActionRetreat545440SelectedRelatedSpellSkipsFallback(t *testing.T) {
	unit := retreatMonsterTestObject545440(t)
	unit.UpdateDataMonster().CurrentEnemy = &Object{}
	var calls int
	monsterActionRetreat545440(unit, monsterActionRetreatHooks545440{
		castRelated: func(got *Object) bool {
			if got != unit {
				t.Fatal("related selector lost native unit")
			}
			calls++
			unit.UpdateDataMonster().CurrentEnemy = nil
			return true
		},
		tickRate: func() uint32 { t.Fatal("selected spell read fallback FPS"); return 0 },
		random:   func(int, int) int { t.Fatal("selected spell consumed fallback RNG"); return 0 },
		frame:    func() uint32 { t.Fatal("selected spell read fallback frame"); return 0 },
		push:     func(ai.ActionType, ...any) *AIStackItem { t.Fatal("selected spell pushed fallback"); return nil },
		pop:      func() int { t.Fatal("selected spell popped retreat"); return 0 },
	})
	if calls != 1 {
		t.Fatalf("related calls = %d", calls)
	}
}

type monsterRetreatCancelAction545440 struct{ cancel func(*Object) }

func (monsterRetreatCancelAction545440) Type() ai.ActionType   { return ai.ACTION_RETREAT }
func (monsterRetreatCancelAction545440) Start(*Object)         {}
func (monsterRetreatCancelAction545440) Update(*Object)        {}
func (monsterRetreatCancelAction545440) End(*Object)           {}
func (a monsterRetreatCancelAction545440) Cancel(unit *Object) { a.cancel(unit) }

// Only the ordinary Cancel callback mutates the enemy. Both pushes and RNG
// are real server services; this is not a replacement cast or damage result.
func TestMonsterRetreatServer545440FallbackUsesRealCancelAndStackGates(t *testing.T) {
	for _, slots := range []int{0, 1, 2, 5} {
		t.Run(fmt.Sprintf("slots_%d", slots), func(t *testing.T) {
			s, unit, update := monsterRetreatSpellFixture545440(t, false)
			update.StatusFlags = 0
			update.AIStackInd = int8(len(update.AIStack) - 1 - slots)
			update.AIStackHead().Action = uint32(ai.ACTION_RETREAT)
			update.AIStackHead().Field5 = 1
			second := &Object{PosVec: types.Ptf(-70, 80)}
			var cancelCalls int
			old, existed := aiActions[ai.ACTION_RETREAT]
			aiActions[ai.ACTION_RETREAT] = monsterRetreatCancelAction545440{cancel: func(got *Object) {
				if got != unit {
					t.Fatal("real Cancel lost native unit")
				}
				cancelCalls++
				update.CurrentEnemy = second
			}}
			t.Cleanup(func() {
				if existed {
					aiActions[ai.ACTION_RETREAT] = old
				} else {
					delete(aiActions, ai.ACTION_RETREAT)
				}
			})
			control := prand.New(545440)
			expectedDeadline := uint32(100)
			if slots > 0 {
				// Mirror only the original successful TIME push's real RNG call.
				expectedDeadline += uint32(control.IntClamp(120, 180))
			}
			s.MonsterActionRetreat545440(unit)
			if slots == 0 {
				if cancelCalls != 0 || s.AI.StackChanged || update.AIStackHead().Type() != ai.ACTION_RETREAT {
					t.Fatal("full stack invoked Cancel or changed RETREAT")
				}
			} else {
				if cancelCalls != 1 || !s.AI.StackChanged || update.CurrentEnemy != second {
					t.Fatalf("real Cancel prefix: calls=%d changed=%t enemy=%p", cancelCalls, s.AI.StackChanged, update.CurrentEnemy)
				}
				index := int(update.AIStackInd)
				if slots == 1 {
					if update.AIStackHead().Type() != ai.DEPENDENCY_TIME {
						t.Fatal("one slot lost accepted TIME")
					}
				} else {
					if update.AIStackHead().Type() != ai.ACTION_FLEE || update.AIStackHead().ArgPos(0) != second.PosVec || update.AIStackHead().ArgU32(2) != 0 {
						t.Fatalf("real Cancel left stale FLEE target: head=%+v", update.AIStackHead())
					}
					index--
				}
				if update.AIStack[index].ArgU32(0) != expectedDeadline {
					t.Fatalf("real TIME deadline=%d, want %d", update.AIStack[index].ArgU32(0), expectedDeadline)
				}
			}
			if got, want := s.Rand.Logic.IntClamp(0, math.MaxInt32), control.IntClamp(0, math.MaxInt32); got != want {
				t.Fatalf("real fallback consumed RNG outside successful TIME push: %d, want %d", got, want)
			}
			runtime.KeepAlive(unit)
			runtime.KeepAlive(second)
		})
	}
}
