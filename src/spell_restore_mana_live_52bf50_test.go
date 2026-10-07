package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

// GAME.EXE 0052BF71 adds the unsigned WORD maximum to the entry player,
// then 0052BF76 reloads the acceptance target for audio without another
// nil/class gate. These callbacks exercise that contract, not game outputs.
func TestRestoreManaLive52BF50ReloadsAfterAddition(t *testing.T) {
	for _, tc := range []struct {
		name string
		next uint64
	}{
		{"unchanged", restoreHighTarget52BF20},
		{"different-high-bits-same-low-word", restoreHighTarget52BF20 | uint64(1)<<48},
		{"nil-after-addition", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
			var events []string
			got := restoreMana52BF50(arg, restoreManaHooks52BF50[uint64, *restoreTestArg52BF20]{
				loadTarget: func(record *restoreTestArg52BF20) uint64 {
					if record != arg {
						t.Fatal("acceptance record identity changed")
					}
					events = append(events, "load")
					return record.target
				},
				loadClassLow: func(target uint64) uint8 {
					if target != restoreHighTarget52BF20 {
						t.Fatalf("entry class target = %#x", target)
					}
					events = append(events, "class")
					return 4
				},
				loadMaxMana: func(target uint64) uint16 {
					if target != restoreHighTarget52BF20 {
						t.Fatalf("maximum target = %#x", target)
					}
					events = append(events, "maximum")
					return 0xffff
				},
				addMana: func(target uint64, amount uint16) {
					if target != restoreHighTarget52BF20 || amount != 0xffff {
						t.Fatalf("addition target/amount = %#x/%#x", target, amount)
					}
					events = append(events, "addition")
					arg.target = tc.next
				},
				audio: func(id sound.ID, target uint64) {
					if id != sound.SoundRestoreMana || target != tc.next {
						t.Fatalf("audio id/target = %d/%#x, want %d/%#x", id, target, sound.SoundRestoreMana, tc.next)
					}
					events = append(events, "audio")
				},
			})
			want := []string{"load", "class", "maximum", "addition", "load", "audio"}
			if got != 1 || !reflect.DeepEqual(events, want) {
				t.Fatalf("result/events = %d/%v, want 1/%v", got, events, want)
			}
		})
	}
}

func TestRestoreManaLive52BF50RequiredFaultPrefixes(t *testing.T) {
	order := []string{"entry", "class", "maximum", "addition", "reload", "audio"}
	for i, site := range order {
		t.Run(site, func(t *testing.T) {
			arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
			var events []string
			step := func(call string) {
				events = append(events, call)
				if call == site {
					panic(site)
				}
			}
			faulted := false
			func() {
				defer func() { faulted = recover() == site }()
				restoreMana52BF50(arg, restoreManaHooks52BF50[uint64, *restoreTestArg52BF20]{
					loadTarget: func(record *restoreTestArg52BF20) uint64 {
						if len(events) == 0 {
							step("entry")
						} else {
							step("reload")
						}
						return record.target
					},
					loadClassLow: func(uint64) uint8 { step("class"); return 4 },
					loadMaxMana:  func(uint64) uint16 { step("maximum"); return 0xfedc },
					addMana:      func(uint64, uint16) { step("addition") },
					audio:        func(sound.ID, uint64) { step("audio") },
				})
			}()
			if !faulted || !reflect.DeepEqual(events, order[:i+1]) {
				t.Fatalf("fault/events = %t/%v, want true/%v", faulted, events, order[:i+1])
			}
		})
	}
}
