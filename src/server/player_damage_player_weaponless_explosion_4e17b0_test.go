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

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionArmor(t *testing.T) {
	for _, kind := range []string{"none", "player", "npc", "self"} {
		for _, protected := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/protected-%t", kind, protected), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, true)
				source := damageWeaponlessExplosionSource4E17B0(t, target, kind)
				armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
				ud := target.UpdateDataPlayer()
				ud.Field76, ud.Field75 = 99, 77
				protection, wantHP := float64(0), uint16(195)
				if protected {
					protection, wantHP = 0.5, 198
				}
				r := damageFlameRuntime4E17B0(t, protection)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
				if h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r); !h || !result {
					t.Fatalf("handled=%t result=%t", h, result)
				}
				pos := types.Pointf{}
				if source != nil {
					pos = source.PrevPos
				}
				if target.HealthData.Cur != wantHP || armor.HealthData.Cur != 21 ||
					ud.Field76 != 2 || ud.Field75 != 7 || math.Abs(float64(math.Float32frombits(ud.Field21))+0.1) > 1e-6 ||
					target.Obj130 != source || target.Pos132 != pos || target.Field131 != 7 || target.Frame134 != 1400 {
					t.Fatalf("HP=%d/%d armor=%d marker=%d/%d carry=%g source=%p", target.HealthData.Cur, wantHP,
						armor.HealthData.Cur, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21), target.Obj130)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionRounding(t *testing.T) {
	for _, kind := range []string{"none", "player", "npc", "self"} {
		for _, tc := range []struct {
			raw, effective int32
			armor          float32
			carry          uint32
		}{
			{9, 4, 0.5, math.Float32bits(0.5)}, {11, 6, 0.5, math.Float32bits(-0.5)},
			{-9, -4, 0.5, math.Float32bits(-0.5)}, {1, 1, 1, 0}, {0, 0, 0.5, 0},
			{3, 3, 0.1, 0xbe999998}, {9, 7, 0.2, 0x3e4cccc0}, {3, 2, 0.4, 0xbe4cccd0},
		} {
			t.Run(fmt.Sprintf("%s/%d/%g", kind, tc.raw, tc.armor), func(t *testing.T) {
				target := damageMeleeUnitFixture4E17B0(t, true)
				source := damageWeaponlessExplosionSource4E17B0(t, target, kind)
				ud := target.UpdateDataPlayer()
				ud.Field57, ud.Field76, ud.Field75 = math.Float32bits(tc.armor), 99, 77
				r := damageMeleeRuntimeFixture4E17B0(t)
				got := int32(math.MaxInt32)
				r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
					if v != target || s != source || w != nil || typ != object.DamageExplosion {
						t.Fatal("weapon-less player explosion identity changed")
					}
					got = amount
					return false
				}
				if h, result := PlayerDamageNative4E17B0(target, source, nil, tc.raw, object.DamageExplosion, r); !h || result {
					t.Fatalf("handled=%t result=%t", h, result)
				}
				if got != tc.effective || ud.Field76 != 2 || ud.Field75 != 7 || ud.Field21 != tc.carry {
					t.Fatalf("damage=%d/%d marker=%d/%d carry=%#x/%#x", got, tc.effective, ud.Field76, ud.Field75, ud.Field21, tc.carry)
				}
			})
		}
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionEarly(t *testing.T) {
	for _, gate := range []string{"no-update", "dead", "invulnerable-sound", "invulnerable-silent", "Coop-self", "player-status"} {
		t.Run(gate, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 99, 77
			source := damageMeleeUnitFixture4E17B0(t, false)
			r := damageMeleeRuntimeFixture4E17B0(t)
			sounds, coops := 0, 0
			r.CoopMode = func() bool { coops++; return gate == "Coop-self" }
			r.Audio = func(id int, v *Object) {
				if gate != "invulnerable-sound" || id != 71 || v != target {
					t.Fatal("early sound identity")
				}
				sounds++
			}
			r.Frame = func() uint32 {
				if gate == "invulnerable-silent" {
					return 1401
				}
				if gate != "invulnerable-sound" {
					t.Fatal("unbuffed early entry queried frame")
				}
				return 1400
			}
			switch gate {
			case "no-update", "dead":
				target.ObjFlags |= map[string]object.Flags{"no-update": object.FlagNoUpdate, "dead": object.FlagDead}[gate]
				target.UpdateData = nil
			case "invulnerable-sound", "invulnerable-silent":
				target.Buffs |= 1 << 23
				target.UpdateData = nil
			case "Coop-self":
				source = target
				target.UpdateData = nil // Owner gate must precede the cached player record.
			case "player-status":
				ud.Player.Field3680 |= 1
			}
			r.BlockSourceOnlyExcluded = func(*Object) bool { t.Fatal("early gate reached exclusions"); return false }
			r.ObserveClear = func(*Object) { t.Fatal("early gate cleared possession") }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("early gate reached HP")
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			wantResult := gate == "invulnerable-sound" || gate == "invulnerable-silent"
			wantCoops := 0
			if gate == "Coop-self" || gate == "player-status" {
				wantCoops = 1
			}
			if !h || result != wantResult || ud.Field76 != 99 || ud.Field75 != 77 || target.HealthData.Cur != 200 ||
				coops != wantCoops || sounds != map[bool]int{false: 0, true: 1}[gate == "invulnerable-sound"] {
				t.Fatalf("handled/result=%t/%t marker=%d/%d Coop=%d/%d sounds=%d", h, result, ud.Field76, ud.Field75, coops, wantCoops, sounds)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionCachedLiveOrder(t *testing.T) {
	for _, callbackMarker := range []bool{false, true} {
		t.Run(fmt.Sprintf("callback-marker-%t", callbackMarker), func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			cached := target.UpdateDataPlayer()
			cached.Field76, cached.Field75 = 99, 77
			cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
			live := &PlayerUpdateData{Player: &Player{}, State: PlayerState13, Field57: math.Float32bits(0.25), Field21: math.Float32bits(0.25), Field76: 33, Field75: 44}
			pos := source.PrevPos
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
			var events []string
			r.ObserveClear = func(v *Object) {
				if v != target || cached.Field76 != 0 || cached.Field75 != 77 {
					t.Fatal("ObserveClear preceded cached marker clear")
				}
				events = append(events, "observe")
				target.UpdateData = unsafe.Pointer(live)
			}
			r.BlockSourceExcluded = func(*Object) bool { t.Fatal("nil weapon used six-type exclusions"); return true }
			r.BlockSourceOnlyExcluded = func(v *Object) bool {
				if v != source || cached.Field76 != 0 {
					t.Fatal("source-only explosion was marked as a weapon")
				}
				events = append(events, "exclude")
				source.PrevPos = types.Ptf(999, 999)
				return false
			}
			r.BlockDirection = func(v *Object, p types.Pointf) bool {
				if v != target || p != pos {
					t.Fatal("facing lost pre-exclusion PrevPos")
				}
				events = append(events, "facing")
				if callbackMarker {
					cached.Field76, cached.Field75 = 88, 89
				}
				return false
			}
			r.DamageArmor = func(item, s, w *Object, amount int32, typ object.DamageType) bool {
				if item != armor || s != source || w != nil || amount != 4 || typ != object.DamageExplosion ||
					cached.Field21 != math.Float32bits(0.4) || live.Field21 != math.Float32bits(-0.25) ||
					live.Field76 != 33 || live.Field75 != 44 {
					t.Fatal("armor used stale carry or duplicated marker/armor prefix")
				}
				events = append(events, "armor")
				return true
			}
			r.GodMode = func() bool { events = append(events, "God"); return false }
			r.QuestMode = func() bool { events = append(events, "Quest"); return true }
			r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0.5 }
			r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
				wantMark, wantType := uint32(2), uint32(7)
				if callbackMarker {
					wantMark, wantType = 88, 89
				}
				if v != target || s != source || w != nil || amount != 2 || typ != object.DamageExplosion ||
					cached.Field76 != wantMark || cached.Field75 != wantType || live.Field76 != 33 || live.Field75 != 44 {
					t.Fatal("late Quest/default cached-marker identity lost")
				}
				events = append(events, "default")
				return false
			}
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			if !h || result || !slices.Equal(events, []string{"observe", "exclude", "facing", "armor", "God", "Quest", "scale", "default"}) {
				t.Fatalf("handled=%t result=%t events=%v", h, result, events)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionShield(t *testing.T) {
	for _, stance := range []string{"ordinary", "berserker"} {
		for _, reflected := range []string{"unit", "keeps-owner", "transfers-owner", "loses-missile"} {
			for _, broken := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/broken-%t", stance, reflected, broken), func(t *testing.T) {
					target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
					cached := target.UpdateDataPlayer()
					cached.Player.ArmorEquip, cached.Player.WeaponEquip, cached.Field76, cached.Field75 = 0x3000000, 0, 99, 77
					cached.State = PlayerState16
					if stance == "berserker" {
						cached.State = PlayerState1
					}
					target.Buffs |= 1 << playerDamageReflectEnchant4E17B0 // A unit Explosion must not use Reflect's PosVec test.
					shield := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped}
					live := &PlayerUpdateData{Player: &Player{}, State: PlayerState13, Field76: 33, Field75: 44}
					var events []string
					r := damageMeleeRuntimeFixture4E17B0(t)
					r.BlockSourceExcluded = func(*Object) bool { t.Fatal("nil weapon used six-type exclusions"); return false }
					r.BlockSourceOnlyExcluded = func(v *Object) bool {
						if v != source || cached.Field76 != 0 || cached.Field75 != 77 {
							t.Fatal("exclusion preceded marker clear")
						}
						events = append(events, "exclude")
						cached.Player.ArmorEquip = 0 // Entry equipment remains cached.
						target.UpdateData = unsafe.Pointer(live)
						return false
					}
					r.BlockDirection = func(v *Object, p types.Pointf) bool {
						if v != target || p != source.PrevPos {
							t.Fatal("unit Explosion incorrectly entered Reflect direction")
						}
						events = append(events, "facing")
						return true
					}
					r.BerserkShieldBlock = func(v *Object) bool {
						if stance != "berserker" || v != target {
							t.Fatal("ordinary shield queried berserker action")
						}
						events = append(events, "berserker")
						return true
					}
					r.Audio = func(id int, v *Object) {
						if id != 878 || v != target || cached.Field76 != 0 {
							t.Fatal("shield audio/marker identity")
						}
						events = append(events, "audio")
						if reflected != "unit" {
							source.ObjClass, source.ObjSubClass = object.ClassMissile, 0
						}
					}
					r.ProjectileReflect = func(v, by *Object) {
						if v != source || by != target || reflected == "unit" {
							t.Fatal("reflection identity")
						}
						events = append(events, "reflect")
						if reflected == "keeps-owner" {
							source.ObjSubClass = 2 // Must reload after reflection.
						}
						if reflected == "loses-missile" {
							source.ObjClass = object.ClassMonster
						}
					}
					r.ClearOwner = func(v *Object) {
						if v != source || reflected != "transfers-owner" {
							t.Fatal("stale class/subclass ownership clear")
						}
						events = append(events, "clear")
					}
					r.SetOwner = func(by, v *Object) {
						if by != target || v != source {
							t.Fatal("owner identity")
						}
						events = append(events, "owner")
					}
					r.BlockDamagePercent = func() float64 {
						events = append(events, "balance")
						target.InvFirstItem = shield // Shield selection is live, after balance.
						return 0.25
					}
					r.CanDamageBlockItem = func(v *Object) bool { return v == shield }
					r.DamageBlockItem = func(v, by, s, w *Object, amount float32, typ object.DamageType) bool {
						if v != shield || by != target || s != source || w != nil || amount != 2.25 || typ != object.DamageExplosion {
							t.Fatal("live shield wear identity/order")
						}
						events = append(events, "wear")
						if broken {
							shield.ObjFlags |= object.FlagDestroyed
						}
						return true
					}
					r.PlayerSetState = func(v *Object, state PlayerState) bool {
						if !broken || v != target || state != PlayerState13 || v.UpdateDataPlayer() != live {
							t.Fatal("broken shield used cached player state")
						}
						events = append(events, "state")
						return true
					}
					r.GodMode = func() bool { t.Fatal("blocked explosion queried GodMode"); return false }
					r.QuestMode = func() bool { t.Fatal("blocked explosion queried Quest"); return false }
					r.DefaultDamage = nil // A complete block needs no HP tail service.
					h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
					want := []string{"exclude", "facing"}
					if stance == "berserker" {
						want = append(want, "berserker")
					}
					want = append(want, "audio")
					if reflected != "unit" {
						want = append(want, "reflect")
					}
					if reflected == "transfers-owner" {
						want = append(want, "clear", "owner")
					}
					want = append(want, "balance", "wear")
					if broken {
						want = append(want, "state")
					}
					if !h || result || target.HealthData.Cur != 200 || cached.Field76 != 0 || cached.Field75 != 77 ||
						live.Field76 != 33 || live.Field75 != 44 || !slices.Equal(events, want) {
						t.Fatalf("handled=%t result=%t events=%v want=%v", h, result, events, want)
					}
				})
			}
		}
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionBlockAlternatives(t *testing.T) {
	for _, alternative := range []string{"nil-source", "excluded", "behind", "great-sword", "staff", "no-shield", "berserker-off"} {
		t.Run(alternative, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
			ud := target.UpdateDataPlayer()
			ud.State, ud.Player.ArmorEquip = PlayerState13, 0x3000000
			if alternative == "nil-source" {
				source = nil
			}
			if alternative == "great-sword" {
				ud.Player.WeaponEquip = 0x400
			}
			if alternative == "staff" {
				ud.Player.WeaponEquip = 0x8000
			}
			if alternative == "no-shield" {
				ud.State, ud.Player.ArmorEquip = PlayerState16, 0
			}
			if alternative == "berserker-off" {
				ud.State = PlayerState1
			}
			r := damageMeleeRuntimeFixture4E17B0(t)
			queries, directions, berserks := 0, 0, 0
			r.BlockSourceOnlyExcluded = func(*Object) bool { queries++; return alternative == "excluded" }
			r.BlockDirection = func(*Object, types.Pointf) bool { directions++; return alternative != "behind" }
			r.BerserkShieldBlock = func(*Object) bool { berserks++; return false }
			r.Audio = func(int, *Object) { t.Fatal("unit case 7 entered unsupported great-sword/staff/shield alternative") }
			r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
				if v != target || s != source || w != nil || amount != 9 || typ != object.DamageExplosion || ud.Field76 != 2 || ud.Field75 != 7 {
					t.Fatal("unblocked alternative lost damage type marker")
				}
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			wantQueries, wantDirections := 1, 1
			if alternative == "nil-source" {
				wantQueries, wantDirections = 0, 0
			}
			if alternative == "excluded" {
				wantDirections = 0
			}
			wantBerserks := map[bool]int{false: 0, true: 1}[alternative == "berserker-off"]
			if !h || !result || queries != wantQueries || directions != wantDirections || berserks != wantBerserks {
				t.Fatalf("handled/result=%t/%t exclusions/directions/berserks=%d/%d/%d", h, result, queries, directions, berserks)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionLateMode(t *testing.T) {
	for _, mode := range []string{"God", "Quest-positive", "Quest-zero", "Quest-signed"} {
		t.Run(mode, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0)
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 99, 77
			raw, effective, wear, carry := int32(1), int32(0), int32(1), float32(0.5)
			wantDefault := int32(1)
			if mode == "Quest-zero" {
				raw, effective, wear, carry, wantDefault = 0, 0, 0, 0, 0
			}
			if mode == "Quest-signed" {
				raw, effective, wear, carry, wantDefault = -9, -4, -5, -0.5, 0
			}
			var events []string
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			r.ItemArmorValue = func(v *Object) float32 {
				if v != armor || ud.Field76 != 0 || ud.Field75 != 77 || ud.Field21 != math.Float32bits(carry) {
					t.Fatal("armor lookup did not precede minimum/marker/mode checks")
				}
				events = append(events, "lookup")
				return 0.5
			}
			r.DamageArmor = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
				if v != armor || s != nil || w != nil || amount != wear || typ != object.DamageExplosion ||
					ud.Field76 != 0 || ud.Field75 != 77 || ud.Field21 != math.Float32bits(carry) || raw-effective != wear {
					t.Fatal("armor wear did not precede minimum/marker/mode checks")
				}
				events = append(events, "wear")
				return true
			}
			r.GodMode = func() bool {
				if ud.Field76 != 2 || ud.Field75 != 7 {
					t.Fatal("GodMode preceded marker fallback")
				}
				events = append(events, "God")
				return mode == "God"
			}
			r.QuestMode = func() bool {
				if mode == "God" {
					t.Fatal("GodMode queried Quest")
				}
				events = append(events, "Quest")
				return true
			}
			r.QuestDamageScale = func() float32 { events = append(events, "scale"); return 0 }
			r.DefaultDamage = nil
			if mode != "God" {
				r.DefaultDamage = func(v, s, w *Object, amount int32, typ object.DamageType) bool {
					if v != target || s != nil || w != nil || amount != wantDefault || typ != object.DamageExplosion {
						t.Fatal("Quest signed/zero/minimum identity")
					}
					events = append(events, "default")
					return false
				}
			}
			h, result := PlayerDamageNative4E17B0(target, nil, nil, raw, object.DamageExplosion, r)
			want := []string{"lookup"}
			if wear > 0 {
				want = append(want, "wear")
			} // EquipDamage's existing zero/signed durability no-op.
			want = append(want, "God")
			if mode != "God" {
				want = append(want, "Quest", "scale", "default")
			}
			if !h || result != (mode == "God") || !slices.Equal(events, want) {
				t.Fatalf("handled/result=%t/%t events=%v want=%v", h, result, events, want)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionLiveRecordGuard(t *testing.T) {
	for _, at := range []string{"observe", "facing", "armor"} {
		t.Run(at, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			cached := target.UpdateDataPlayer()
			cached.Field76, cached.Field75 = 99, 77
			cached.Player.Field3680, cached.Player.CameraFollowObj = 2, source
			live := &MonsterUpdateData{Field547: 33, Field546: 44}
			change := func() { target.ObjClass = object.ClassMonster; target.UpdateData = unsafe.Pointer(live) }
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			r.ObserveClear = func(*Object) {
				if at == "observe" {
					change()
				}
			}
			r.BlockDirection = func(*Object, types.Pointf) bool {
				if at == "facing" {
					change()
				}
				return at == "facing"
			}
			r.DamageArmor = func(*Object, *Object, *Object, int32, object.DamageType) bool {
				if at == "armor" {
					change()
				}
				return true
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			r.DefaultDamage = func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				t.Fatal("unsupported live record reached HP")
				return true
			}
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			if h || result || reason == "" || cached.Field76 != 0 || cached.Field75 != 77 || live.Field547 != 33 || live.Field546 != 44 || target.HealthData.Cur != 200 {
				t.Fatalf("live layout was silently reinterpreted: handled/result=%t/%t reason=%s", h, result, reason)
			}
		})
	}
}

func TestPlayerDamageNative4E17B0PlayerWeaponlessExplosionMissingServices(t *testing.T) {
	for _, missing := range []string{"player", "observe", "exclude", "direction", "armor", "default", "Quest"} {
		t.Run(missing, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, false)
			armor := damageMeleeArmorFixture4E17B0(target, 0.5, 0.4)
			ud := target.UpdateDataPlayer()
			ud.Field76, ud.Field75 = 99, 77
			r := damageMeleeRuntimeFixture4E17B0(t)
			damageMeleeArmorRuntime4E17B0(&r, armor, 0.5)
			switch missing {
			case "player":
				ud.Player = nil
			case "observe":
				ud.Player.Field3680, ud.Player.CameraFollowObj = 2, source
				r.ObserveClear = nil
			case "exclude":
				r.BlockSourceOnlyExcluded = nil
			case "direction":
				r.BlockDirection = nil
			case "armor":
				r.DamageArmor = nil
			case "default":
				r.DefaultDamage = nil
			case "Quest":
				r.QuestMode = func() bool { return true }
				r.QuestDamageScale = nil
			}
			reason := ""
			r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			h, result := PlayerDamageNative4E17B0(target, source, nil, 9, object.DamageExplosion, r)
			if h || result || reason == "" || target.HealthData.Cur != 200 {
				t.Fatalf("missing service silently reached HP: handled/result=%t/%t reason=%s", h, result, reason)
			}
			late := missing == "default" || missing == "Quest"
			if !late && (ud.Field21 != math.Float32bits(0.4) || armor.HealthData.Cur != 25) {
				t.Fatal("failed admission modified carry/durability")
			}
			if late && (ud.Field76 != 2 || ud.Field75 != 7 || armor.HealthData.Cur != 21) {
				t.Fatal("late failure rolled back the original damage prefix")
			}
		})
	}
}
