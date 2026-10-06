package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// These are callback-order contracts, not stock-map or HP evidence. The
// separate native test exercises the registered C dispatcher and real HP tail.
func TestPlayerDamageNPCCloudPoison4E17B0SignedQuest(t *testing.T) {
	for _, owner := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
		for _, raw := range []int32{-7, 0, 3, 10, 19} {
			for _, scale := range []float32{-1, 0, 0.5, 1.25} {
				t.Run(fmt.Sprintf("%s/raw-%d/scale-%g", owner, raw, scale), func(t *testing.T) {
					target, source, cloud := defaultDamageCloudFixture4E0B30(t, owner, 0x11012)
					update := target.UpdateDataMonster()
					update.Field518, update.ArmorEquipFlags, update.WeaponEquipFlags = math.Float32bits(0.9), 0x3000000, 0x400
					update.AIStack[0].Action = 21
					target.Buffs |= 1 << playerDamageReflectEnchant4E17B0
					armorHealth, freeArmorHealth := alloc.New(HealthData{})
					armor, freeArmor := alloc.New(Object{})
					t.Cleanup(freeArmorHealth)
					t.Cleanup(freeArmor)
					*armorHealth = HealthData{Cur: 50, Max: 50}
					*armor = Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, HealthData: armorHealth}
					target.InvFirstItem = armor
					before, armorBefore := *update, *target.InvFirstItem.HealthData
					marker, markerType := uint32(1), uint32(cloud.TypeInd)
					if source == nil || source == cloud {
						marker, markerType = 2, 5
					}
					var events []string
					r := PlayerDamageRuntime4E17B0{
						Frame: func() uint32 { t.Fatal("ordinary cloud hit read invulnerability frame"); return 1400 },
						GodMode: func() bool {
							if update.Field547 != marker || update.Field546 != markerType {
								t.Fatal("GodMode preceded case-5 marker")
							}
							events = append(events, "god")
							return true // The global flag does not protect an NPC.
						},
						BlockSourceExcluded: func(got *Object) bool {
							if source == nil || got != cloud || update.Field547 != 0 || update.Field546 != 91 {
								t.Fatal("cached marker clear must precede cloud exclusions")
							}
							cloud.TypeInd = 1222 // Marker type must be read after exclusions.
							if source != cloud {
								markerType = 1222
								before.Field546 = markerType
							}
							events = append(events, "excluded")
							return true
						},
						BlockDirection:     func(*Object, types.Pointf) bool { t.Fatal("excluded cloud checked facing"); return true },
						ObserveClear:       func(*Object) { t.Fatal("NPC used player observer layout") },
						ItemArmorValue:     func(*Object) float32 { t.Fatal("poison queried armor"); return 0 },
						CanDamageArmor:     func(*Object) bool { t.Fatal("poison preflighted armor"); return false },
						ElectricArmorScale: func(*Object) float32 { t.Fatal("poison used electric armor"); return 0 },
						ShieldReduce:       func(*Object, *int32, object.DamageType, *Object) { t.Fatal("Shield reduced poison") },
						Audio:              func(int, *Object) { t.Fatal("Reflect/physical shield blocked excluded nonmissile poison") },
						QuestMode:          func() bool { events = append(events, "quest"); return scale >= 0 },
						QuestDamageScale:   func() float32 { events = append(events, "scale"); return scale },
						DefaultDamage: func(got, attacker, attack *Object, damage int32, typ object.DamageType) bool {
							want := raw
							if scale >= 0 {
								want = int32(math.RoundToEven(float64(float32(float64(raw) * float64(scale)))))
								if raw > 0 && want < 1 {
									want = 1
								}
							}
							if got != target || attacker != source || attack != cloud || damage != want || typ != object.DamagePoison ||
								update.Field547 != marker || update.Field546 != markerType {
								t.Fatalf("NPC cloud default args damage=%d want=%d marker=%d/%d", damage, want, update.Field547, update.Field546)
							}
							events = append(events, "default")
							return true
						},
						Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
					}
					h, result := PlayerDamageNative4E17B0(target, source, cloud, raw, object.DamagePoison, r)
					wantEvents := []string{"god", "quest"}
					if source != nil {
						wantEvents = append([]string{"excluded"}, wantEvents...)
					}
					if scale >= 0 {
						wantEvents = append(wantEvents, "scale")
					}
					wantEvents = append(wantEvents, "default")
					before.Field547, before.Field546 = marker, markerType
					if !h || !result || !slices.Equal(events, wantEvents) || *update != before || target.HealthData.Cur != 20 || *target.InvFirstItem.HealthData != armorBefore {
						t.Fatalf("NPC cloud result=%t/%t events=%v want=%v marker=%d/%d carry=%g", h, result, events, wantEvents, update.Field547, update.Field546, math.Float32frombits(update.Field1))
					}
				})
			}
		}
	}
}

func TestPlayerDamageNPCCloudPoison4E17B0CachedLiveAndPosition(t *testing.T) {
	for _, excluded := range []bool{false, true} {
		for _, stage := range []string{"excluded", "direction", "god", "quest", "scale"} {
			if excluded && stage == "direction" {
				continue
			}
			t.Run(fmt.Sprintf("excluded-%t/%s", excluded, stage), func(t *testing.T) {
				target, source, cloud := defaultDamageCloudFixture4E0B30(t, "player", 0x11012)
				cached := target.UpdateDataMonster()
				live, freeLive := alloc.New(MonsterUpdateData{})
				t.Cleanup(freeLive)
				*live = MonsterUpdateData{Field547: 31, Field546: 33, Field1: math.Float32bits(-0.5), Field523_2: 88}
				liveBefore, oldPos := *live, cloud.PrevPos
				mutate := func(at string) {
					if stage == at {
						target.UpdateData = unsafe.Pointer(live)
					}
				}
				calls := 0
				r := PlayerDamageRuntime4E17B0{
					BlockSourceExcluded: func(*Object) bool {
						if cached.Field547 != 0 || cached.Field546 != 91 {
							t.Fatal("exclusion prefix")
						}
						mutate("excluded")
						cloud.PrevPos = types.Ptf(200, 300)
						return excluded
					},
					BlockDirection: func(got *Object, pos types.Pointf) bool {
						if excluded || got != target || pos != oldPos || cached.Field547 != 1 {
							t.Fatal("cached facing position/marker")
						}
						mutate("direction")
						return false
					},
					GodMode:          func() bool { mutate("god"); return false },
					QuestMode:        func() bool { mutate("quest"); return true },
					QuestDamageScale: func() float32 { mutate("scale"); return 0.5 },
					DefaultDamage: func(got, attacker, attack *Object, damage int32, typ object.DamageType) bool {
						if got != target || attacker != source || attack != cloud || damage != 4 || typ != object.DamagePoison || cached.Field547 != 1 || cached.Field546 != uint32(cloud.TypeInd) {
							t.Fatal("cached NPC poison tail")
						}
						calls++
						return false
					},
					Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
				}
				h, result := PlayerDamageNative4E17B0(target, source, cloud, 9, object.DamagePoison, r)
				if !h || result || calls != 1 || *live != liveBefore || cached.Field1 != math.Float32bits(0.25) || target.HealthData.Cur != 20 {
					t.Fatalf("cached/live cloud result=%t/%t calls=%d", h, result, calls)
				}
			})
		}
	}
}

func TestPlayerDamageNPCCloudPoison4E17B0EarlyAndShape(t *testing.T) {
	for _, tc := range []struct {
		name           string
		flags          object.Flags
		invulnerable   bool
		frame          uint32
		result         bool
		frames, sounds int
	}{
		{name: "no-update", flags: object.FlagNoUpdate},
		{name: "dead", flags: object.FlagDead},
		{name: "no-update-before-invulnerable", flags: object.FlagNoUpdate, invulnerable: true},
		{name: "invulnerable-sound", invulnerable: true, frame: 1400, result: true, frames: 1, sounds: 1},
		{name: "invulnerable-quiet", invulnerable: true, frame: 1401, result: true, frames: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source, cloud := defaultDamageCloudFixture4E0B30(t, "NPC", 0x11012)
			target.ObjFlags = tc.flags
			if tc.invulnerable {
				target.Buffs |= 1 << playerDamageInvulnerableEnchant4E17B0
			}
			before := *target.UpdateDataMonster()
			frames, sounds := 0, 0
			r := PlayerDamageRuntime4E17B0{
				Frame: func() uint32 { frames++; return tc.frame },
				Audio: func(sound int, got *Object) {
					if sound != 71 || got != target {
						t.Fatal("early audio")
					}
					sounds++
				},
				BlockSourceExcluded: func(*Object) bool { t.Fatal("early cloud exclusion"); return true },
				DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("early cloud default")
					return false
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			h, result := PlayerDamageNative4E17B0(target, source, cloud, 19, object.DamagePoison, r)
			if !h || result != tc.result || frames != tc.frames || sounds != tc.sounds || *target.UpdateDataMonster() != before || target.HealthData.Cur != 20 {
				t.Fatalf("early result=%t/%t frames=%d sounds=%d", h, result, frames, sounds)
			}
		})
	}
	for _, tc := range []struct {
		name  string
		class object.Class
		typ   object.DamageType
		want  bool
	}{
		{"stock", 0x190008, object.DamagePoison, true},
		{"simple-dangerous", object.ClassSimple | object.ClassDangerous, object.DamagePoison, true},
		{"nil-class", 0, object.DamagePoison, false},
		{"simple-only", object.ClassSimple, object.DamagePoison, false},
		{"dangerous-only", object.ClassDangerous, object.DamagePoison, false},
		{"missile", 0x190008 | object.ClassMissile, object.DamagePoison, false},
		{"player", 0x190008 | object.ClassPlayer, object.DamagePoison, false},
		{"monster", 0x190008 | object.ClassMonster, object.DamagePoison, false},
		{"weapon", 0x190008 | object.ClassWeapon, object.DamagePoison, false},
		{"wand", 0x190008 | object.ClassWand, object.DamagePoison, false},
		{"wrong-type", 0x190008, object.DamageFlame, false},
	} {
		t.Run("shape/"+tc.name, func(t *testing.T) {
			if got := playerDamageCloudPoisonShape4E17B0(&Object{ObjClass: tc.class}, tc.typ); got != tc.want {
				t.Fatalf("shape=%t want=%t", got, tc.want)
			}
		})
	}
	if playerDamageCloudPoisonShape4E17B0(nil, object.DamagePoison) {
		t.Fatal("weapon-less poison admitted as cloud")
	}
}

func TestPlayerDamageNPCCloudPoison4E17B0MissingServicesKeepPrefix(t *testing.T) {
	for _, service := range []string{"excluded", "direction", "quest-scale", "default"} {
		for _, owner := range []string{"self", "player"} {
			t.Run(service+"/"+owner, func(t *testing.T) {
				target, source, cloud := defaultDamageCloudFixture4E0B30(t, owner, 0x11012)
				update, reports := target.UpdateDataMonster(), 0
				r := PlayerDamageRuntime4E17B0{
					BlockSourceExcluded: func(*Object) bool { return service != "direction" },
					QuestMode:           func() bool { return service == "quest-scale" },
					DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("missing-service default")
						return false
					},
					Unsupported: func(why string, got, attacker, weapon *Object, damage int32, typ object.DamageType) {
						if why == "" || got != target || attacker != source || weapon != cloud || damage != 19 || typ != object.DamagePoison {
							t.Fatal("unsupported context")
						}
						reports++
					},
				}
				if service == "excluded" {
					r.BlockSourceExcluded = nil
				}
				if service == "default" {
					r.DefaultDamage = nil
				}
				h, result := PlayerDamageNative4E17B0(target, source, cloud, 19, object.DamagePoison, r)
				marker, markerType := uint32(0), uint32(91)
				if service != "excluded" && owner != "self" {
					marker, markerType = 1, uint32(cloud.TypeInd)
				}
				if service != "excluded" && service != "direction" && owner == "self" {
					marker, markerType = 2, 5
				}
				if h || result || reports != 1 || update.Field547 != marker || update.Field546 != markerType || update.Field1 != math.Float32bits(0.25) || target.HealthData.Cur != 20 {
					t.Fatalf("service=%s result=%t/%t reports=%d marker=%d/%d", service, h, result, reports, update.Field547, update.Field546)
				}
			})
		}
	}
}

func TestPlayerDamageNPCCloudPoison4E17B0MarkerAtUse(t *testing.T) {
	for _, stage := range []string{"excluded", "direction", "god", "quest", "scale"} {
		t.Run(stage, func(t *testing.T) {
			target, source, cloud := defaultDamageCloudFixture4E0B30(t, "self", 0x11012)
			update := target.UpdateDataMonster()
			mark := func(at string) {
				if at == stage {
					update.Field547, update.Field546 = 7, 8
				}
			}
			r := PlayerDamageRuntime4E17B0{
				BlockSourceExcluded: func(*Object) bool { mark("excluded"); return false },
				BlockDirection:      func(*Object, types.Pointf) bool { mark("direction"); return false },
				GodMode:             func() bool { mark("god"); return false },
				QuestMode:           func() bool { mark("quest"); return true },
				QuestDamageScale:    func() float32 { mark("scale"); return 1 },
				DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					if update.Field547 != 7 || update.Field546 != 8 {
						t.Fatal("marker was reselected or overwritten")
					}
					return true
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			h, result := PlayerDamageNative4E17B0(target, source, cloud, 19, object.DamagePoison, r)
			if !h || !result || update.Field1 != math.Float32bits(0.25) {
				t.Fatal("marker tail")
			}
		})
	}
}

func TestPlayerDamageNPCCloudPoison4E17B0CustomShield(t *testing.T) {
	for _, raw := range []int32{-7, 0, 19} {
		for _, variant := range []string{"ordinary", "broken", "no-item", "late-missile", "reflection-clears-class", "reflection-sets-subclass"} {
			t.Run(fmt.Sprintf("raw-%d/%s", raw, variant), func(t *testing.T) {
				target, source, cloud := defaultDamageCloudFixture4E0B30(t, "player", 0x11012)
				cached := target.UpdateDataMonster()
				cached.ArmorEquipFlags = 0x3000000
				live, freeLive := alloc.New(MonsterUpdateData{})
				t.Cleanup(freeLive)
				live.AIStack[0].Action = 21
				shield, freeShield := alloc.New(Object{})
				t.Cleanup(freeShield)
				*shield = Object{ObjFlags: object.FlagEquipped, ObjSubClass: 2}
				if variant != "no-item" {
					target.InvFirstItem = shield
				}
				var events []string
				r := PlayerDamageRuntime4E17B0{
					BlockSourceExcluded: func(*Object) bool {
						events = append(events, "excluded")
						target.UpdateData = unsafe.Pointer(live)
						return false
					},
					BlockDirection: func(*Object, types.Pointf) bool { events = append(events, "direction"); return true },
					Audio: func(sound int, got *Object) {
						if sound != 878 || got != target || cached.Field547 != 1 {
							t.Fatal("custom block audio")
						}
						events = append(events, "audio")
						if variant == "late-missile" || variant == "reflection-clears-class" || variant == "reflection-sets-subclass" {
							cloud.ObjClass = object.ClassMissile
						}
					},
					ProjectileReflect: func(got, unit *Object) {
						if got != cloud || unit != target {
							t.Fatal("reflect args")
						}
						events = append(events, "reflect")
						if variant == "reflection-clears-class" {
							cloud.ObjClass = object.ClassSimple
						}
						if variant == "reflection-sets-subclass" {
							cloud.ObjSubClass = 2
						}
					},
					ClearOwner: func(got *Object) {
						if got != cloud {
							t.Fatal("clear owner")
						}
						events = append(events, "clear-owner")
					},
					SetOwner: func(got, attack *Object) {
						if got != target || attack != cloud {
							t.Fatal("set owner")
						}
						events = append(events, "set-owner")
					},
					BlockDamagePercent: func() float64 { events = append(events, "percent"); return 0.25 },
					CanDamageBlockItem: func(got *Object) bool {
						if got != shield {
							t.Fatal("shield preflight")
						}
						events = append(events, "ready")
						return true
					},
					DamageBlockItem: func(got, unit, attacker, attack *Object, amount float32, typ object.DamageType) bool {
						wantItem := shield
						if variant == "no-item" {
							wantItem = nil
						}
						if got != wantItem || unit != target || attacker != source || attack != cloud || amount != float32(float64(raw)*0.25) || typ != object.DamagePoison {
							t.Fatal("block amount/type (2 is the shield mask, not damage type)")
						}
						events = append(events, "wear")
						if variant == "broken" {
							shield.ObjFlags |= object.FlagDestroyed
						}
						return true
					},
					Melee: PlayerDamageMeleeRuntime4E17B0{MonsterPopBlockAction: func(got *Object) {
						if got != target {
							t.Fatal("broken block action")
						}
						events = append(events, "pop")
					}},
					QuestMode: func() bool { t.Fatal("blocked cloud reached quest tail"); return false },
					DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
						t.Fatal("blocked cloud reached HP tail")
						return false
					},
					Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
				}
				h, result := PlayerDamageNative4E17B0(target, source, cloud, raw, object.DamagePoison, r)
				wantEvents := []string{"excluded", "direction", "audio"}
				if variant == "late-missile" || variant == "reflection-clears-class" || variant == "reflection-sets-subclass" {
					wantEvents = append(wantEvents, "reflect")
				}
				if variant == "late-missile" {
					wantEvents = append(wantEvents, "clear-owner", "set-owner")
				}
				wantEvents = append(wantEvents, "percent")
				if variant != "no-item" {
					wantEvents = append(wantEvents, "ready")
				}
				wantEvents = append(wantEvents, "wear")
				if variant == "broken" {
					wantEvents = append(wantEvents, "pop")
				}
				if !h || result || !slices.Equal(events, wantEvents) || cached.Field547 != 1 || cached.Field546 != uint32(1221) || cached.Field1 != math.Float32bits(0.25) || target.HealthData.Cur != 20 || live.Field547 != 0 {
					t.Fatalf("custom shield result=%t/%t events=%v want=%v", h, result, events, wantEvents)
				}
			})
		}
	}
}
