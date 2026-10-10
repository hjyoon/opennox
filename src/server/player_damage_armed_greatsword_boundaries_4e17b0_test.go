package server

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageArmedGreatSword4E17B0ExecutedServiceBoundaries(t *testing.T) {
	for ti := range 2 {
		for _, missing := range []string{"reflect", "clear-owner", "set-owner", "audio", "rng", "state", "npc-action", "balance", "wear", "admission", "wear-failure", "npc-pop", "live-update"} {
			if ti == 0 && (missing == "rng" || missing == "state") || ti == 1 && (missing == "npc-action" || missing == "npc-pop") {
				continue
			}
			t.Run(fmt.Sprintf("player-%d/%s", ti, missing), func(t *testing.T) {
				mode := 0
				if missing == "npc-pop" {
					mode = 11
				}
				v, a, w, r, events, collect := armedGreatSwordFixtureR84(t, [6]int{0, ti, 0, mode, 0, 4})
				want := []uint64{1, 4}
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				base := []uint64{1, 4, 9, 10, 11, 7}
				action := []uint64{23}
				if ti == 1 {
					action = []uint64{21, 22}
				}
				switch missing {
				case "reflect":
					r.ProjectileReflect = nil
				case "clear-owner":
					r.ClearOwner, want = nil, []uint64{1, 4, 9}
				case "set-owner":
					r.SetOwner, want = nil, []uint64{1, 4, 9, 10}
				case "audio":
					r.Audio, want = nil, base[:5]
				case "rng":
					r.Melee.RandomInt, want = nil, base
				case "state":
					r.PlayerSetState, want = nil, append(slices.Clone(base), 21)
				case "npc-action":
					r.Melee.MonsterBlockAction, want = nil, base
				case "live-update":
					old := r.Audio
					r.Audio = func(id int, got *Object) { old(id, got); got.UpdateData = nil }
					want = base
				case "balance":
					r.BlockDamagePercent, want = nil, append(slices.Clone(base), action...)
				default:
					want = append(append(slices.Clone(base), action...), 13)
					switch missing {
					case "wear":
						r.Melee.DamageBlockWeapon = nil
					case "admission":
						r.Melee.CanDamageBlockWeapon = func(*Object) bool { return false }
					case "wear-failure":
						old := r.Melee.DamageBlockWeapon
						r.Melee.DamageBlockWeapon = func(i, o, s, w *Object, amount float32, typ object.DamageType) bool {
							old(i, o, s, w, amount, typ)
							return false
						}
						want = append(want, 16)
					case "npc-pop":
						r.Melee.MonsterPopBlockAction = nil
						want = append(want, 16)
					}
				}
				h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
				if h != (missing == "wear-failure") || result || reason == "" || !strings.Contains(reason, "armed EXPLOSION GreatSword") || !slices.Equal(*events, want) {
					t.Fatalf("use boundary h/result=%t/%t reason=%q events=%v want=%v", h, result, reason, *events, want)
				}
				if missing == "set-owner" && w.ObjOwner != nil {
					t.Fatal("missing SetOwner rolled back the completed ClearOwner")
				}
				if missing != "live-update" {
					row := collect(result)
					if row[7] != 1 || row[8] != 1174 || row[16] != 0 {
						t.Fatal("missing service replayed the prefix or reached HP")
					}
				}
			})
		}
	}
}

func TestPlayerDamageArmedGreatSword4E17B0UnusedServices(t *testing.T) {
	for ti := range 2 {
		t.Run(fmt.Sprintf("player-%d", ti), func(t *testing.T) {
			v, a, w, r, _, collect := armedGreatSwordFixtureR84(t, [6]int{0, ti, 0, 0, 2, 4})
			r.ClearOwner, r.SetOwner, r.ChangeOwner = nil, nil, nil
			r.DefaultDamage, r.GodMode, r.QuestMode = nil, nil, nil
			r.Melee.CanDamageBlockWeapon = nil
			v.InvFirstItem.HealthData = nil
			v.InvFirstItem.InvNextItem.HealthData = nil
			h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
			row := collect(result)
			if !h || result || row[16] != 0 || row[18] != 1 || row[19] != 0 || w.ObjOwner != a {
				t.Fatal("unneeded owner/HP/availability services blocked a valid original branch")
			}
		})
	}
}

// Invalid live records are negative safety boundaries, not successful PE32
// replay rows. Keep completed prefix/owner/audio/wear effects, but never pass
// a mismatched update layout or nil Player into an actual state/pop service.
func TestPlayerDamageArmedGreatSword4E17B0InvalidLiveActionLayout(t *testing.T) {
	for _, tc := range []struct {
		name      string
		ti        int
		phase     string
		breakItem bool
	}{
		{"player-info-after-audio", 1, "audio", false},
		{"player-info-after-rng", 1, "rng", false},
		{"npc-class-after-audio", 0, "audio", false},
		{"player-info-after-wear", 1, "wear", true},
		{"npc-update-after-wear", 0, "wear", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mode := 0
			if tc.breakItem {
				mode = 11
			}
			v, a, w, r, events, _ := armedGreatSwordFixtureR84(t, [6]int{0, tc.ti, 0, mode, 0, 4})
			invalidate := func() {
				if tc.ti == 1 {
					v.UpdateDataPlayer().Player = nil
				} else if tc.phase == "wear" {
					v.UpdateData = nil
				} else {
					v.ObjClass = object.ClassSimple
				}
			}
			want := []uint64{1, 4, 9, 10, 11, 7}
			switch tc.phase {
			case "audio":
				old := r.Audio
				r.Audio = func(id int, got *Object) { old(id, got); invalidate() }
			case "rng":
				old := r.Melee.RandomInt
				r.Melee.RandomInt = func(min, max int) int { value := old(min, max); invalidate(); return value }
				want = append(want, 21)
			case "wear":
				old := r.Melee.DamageBlockWeapon
				r.Melee.DamageBlockWeapon = func(i, o, s, attack *Object, amount float32, typ object.DamageType) bool {
					result := old(i, o, s, attack, amount, typ)
					invalidate()
					return result
				}
				if tc.ti == 1 {
					want = append(want, 21, 22)
				} else {
					want = append(want, 23)
				}
				want = append(want, 13, 16)
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(v, a, w, 9, object.DamageExplosion, r)
			if h || result || !strings.Contains(reason, "layout") || !slices.Equal(*events, want) || w.ObjOwner != v {
				t.Fatalf("invalid live layout h/result=%t/%t reason=%q events=%v want=%v owner=%p", h, result, reason, *events, want, w.ObjOwner)
			}
		})
	}
}
