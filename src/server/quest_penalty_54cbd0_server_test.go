package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestQuestPenalty54CBD0NativeCachedUpdateAndLateClass(t *testing.T) {
	for _, class := range []uint8{0, 1, 2, 128, 255} {
		p := &Player{GoldVal: math.MaxUint32, Level: 9}
		p.info[66] = 17
		late := &Player{}
		late.info[66] = class
		cached := &PlayerUpdateData{Player: p}
		other := &Player{GoldVal: 37}
		other.info[66] = class ^ 1
		replacement := &PlayerUpdateData{Player: other}
		u := &Object{UpdateData: unsafe.Pointer(cached), ObjClass: 0xffffffff, Worth: 0xaabbccdd}
		beforeUnit, beforeCached, beforeReplacement, beforeP, beforeOther, beforeLate := *u, *cached, *replacement, *p, *other, *late
		var trace []string
		var amount uint32
		armors := 0
		questPenaltyNative54CBD0(u, questPenaltyNativeDeps54CBD0{
			getGold: func(unit *Object) uint32 {
				if unit != u {
					t.Fatal("wrong native unit")
				}
				trace = append(trace, "gold")
				unit.UpdateData = unsafe.Pointer(replacement)
				return p.GoldVal
			},
			subGold:    func(unit *Object, v uint32) { trace = append(trace, "sub"); amount = v },
			loseGems:   func(*Object) { trace = append(trace, "gems") },
			loseWeapon: func(*Object) { trace = append(trace, "weapon"); cached.Player = nil },
			loseArmor: func(*Object) {
				trace = append(trace, "armor")
				armors++
				if armors == 1 {
					cached.Player = late
				} else {
					cached.Player = nil
				}
			},
			loseSpell:   func(*Object) { trace = append(trace, "spell") },
			loseGuide:   func(*Object) { trace = append(trace, "guide") },
			loseAbility: func(*Object) int8 { trace = append(trace, "ability"); return -128 },
		})
		wantTrace := []string{"gold", "sub", "gems", "weapon", "armor"}
		if class == 0 {
			wantTrace = append(wantTrace, "armor")
		}
		wantTrace = append(wantTrace, "spell", "spell", "guide", "guide", "ability")
		beforeUnit.UpdateData = unsafe.Pointer(replacement)
		beforeCached.Player = late
		if class == 0 {
			beforeCached.Player = nil
		}
		if amount != 0x7fffffff || !reflect.DeepEqual(trace, wantTrace) || *u != beforeUnit || *cached != beforeCached || *replacement != beforeReplacement || *p != beforeP || *other != beforeOther || *late != beforeLate {
			t.Fatalf("class=%d amount=%x trace=%v or unrelated native fields changed", class, amount, trace)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(u), unsafe.Pointer(cached), unsafe.Pointer(replacement), unsafe.Pointer(p), unsafe.Pointer(other), unsafe.Pointer(late)} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("native pointer %p must exceed 4 GiB", pointer)
				}
			}
		}
	}
}

func TestQuestPenalty54CBD0NativeMissingUpdateAndPlayerFaultAfterFirstArmor(t *testing.T) {
	for _, update := range []*PlayerUpdateData{nil, {}} {
		u := &Object{UpdateData: unsafe.Pointer(update)}
		var trace []string
		deps := questPenaltyNativeDeps54CBD0{
			getGold:    func(*Object) uint32 { trace = append(trace, "gold"); return 0 },
			subGold:    func(*Object, uint32) { trace = append(trace, "sub") },
			loseGems:   func(*Object) { trace = append(trace, "gems") },
			loseWeapon: func(*Object) { trace = append(trace, "weapon") },
			loseArmor:  func(*Object) { trace = append(trace, "armor") },
		}
		func() {
			defer func() {
				if recover() == nil || !reflect.DeepEqual(trace, []string{"gold", "sub", "gems", "weapon", "armor"}) {
					t.Fatalf("missing binding trace=%v", trace)
				}
			}()
			questPenaltyNative54CBD0(u, deps)
		}()
	}
	called := false
	func() {
		defer func() {
			if recover() == nil || called {
				t.Fatal("nil unit must fault before gold")
			}
		}()
		questPenaltyNative54CBD0(nil, questPenaltyNativeDeps54CBD0{getGold: func(*Object) uint32 { called = true; return 0 }})
	}()
}

func TestQuestPenalty54CBD0NativeRuntimeGoldReadAndRequiredZeroSubtraction(t *testing.T) {
	for _, gold := range []uint32{0, 1, 0x7fffffff, 0x80000000, math.MaxUint32} {
		p := &Player{GoldVal: gold}
		p.info[66] = 255 // All three later learned-item helpers reject this raw byte without RNG.
		u := &Object{ObjClass: 0, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
		var trace []string
		new(Server).QuestPenalty54CBD0(u, QuestPenaltyRuntime54CBD0{
			SubGold: func(unit *Object, amount uint32) {
				if unit != u || amount != gold>>1 {
					t.Fatalf("unit=%p amount=%08x", unit, amount)
				}
				trace = append(trace, "sub")
			},
			LoseGems:   func(*Object) { trace = append(trace, "gems") },
			LoseWeapon: func(*Object) { trace = append(trace, "weapon") },
			LoseArmor:  func(*Object) { trace = append(trace, "armor") },
		})
		if !reflect.DeepEqual(trace, []string{"sub", "gems", "weapon", "armor"}) || p.GoldVal != gold {
			t.Fatalf("gold=%x trace=%v", gold, trace)
		}
	}
	p := &Player{}
	p.info[66] = 255
	u := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
	defer func() {
		if recover() == nil {
			t.Fatal("zero gold silently skipped required SubGold")
		}
	}()
	new(Server).QuestPenalty54CBD0(u, QuestPenaltyRuntime54CBD0{})
}

func TestQuestPenalty54CBD0NativeRuntimeMissingBindingsFaultPrefix(t *testing.T) {
	for _, missing := range []string{"sub", "gems", "weapon", "armor"} {
		t.Run(missing, func(t *testing.T) {
			p := &Player{}
			p.info[66] = 255
			u := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
			var trace []string
			runtime := QuestPenaltyRuntime54CBD0{
				SubGold:    func(*Object, uint32) { trace = append(trace, "sub") },
				LoseGems:   func(*Object) { trace = append(trace, "gems") },
				LoseWeapon: func(*Object) { trace = append(trace, "weapon") },
				LoseArmor:  func(*Object) { trace = append(trace, "armor") },
			}
			switch missing {
			case "sub":
				runtime.SubGold = nil
			case "gems":
				runtime.LoseGems = nil
			case "weapon":
				runtime.LoseWeapon = nil
			case "armor":
				runtime.LoseArmor = nil
			}
			want := []string{"sub", "gems", "weapon", "armor"}
			for i, event := range want {
				if event == missing {
					want = want[:i]
					break
				}
			}
			defer func() {
				if recover() == nil || len(trace) != len(want) || len(want) != 0 && !reflect.DeepEqual(trace, want) {
					t.Fatalf("missing %s trace=%v, want fault after %v", missing, trace, want)
				}
			}()
			new(Server).QuestPenalty54CBD0(u, runtime)
		})
	}
	for _, unit := range []*Object{nil, {}, {UpdateData: unsafe.Pointer(&PlayerUpdateData{})}} {
		called := false
		func() {
			defer func() {
				if recover() == nil || called {
					t.Fatal("native gold read must fault before subtraction on missing update/Player")
				}
			}()
			new(Server).QuestPenalty54CBD0(unit, QuestPenaltyRuntime54CBD0{SubGold: func(*Object, uint32) { called = true }})
		}()
	}
}
