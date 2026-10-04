package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// Unlike the no-block fixture, these entry masks can select a shield or a
// BLADE weapon block. State belongs to the cached update, after facing; the
// selected inventory item and damage carry belong to the live target.
func normalMeleeBlockCases4E17B0(t *testing.T, fn func(*testing.T, bool, possessionMeleeAttack4E17B0, string)) {
	t.Helper()
	possessionMeleeCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0) {
		defenses := []string{"shield"}
		if kind.typ == object.DamageBlade {
			defenses = append(defenses, "GreatSword", "staff")
		}
		for _, defense := range defenses {
			t.Run(defense, func(t *testing.T) { fn(t, playerSource, kind, defense) })
		}
	})
}

func normalMeleeBlockFixture4E17B0(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) (target, source, weapon, attack, item *Object) {
	t.Helper()
	target, source, weapon, attack = normalMeleePrefixFixture4E17B0(t, playerSource, kind)
	ud := target.UpdateDataPlayer()
	ud.State = PlayerState30 // The prefix, not this entry state, supplies the stance.
	mask := uint32(2)
	ud.Player.ArmorEquip = 0x1000000
	if defense != "shield" {
		ud.Player.ArmorEquip = 0
		mask = 0x400
		if defense == "staff" {
			mask = 0x8000
		}
		ud.Player.WeaponEquip = mask
	}
	item = &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}
	item.InvNextItem = &Object{ObjSubClass: object.SubClass(mask), ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}
	target.InvFirstItem = &Object{ObjSubClass: object.SubClass(mask), InvNextItem: item}
	return
}

func normalMeleeBlockState4E17B0(defense string) PlayerState {
	if defense == "shield" {
		return PlayerState16
	}
	if defense == "staff" {
		return PlayerState21
	}
	return PlayerState13
}

func normalMeleeBlockRuntime4E17B0(t *testing.T, item *Object) PlayerDamageRuntime4E17B0 {
	t.Helper()
	r := damageMeleeRuntimeFixture4E17B0(t)
	r.Audio = func(int, *Object) {}
	r.PlayerSetState = func(v *Object, state PlayerState) bool { v.UpdateDataPlayer().State = state; return true }
	r.Melee.RandomInt = func(int, int) int { return 19 }
	r.BlockDamagePercent = func() float64 { return 0.5 }
	r.CanDamageBlockItem = func(v *Object) bool { return v == item }
	r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
	r.DamageBlockItem = func(_, _, _, _ *Object, _ float32, _ object.DamageType) bool { return true }
	r.Melee.DamageBlockWeapon = r.DamageBlockItem
	return r
}

func TestPlayerDamageNormalMeleeBlockPrefix4E17B0(t *testing.T) {
	normalMeleeBlockCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) {
		type blockCase struct {
			name            string
			state           PlayerState
			front, excluded bool
			option, blocked bool
		}
		cases := []blockCase{
			{"front", normalMeleeBlockState4E17B0(defense), true, false, false, true},
			{"rear", normalMeleeBlockState4E17B0(defense), false, false, false, false},
			{"excluded", normalMeleeBlockState4E17B0(defense), true, true, false, false},
			{"inactive", PlayerState30, true, false, false, false},
		}
		if defense == "shield" {
			cases = append(cases,
				blockCase{"berserk off", PlayerState1, true, false, false, false},
				blockCase{"berserk on", PlayerState1, true, false, true, true})
		} else if defense == "staff" {
			cases[0].state = PlayerState13
			cases = append(cases,
				blockCase{"repeated", PlayerState21, true, false, false, true},
				blockCase{"walking off", PlayerState0, true, false, false, false},
				blockCase{"walking on", PlayerState0, true, false, true, true})
		}
		for _, damage := range []int32{5, 0, -5} {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("damage-%d/%s", damage, tc.name), func(t *testing.T) {
					target, source, weapon, attack, item := normalMeleeBlockFixture4E17B0(t, playerSource, kind, defense)
					cached, position := target.UpdateDataPlayer(), attack.PrevPos
					var events []string
					r := normalMeleeBlockRuntime4E17B0(t, item)
					r.ObserveClear = func(*Object) { t.Fatal("normal block cleared possession") }
					exclude := func(v *Object) bool {
						if v != attack || len(events) != 0 || cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("block exclusion preceded cached marker clear or ran twice")
						}
						events = append(events, "exclude")
						attack.TypeInd++
						attack.PrevPos = types.Ptf(-72, 87)
						return tc.excluded
					}
					wrong := func(*Object) bool { t.Fatal("wrong four/six exclusion set"); return false }
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, wrong
					if weapon == nil {
						r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = wrong, exclude
					}
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || pos != position || !slices.Equal(events, []string{"exclude"}) || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) {
							t.Fatal("block facing lost snapshot/attribution ordering")
						}
						events = append(events, "direction")
						cached.State, cached.Field76, cached.Field75 = tc.state, 9, 321
						cached.Field57 = math.Float32bits(0.9)
						cached.Player.WeaponEquip, cached.Player.ArmorEquip = 0, 0
						return tc.front
					}
					r.BerserkShieldBlock = func(v *Object) bool {
						if v != target || tc.state != PlayerState1 {
							t.Fatal("berserk used stale stance")
						}
						events = append(events, "berserk")
						return tc.option
					}
					r.Melee.StaffWalkingBlock = func() bool { events = append(events, "walking"); return tc.option }
					r.Audio = func(id int, v *Object) {
						want := 878
						if defense == "GreatSword" {
							want = 890
						}
						if defense == "staff" {
							want = 894
						}
						if !tc.blocked || v != target || id != want || cached.Field76 != 9 || cached.Field75 != 321 {
							t.Fatal("block audio lost callback marker")
						}
						events = append(events, "audio")
					}
					r.Melee.RandomInt = func(min, max int) int {
						if defense != "GreatSword" || min != 18 || max != 20 {
							t.Fatal("block RNG")
						}
						events = append(events, "rng")
						return 19
					}
					r.PlayerSetState = func(v *Object, state PlayerState) bool {
						want := PlayerState19
						if defense == "staff" {
							want = PlayerState21
						}
						if v != target || state != want {
							t.Fatal("block state")
						}
						events = append(events, "state")
						cached.State = state
						return true
					}
					r.BlockDamagePercent = func() float64 { events = append(events, "balance"); return 0.5 }
					wear := func(v, owner, attacker, effective *Object, amount float32, typ object.DamageType) bool {
						if v != item || owner != target || attacker != source || effective != attack || amount != float32(damage)*0.5 || typ != kind.typ || cached.Field76 != 9 || cached.Field75 != 321 {
							t.Fatal("signed block wear or first equipped item/marker arguments")
						}
						events = append(events, "wear")
						return true
					}
					r.DamageBlockItem, r.Melee.DamageBlockWeapon = wear, wear
					r.GodMode = func() bool {
						if tc.blocked {
							t.Fatal("blocked hit reached God tail")
						}
						events = append(events, "God")
						return false
					}
					r.QuestMode = func() bool {
						if tc.blocked {
							t.Fatal("blocked hit reached Quest tail")
						}
						events = append(events, "Quest")
						return false
					}
					r.QuestDamageScale = func() float32 { t.Fatal("inactive Quest scaled damage"); return 0 }
					absorption := float64(0.25)
					if kind.typ == object.DamageCrush {
						absorption *= 0.5
					}
					accumulated := float32((1-absorption)*float64(damage)) + 0.125
					effective := playerDamageRound4E17B0(accumulated)
					carry := math.Float32bits(accumulated - float32(effective))
					if damage > 0 && effective == 0 {
						effective = 1
					}
					r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						if tc.blocked || v != target || a != source || w != weapon || d != effective || typ != kind.typ {
							t.Fatal("block/unblocked DefaultDamage arguments")
						}
						events = append(events, "default")
						return true
					}
					want := []string{"exclude"}
					marker, markerType := uint32(1), uint32(attack.TypeInd+1)
					if !tc.excluded {
						want = append(want, "direction")
						marker, markerType = 9, 321
						if tc.front && defense == "shield" && tc.state == PlayerState1 {
							want = append(want, "berserk")
						}
						if tc.front && defense == "staff" && tc.state == PlayerState0 {
							want = append(want, "walking")
						}
					}
					if tc.blocked {
						carry = math.Float32bits(0.125)
						want = append(want, "audio")
						if defense == "GreatSword" {
							want = append(want, "rng")
						}
						if defense != "shield" {
							want = append(want, "state")
						}
						want = append(want, "balance", "wear")
					} else {
						want = append(want, "God", "Quest", "default")
					}
					h, result := PlayerDamageNative4E17B0(target, source, weapon, damage, kind.typ, r)
					if !h || result == tc.blocked || !slices.Equal(events, want) || cached.Field76 != marker || cached.Field75 != markerType || cached.Field21 != carry || target.HealthData.Cur != 200 {
						t.Fatalf("normal block=%t/%t marker=%d/%d carry=%g events=%v want=%v", h, result, cached.Field76, cached.Field75, math.Float32frombits(cached.Field21), events, want)
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalMeleeBlockPrefixNoTail4E17B0(t *testing.T) {
	normalMeleeBlockCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) {
		for _, damage := range []int32{5, 0, -5} {
			t.Run(fmt.Sprintf("damage-%d", damage), func(t *testing.T) {
				target, source, weapon, attack, item := normalMeleeBlockFixture4E17B0(t, playerSource, kind, defense)
				ud := target.UpdateDataPlayer()
				r := normalMeleeBlockRuntime4E17B0(t, item)
				r.DefaultDamage, r.QuestDamageScale, r.ItemArmorValue, r.DamageArmor = nil, nil, nil, nil
				r.QuestMode = func() bool { t.Fatal("blocked hit must not query Quest without its scale"); return true }
				r.GodMode = func() bool { t.Fatal("blocked hit reached God tail"); return true }
				exclude := func(*Object) bool {
					if ud.Field76 != 0 {
						t.Fatal("block prefix did not clear marker")
					}
					return false
				}
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
				r.BlockDirection = func(*Object, types.Pointf) bool { ud.State = normalMeleeBlockState4E17B0(defense); return true }
				wear := 0
				r.DamageBlockItem = func(v, owner, a, w *Object, amount float32, typ object.DamageType) bool {
					if v != item || owner != target || a != source || w != attack || amount != float32(damage)*0.5 || typ != kind.typ {
						t.Fatal("block-only wear")
					}
					wear++
					return true
				}
				r.Melee.DamageBlockWeapon = r.DamageBlockItem
				if h, result := PlayerDamageNative4E17B0(target, source, weapon, damage, kind.typ, r); !h || result || wear != 1 || ud.Field76 != 1 || ud.Field75 != uint32(attack.TypeInd) || ud.Field21 != math.Float32bits(0.125) || target.HealthData.Cur != 200 {
					t.Fatalf("block without HP/armor/Quest services=%t/%t wear=%d", h, result, wear)
				}
			})
		}
	})
}

func TestPlayerDamageNormalMeleeBlockPrefixLive4E17B0(t *testing.T) {
	normalMeleeBlockCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) {
		for _, stage := range []string{"exclusion", "direction"} {
			for _, change := range []string{"replacement front", "replacement rear", "class", "nil update"} {
				t.Run(stage+"/"+change, func(t *testing.T) {
					target, source, weapon, attack, item := normalMeleeBlockFixture4E17B0(t, playerSource, kind, defense)
					cached, position := target.UpdateDataPlayer(), attack.PrevPos
					live := &PlayerUpdateData{Player: &Player{Field3680: 3, CameraFollowObj: source}, State: PlayerState30, Field57: math.Float32bits(0.9), Field21: math.Float32bits(0.5), Field76: 31, Field75: 33}
					liveItem := &Object{ObjSubClass: item.ObjSubClass, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}
					replace := func() {
						cached.Field57, cached.State = math.Float32bits(0.8), normalMeleeBlockState4E17B0(defense)
						cached.Player.WeaponEquip, cached.Player.ArmorEquip = 0, 0
						target.InvFirstItem = liveItem
						switch change {
						case "replacement front", "replacement rear":
							target.UpdateData = unsafe.Pointer(live)
						case "class":
							target.ObjClass = object.ClassSimple
						case "nil update":
							target.UpdateData = nil
						}
					}
					r := normalMeleeBlockRuntime4E17B0(t, liveItem)
					r.ObserveClear = func(*Object) { t.Fatal("replacement observer restarted block prefix") }
					exclude := func(v *Object) bool {
						if v != attack || cached.Field76 != 0 {
							t.Fatal("live block exclusion ordering")
						}
						if stage == "exclusion" {
							replace()
						}
						return false
					}
					r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
					r.BlockDirection = func(v *Object, pos types.Pointf) bool {
						if v != target || pos != position || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) {
							t.Fatal("live block facing ordering")
						}
						if stage == "direction" {
							replace()
						}
						return change != "replacement rear"
					}
					wear, defaultCalls, reason := 0, 0, ""
					r.Audio = func(int, *Object) {
						if change != "replacement front" {
							t.Fatal("invalid live record reached block audio")
						}
					}
					r.DamageBlockItem = func(v, owner, a, w *Object, amount float32, typ object.DamageType) bool {
						if change != "replacement front" || v != liveItem || owner != target || a != source || w != attack || amount != 2.5 || typ != kind.typ || cached.Field76 != 1 || live.Field76 != 31 {
							t.Fatal("block used cached inventory or live marker base")
						}
						wear++
						return true
					}
					r.Melee.DamageBlockWeapon = r.DamageBlockItem
					absorption := float64(0.25)
					if kind.typ == object.DamageCrush {
						absorption *= 0.5
					}
					accumulated := float32((1-absorption)*5) + 0.5
					effective := playerDamageRound4E17B0(accumulated)
					r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
						if change != "replacement rear" || v != target || a != source || w != weapon || d != effective || typ != kind.typ {
							t.Fatal("live block tail arguments/record guard")
						}
						defaultCalls++
						return true
					}
					r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
					h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
					if change == "replacement front" {
						if !h || result || wear != 1 || defaultCalls != 0 || reason != "" || live.Field21 != math.Float32bits(0.5) {
							t.Fatalf("live block=%t/%t wear=%d default=%d reason=%q", h, result, wear, defaultCalls, reason)
						}
					} else if change == "replacement rear" {
						if !h || !result || wear != 0 || defaultCalls != 1 || reason != "" || live.Field21 != math.Float32bits(accumulated-float32(effective)) {
							t.Fatalf("live rear=%t/%t wear=%d default=%d reason=%q", h, result, wear, defaultCalls, reason)
						}
					} else if h || result || wear != 0 || defaultCalls != 0 || reason == "" {
						t.Fatalf("live guard=%t/%t wear=%d default=%d reason=%q", h, result, wear, defaultCalls, reason)
					}
					if cached.Field21 != math.Float32bits(0.125) || cached.Field76 != 1 || cached.Field75 != uint32(attack.TypeInd) || live.Field76 != 31 || live.Field75 != 33 || target.HealthData.Cur != 200 {
						t.Fatal("live block/carry crossed cached marker base")
					}
				})
			}
		}
	})
}

func TestPlayerDamageNormalMeleeBlockPrefixTail4E17B0(t *testing.T) {
	normalMeleeBlockCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) {
		for _, tc := range []struct {
			name                    string
			damage                  int32
			clearMarker, quest, god bool
			dropPlayer              string
		}{
			{"retain callback marker", 5, false, false, false, ""},
			{"fallback cleared marker", 5, true, false, false, ""},
			{"live Quest after wear", 5, false, true, false, ""},
			{"God after wear", 5, false, true, true, ""},
			{"God uses live class after wear", 5, false, true, true, "wear"},
			{"God flag precedes live class", 5, false, true, true, "God"},
			{"positive minimum", 1, false, false, false, ""},
			{"zero", 0, false, false, false, ""},
			{"signed negative", -5, false, false, false, ""},
		} {
			t.Run(tc.name, func(t *testing.T) {
				target, source, weapon, attack, item := normalMeleeBlockFixture4E17B0(t, playerSource, kind, defense)
				ud, position := target.UpdateDataPlayer(), attack.PrevPos
				armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.125)
				armor.InvNextItem = item
				modifier := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
				armor.InitDataModifier().Modifiers[1] = modifier
				if tc.damage == 1 {
					ud.Field21 = math.Float32bits(-0.5)
				}
				r := normalMeleeBlockRuntime4E17B0(t, item)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				var events []string
				exclude := func(v *Object) bool {
					if v != attack || ud.Field76 != 0 || ud.Field75 != 77 || len(events) != 0 {
						t.Fatal("unblocked equipped prefix ordering")
					}
					events = append(events, "exclude")
					return false
				}
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
				r.BlockDirection = func(v *Object, pos types.Pointf) bool {
					if v != target || pos != position || ud.Field76 != 1 || ud.Field75 != uint32(attack.TypeInd) {
						t.Fatal("unblocked equipped facing")
					}
					events = append(events, "direction")
					ud.State, ud.Field76, ud.Field75 = normalMeleeBlockState4E17B0(defense), 9, 321
					return false // Rear damage must enter the ordinary armor/tail.
				}
				r.Audio = func(int, *Object) { t.Fatal("rear melee reached block effects") }
				absorption := float64(0.5)
				if kind.typ == object.DamageCrush {
					absorption *= 0.5
				}
				accumulated := float32((1-absorption)*float64(tc.damage)) + math.Float32frombits(ud.Field21)
				effective := playerDamageRound4E17B0(accumulated)
				carry := math.Float32bits(accumulated - float32(effective))
				wear := float32(tc.damage - effective)
				liveQuest := false
				r.CanApplyArmorDefend = func(m *ModifierEff) bool { return m == modifier }
				r.ApplyArmorDefend = func(m *ModifierEff, v, owner, w, a *Object, d *float32) bool {
					if m != modifier || v != armor || owner != target || w != weapon || a != source || *d != wear || ud.Field21 != carry || ud.Field76 != 9 || ud.Field75 != 321 {
						t.Fatal("equipped tail lost callback marker/live carry or signed wear")
					}
					events = append(events, "wear")
					if tc.clearMarker {
						ud.Field76, ud.Field75 = 0, 99
					}
					liveQuest = tc.quest
					if tc.dropPlayer == "wear" {
						target.ObjClass = object.ClassSimple
					}
					return true
				}
				r.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool { return true }
				r.GodMode = func() bool {
					events = append(events, "God")
					if tc.dropPlayer == "God" {
						target.ObjClass = object.ClassSimple
					}
					return tc.god
				}
				r.QuestMode = func() bool { events = append(events, "Quest"); return liveQuest }
				r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
				if tc.damage > 0 && effective == 0 {
					effective = 1
				}
				god := tc.god && tc.dropPlayer == ""
				if tc.quest && !god {
					before := effective
					effective = playerDamageRound4E17B0(float32(float64(effective) * 0.5))
					if before > 0 && effective < 1 {
						effective = 1
					}
				}
				r.DefaultDamage = func(v, a, w *Object, d int32, typ object.DamageType) bool {
					if v != target || a != source || w != weapon || d != effective || typ != kind.typ {
						t.Fatal("equipped late Quest/God tail arguments")
					}
					events = append(events, "default")
					return true
				}
				want := []string{"exclude", "direction", "wear", "God"}
				if !god {
					want = append(want, "Quest")
					if tc.quest {
						want = append(want, "scale")
					}
					want = append(want, "default")
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, tc.damage, kind.typ, r)
				marker, markerType := uint32(9), uint32(321)
				if tc.clearMarker {
					marker, markerType = 2, uint32(kind.typ)
				}
				if !h || !result || !slices.Equal(events, want) || ud.Field21 != carry || ud.Field76 != marker || ud.Field75 != markerType || target.HealthData.Cur != 200 || armor.HealthData.Cur != 25 {
					t.Fatalf("equipped tail=%t/%t marker=%d/%d events=%v want=%v", h, result, ud.Field76, ud.Field75, events, want)
				}
			})
		}
	})
}

// Prefix service admission is early; block inventory/effects and the unblocked
// armor/default tail are checked only after the original marker/facing stage.
// Missing live data must never enter wear, audio, state, RNG or HP callbacks.
func TestPlayerDamageNormalMeleeBlockPrefixAdmission4E17B0(t *testing.T) {
	normalMeleeBlockCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) {
		for _, service := range []string{"exclusion", "direction", "item", "audio", "balance", "state", "durability check", "durability", "unblocked default", "unblocked armor"} {
			t.Run(service, func(t *testing.T) {
				target, source, weapon, attack, item := normalMeleeBlockFixture4E17B0(t, playerSource, kind, defense)
				ud := target.UpdateDataPlayer()
				r := normalMeleeBlockRuntime4E17B0(t, item)
				excludedCalls, directionCalls, reason := 0, 0, ""
				exclude := func(v *Object) bool {
					if v != attack || ud.Field76 != 0 || ud.Field75 != 77 {
						t.Fatal("missing live block service preceded prefix")
					}
					excludedCalls++
					return false
				}
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
				r.BlockDirection = func(*Object, types.Pointf) bool {
					directionCalls++
					ud.State = normalMeleeBlockState4E17B0(defense)
					return service != "unblocked default" && service != "unblocked armor"
				}
				r.Audio = func(int, *Object) { t.Fatal("missing service reached audio") }
				r.PlayerSetState = func(*Object, PlayerState) bool { t.Fatal("missing service reached state"); return true }
				r.BlockDamagePercent = func() float64 { t.Fatal("missing service reached balance"); return 0.5 }
				r.Melee.RandomInt = func(int, int) int { t.Fatal("missing service reached RNG"); return 19 }
				r.DamageBlockItem = func(_, _, _, _ *Object, _ float32, _ object.DamageType) bool {
					t.Fatal("missing service reached block wear")
					return true
				}
				r.Melee.DamageBlockWeapon = r.DamageBlockItem
				r.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("missing service reached armor wear")
					return true
				}
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("missing service reached HP tail")
					return true
				}
				switch service {
				case "exclusion":
					if weapon == nil {
						r.BlockSourceOnlyExcluded = nil
					} else {
						r.BlockSourceExcluded = nil
					}
				case "direction":
					r.BlockDirection = nil
				case "item":
					target.InvFirstItem = nil
				case "audio":
					r.Audio = nil
				case "balance":
					r.BlockDamagePercent = nil
				case "state":
					r.PlayerSetState = nil
				case "durability check":
					r.CanDamageBlockItem, r.Melee.CanDamageBlockWeapon = nil, nil
				case "durability":
					r.DamageBlockItem, r.Melee.DamageBlockWeapon = nil, nil
				case "unblocked default":
					r.DefaultDamage = nil
				case "unblocked armor":
					armor := damageMeleeArmorFixture4E17B0(target, 0.25, 0.125)
					armor.InvNextItem = item
					r.ItemArmorValue = nil
				}
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				before, beforeUD, beforePlayer, beforeItem := *target, *ud, *ud.Player, *item
				beforeArmor := target.InvFirstItem
				var armorCopy Object
				if beforeArmor != nil {
					armorCopy = *beforeArmor
				}
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
				if service == "exclusion" || service == "direction" {
					if excludedCalls != 0 || directionCalls != 0 {
						t.Fatal("missing prefix service executed prefix")
					}
				} else {
					if excludedCalls != 1 || directionCalls != 1 {
						t.Fatal("late service skipped or repeated prefix")
					}
					beforeUD.Field76, beforeUD.Field75, beforeUD.State = 1, uint32(attack.TypeInd), normalMeleeBlockState4E17B0(defense)
				}
				if h || result || reason == "" || *target != before || *ud != beforeUD || *ud.Player != beforePlayer || *item != beforeItem || target.HealthData.Cur != 200 || (beforeArmor != nil && *beforeArmor != armorCopy) {
					t.Fatalf("phase admission=%t/%t reason=%q exclusions=%d directions=%d", h, result, reason, excludedCalls, directionCalls)
				}
			})
		}
	})
}

func TestPlayerDamageNormalMeleeBlockPrefixEarly4E17B0(t *testing.T) {
	normalMeleeBlockCases4E17B0(t, func(t *testing.T, playerSource bool, kind possessionMeleeAttack4E17B0, defense string) {
		for _, gate := range []string{"no update", "dead", "invulnerable", "observer", "Coop self"} {
			t.Run(gate, func(t *testing.T) {
				target, source, weapon, _, item := normalMeleeBlockFixture4E17B0(t, playerSource, kind, defense)
				ud := target.UpdateDataPlayer()
				r := normalMeleeBlockRuntime4E17B0(t, item)
				switch gate {
				case "no update":
					target.ObjFlags |= object.FlagNoUpdate
				case "dead":
					target.ObjFlags |= object.FlagDead
				case "invulnerable":
					target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
				case "observer":
					ud.Player.Field3680 |= 1
				case "Coop self":
					source = target
					r.CoopMode = func() bool { return true }
				}
				exclude := func(*Object) bool { t.Fatal("early gate consumed equipped prefix"); return false }
				r.BlockSourceExcluded, r.BlockSourceOnlyExcluded = exclude, exclude
				r.BlockDirection = func(*Object, types.Pointf) bool { t.Fatal("early gate reached equipped facing"); return false }
				r.Audio = func(id int, v *Object) {
					if gate != "invulnerable" || id != 71 || v != target {
						t.Fatal("equipped early audio")
					}
				}
				before, beforeUD, beforePlayer, beforeItem := *target, *ud, *ud.Player, *item
				h, result := PlayerDamageNative4E17B0(target, source, weapon, 5, kind.typ, r)
				if !h || result != (gate == "invulnerable") || *target != before || *ud != beforeUD || *ud.Player != beforePlayer || *item != beforeItem {
					t.Fatalf("equipped early gate=%t/%t", h, result)
				}
			})
		}
	})
}
