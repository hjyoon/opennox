package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/common/sound"
)

// GAME.EXE 0052BF25 caches the first target for 004EE6F0, then 0052BF33
// reloads the same acceptance record for audio. A nil late target is not
// another failed-cast gate. These callbacks are contract tests, not gameplay
// output supplied to a headless scenario.
func TestRestoreHealthLive52BF20ReloadsAfterRecovery(t *testing.T) {
	for _, tc := range []struct {
		name string
		next uint64
	}{
		{"unchanged", restoreHighTarget52BF20},
		{"different-high-bits-same-low-word", restoreHighTarget52BF20 | uint64(1)<<48},
		{"nil-after-recovery", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
			var events []string
			loads := 0
			got := restoreHealth52BF20(arg, restoreHealthHooks52BF20[uint64, *restoreTestArg52BF20]{
				loadTarget: func(record *restoreTestArg52BF20) uint64 {
					if record != arg {
						t.Fatal("acceptance record identity changed")
					}
					loads++
					events = append(events, "load")
					return record.target
				},
				setMaxHP: func(target uint64) {
					if target != restoreHighTarget52BF20 || loads != 1 {
						t.Fatalf("recovery target/loads = %#x/%d", target, loads)
					}
					events = append(events, "recover")
					arg.target = tc.next
				},
				audio: func(id sound.ID, target uint64) {
					if id != sound.SoundRestoreHealth || target != tc.next || loads != 2 {
						t.Fatalf("audio id/target/loads = %d/%#x/%d, want %d/%#x/2", id, target, loads, sound.SoundRestoreHealth, tc.next)
					}
					events = append(events, "audio")
					arg.target = restoreHighTarget52BF20
				},
			})
			if got != 1 || !reflect.DeepEqual(events, []string{"load", "recover", "load", "audio"}) {
				t.Fatalf("result/events = %d/%v", got, events)
			}
		})
	}
}

func TestRestoreHealthLive52BF20NilEntryOnlyLoadsOnce(t *testing.T) {
	loads := 0
	got := restoreHealth52BF20(&restoreTestArg52BF20{}, restoreHealthHooks52BF20[uint64, *restoreTestArg52BF20]{
		loadTarget: func(record *restoreTestArg52BF20) uint64 {
			loads++
			return record.target
		},
		setMaxHP: func(uint64) { t.Fatal("nil entry recovered") },
		audio:    func(sound.ID, uint64) { t.Fatal("nil entry played audio") },
	})
	if got != 0 || loads != 1 {
		t.Fatalf("result/loads = %d/%d, want 0/1", got, loads)
	}
}

func TestRestoreHealthLive52BF20RequiredFaultPrefixes(t *testing.T) {
	for _, site := range []string{"entry", "recovery", "reload", "audio"} {
		t.Run(site, func(t *testing.T) {
			arg := &restoreTestArg52BF20{target: restoreHighTarget52BF20}
			var events []string
			faulted := false
			func() {
				defer func() { faulted = recover() == site }()
				restoreHealth52BF20(arg, restoreHealthHooks52BF20[uint64, *restoreTestArg52BF20]{
					loadTarget: func(record *restoreTestArg52BF20) uint64 {
						loadSite := "entry"
						if len(events) != 0 {
							loadSite = "reload"
							if !reflect.DeepEqual(events, []string{"entry", "recovery"}) {
								t.Fatalf("premature/extra reload: %v", events)
							}
						}
						events = append(events, loadSite)
						if site == loadSite {
							panic(site)
						}
						return record.target
					},
					setMaxHP: func(uint64) {
						events = append(events, "recovery")
						if site == "recovery" {
							panic(site)
						}
					},
					audio: func(sound.ID, uint64) {
						events = append(events, "audio")
						if site == "audio" {
							panic(site)
						}
					},
				})
			}()
			want := []string{"entry"}
			if site != "entry" {
				want = append(want, "recovery")
			}
			if site == "reload" || site == "audio" {
				want = append(want, "reload")
			}
			if site == "audio" {
				want = append(want, "audio")
			}
			if !faulted || !reflect.DeepEqual(events, want) {
				t.Fatalf("fault/events = %t/%v, want true/%v", faulted, events, want)
			}
		})
	}
}
