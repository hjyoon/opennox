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

func TestPlayerDamagePlayerCloudPoison4E17B0SignedQuest(t *testing.T) {
	for _, owner := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
		for _, raw := range []int32{-7, 0, 3, 10, 19} {
			for _, scale := range []float32{-1, 0, 0.5, 1.25} {
				t.Run(fmt.Sprintf("%s/raw-%d/scale-%g", owner, raw, scale), func(t *testing.T) {
					v, a, w := worldFlameFixture4E17B0(t, owner, "player")
					w.ObjClass, w.TypeInd = 0x190008, 1221
					ud := v.UpdateDataPlayer()
					ud.Field57, ud.Player.ArmorEquip, ud.Player.WeaponEquip, ud.State = math.Float32bits(0.9), 0x3000000, 0x400, PlayerState16
					v.Buffs |= 1<<27 | 1<<26
					before := *ud
					marker, kind := uint32(2), uint32(5)
					if a != nil && a != w {
						marker, kind = 1, 1222
					}
					var events []string
					r := PlayerDamageRuntime4E17B0{
						Frame: func() uint32 { t.Fatal("non-invulnerable cloud read frame"); return 0 },
						BlockSourceExcluded: func(got *Object) bool {
							if got != w || ud.Field76 != 0 || ud.Field75 != 91 {
								t.Fatal("cached clear/exclusion order")
							}
							w.TypeInd = 1222
							events = append(events, "excluded")
							return true
						},
						BlockDirection:   func(*Object, types.Pointf) bool { t.Fatal("excluded cloud checked facing"); return true },
						Audio:            func(int, *Object) { t.Fatal("nonmissile excluded cloud entered defense") },
						ItemArmorValue:   func(*Object) float32 { t.Fatal("poison queried armor"); return 0 },
						CanDamageArmor:   func(*Object) bool { t.Fatal("poison checked armor wear"); return true },
						ShieldReduce:     func(*Object, *int32, object.DamageType, *Object) { t.Fatal("Shield reduced poison") },
						GodMode:          func() bool { events = append(events, "god"); return false },
						QuestMode:        func() bool { events = append(events, "quest"); return scale >= 0 },
						QuestDamageScale: func() float32 { events = append(events, "scale"); return scale },
						DefaultDamage: func(got, src, cloud *Object, d int32, typ object.DamageType) bool {
							want := raw
							if scale >= 0 {
								want = int32(math.RoundToEven(float64(float32(float64(raw) * float64(scale)))))
								if raw > 0 && want < 1 {
									want = 1
								}
							}
							if got != v || src != a || cloud != w || d != want || typ != object.DamagePoison || ud.Field76 != marker || ud.Field75 != kind {
								t.Fatal("raw case-5 tail")
							}
							events = append(events, "default")
							return true
						},
						Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
					}
					h, ok := PlayerDamageNative4E17B0(v, a, w, raw, object.DamagePoison, r)
					want := []string{}
					if a != nil {
						want = append(want, "excluded")
					}
					want = append(want, "god", "quest")
					if scale >= 0 {
						want = append(want, "scale")
					}
					want = append(want, "default")
					before.Field76, before.Field75 = marker, kind
					if !h || !ok || !slices.Equal(events, want) || *ud != before || v.HealthData.Cur != 20 {
						t.Fatalf("prefix=%t/%t events=%v want=%v", h, ok, events, want)
					}
				})
			}
		}
	}
}

func TestPlayerDamagePlayerCloudPoison4E17B0CachedObserve(t *testing.T) {
	v, a, w := worldFlameFixture4E17B0(t, "NPC", "player")
	w.ObjClass, w.TypeInd = 0x190008, 1221
	old := v.UpdateDataPlayer()
	old.Player.Field3680, old.Player.CameraFollowObj = 2, a
	live, free := alloc.New(PlayerUpdateData{})
	t.Cleanup(free)
	*live = PlayerUpdateData{Player: old.Player, State: PlayerState16, Field76: 67, Field75: 68, Field21: math.Float32bits(-0.5)}
	before, pos := *live, w.PrevPos
	var events []string
	r := PlayerDamageRuntime4E17B0{
		ObserveClear: func(got *Object) {
			if got != v || old.Field76 != 0 || old.Field75 != 91 {
				t.Fatal("observe preceded clear")
			}
			v.UpdateData = unsafe.Pointer(live)
			events = append(events, "observe")
		},
		BlockSourceExcluded: func(*Object) bool {
			w.PrevPos = types.Ptf(300, 400)
			w.TypeInd = 1222
			events = append(events, "excluded")
			return false
		},
		BlockDirection: func(got *Object, p types.Pointf) bool {
			if got != v || p != pos || old.Field76 != 1 || old.Field75 != 1222 {
				t.Fatal("cached position/marker")
			}
			events = append(events, "direction")
			return true
		},
		DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
			events = append(events, "default")
			return false
		},
		Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
	}
	h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamagePoison, r)
	if !h || ok || !slices.Equal(events, []string{"observe", "excluded", "direction", "default"}) || *live != before || old.Field76 != 1 || old.Field75 != 1222 {
		t.Fatalf("cached observe %t/%t events=%v", h, ok, events)
	}
}

func TestPlayerDamagePlayerCloudPoison4E17B0RealDefault(t *testing.T) {
	for _, owner := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
		t.Run(owner, func(t *testing.T) {
			v, a, w := worldFlameFixture4E17B0(t, owner, "player")
			w.ObjClass, w.TypeInd = 0x190008, 1221
			ud := v.UpdateDataPlayer()
			ud.Field57 = math.Float32bits(0.9)
			r := damageMeleeRuntimeFixture4E17B0(t)
			r.BlockSourceExcluded = func(*Object) bool { return true }
			h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamagePoison, r)
			marker, kind := uint32(2), uint32(5)
			if a != nil && a != w {
				marker, kind = 1, 1221
			}
			if !h || !ok || v.HealthData.Cur != 11 || v.Obj130 != w || v.Field131 != 5 || v.Frame134 != 1400 || ud.Field76 != marker || ud.Field75 != kind || ud.Field21 != math.Float32bits(0.25) {
				t.Fatalf("real cloud HP %t/%t HP=%d marker=%d/%d", h, ok, v.HealthData.Cur, ud.Field76, ud.Field75)
			}
		})
	}
}

func TestPlayerDamagePlayerCloudPoison4E17B0Guards(t *testing.T) {
	for _, gate := range []string{"no-update", "dead", "invulnerable", "disabled-player", "coop-self", "god"} {
		t.Run(gate, func(t *testing.T) {
			v, a, w := worldFlameFixture4E17B0(t, "self", "player")
			w.ObjClass, w.TypeInd = 0x190008, 1221
			ud := v.UpdateDataPlayer()
			switch gate {
			case "no-update":
				v.ObjFlags |= object.FlagNoUpdate
			case "dead":
				v.ObjFlags |= object.FlagDead
			case "invulnerable":
				v.Buffs |= 1 << 23
			case "disabled-player":
				ud.Player.Field3680 |= 1
			case "coop-self":
				w.ObjOwner = v
			}
			before := *ud
			r := PlayerDamageRuntime4E17B0{
				Frame: func() uint32 {
					if gate != "invulnerable" {
						t.Fatal("unexpected frame read")
					}
					return 1401
				},
				CoopMode: func() bool { return gate == "coop-self" },
				BlockSourceExcluded: func(*Object) bool {
					if gate != "god" {
						t.Fatal("early gate reached exclusion")
					}
					return true
				},
				GodMode: func() bool {
					if gate != "god" || ud.Field76 != 2 || ud.Field75 != 5 {
						t.Fatal("GodMode marker order")
					}
					return true
				},
				QuestMode: func() bool { t.Fatal("early gate reached Quest"); return true },
				DefaultDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
					t.Fatal("early gate reached HP")
					return true
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamagePoison, r)
			if gate == "god" {
				before.Field76, before.Field75 = 2, 5
			}
			if !h || ok != (gate == "god" || gate == "invulnerable") || *ud != before || v.HealthData.Cur != 20 {
				t.Fatalf("guard %t/%t marker=%d/%d", h, ok, ud.Field76, ud.Field75)
			}
		})
	}
}

func TestPlayerDamagePlayerCloudPoison4E17B0CustomShield(t *testing.T) {
	v, a, w := worldFlameFixture4E17B0(t, "NPC", "player")
	w.ObjClass, w.TypeInd = 0x190008, 1221
	ud := v.UpdateDataPlayer()
	ud.State, ud.Player.ArmorEquip = PlayerState16, 0x3000000
	var events []string
	r := PlayerDamageRuntime4E17B0{
		BlockSourceExcluded: func(*Object) bool { events = append(events, "excluded"); return false },
		BlockDirection: func(*Object, types.Pointf) bool {
			if ud.Field76 != 1 || ud.Field75 != 1221 {
				t.Fatal("direction marker")
			}
			events = append(events, "direction")
			return true
		},
		Audio: func(id int, got *Object) {
			if id != 878 || got != v {
				t.Fatal("shield audio")
			}
			events = append(events, "audio")
		},
		BlockDamagePercent: func() float64 { events = append(events, "balance"); return 0.25 },
		DamageBlockItem: func(item, owner, src, cloud *Object, amount float32, typ object.DamageType) bool {
			if item != nil || owner != v || src != a || cloud != w || amount != 2.25 || typ != object.DamagePoison {
				t.Fatal("live shield wear")
			}
			events = append(events, "wear")
			return true
		},
		GodMode:     func() bool { t.Fatal("successful block queried GodMode"); return true },
		Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
	}
	h, ok := PlayerDamageNative4E17B0(v, a, w, 9, object.DamagePoison, r)
	if !h || ok || v.HealthData.Cur != 20 || !slices.Equal(events, []string{"excluded", "direction", "audio", "balance", "wear"}) {
		t.Fatalf("block %t/%t events=%v", h, ok, events)
	}
}
