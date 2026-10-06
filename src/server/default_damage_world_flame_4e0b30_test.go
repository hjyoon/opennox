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

// Stock Flame and all red FlameCleanse sizes are FIRE|DANGEROUS|SIMPLE,
// not MISSILE. Keep every pointer C-owned and above the PE32 address range.
func worldFlameFixture4E17B0(t *testing.T, owner, victim string) (target, source, flame *Object) {
	t.Helper()
	target, source, flame = defaultDamageCloudFixture4E0B30(t, owner, 2)
	flame.ObjClass = object.ClassFire | object.ClassDangerous | object.ClassSimple | object.ClassVisibleEnable | object.ClassLight
	flame.TypeInd = 199
	target.Buffs = 1<<0 | 1<<22
	switch victim {
	case "monster":
	case "NPC":
		target.ObjSubClass = 0x11012
	case "player":
		update, freeUpdate := alloc.New(PlayerUpdateData{})
		player, freePlayer := alloc.New(Player{})
		t.Cleanup(freeUpdate)
		t.Cleanup(freePlayer)
		*player = Player{}
		*update = PlayerUpdateData{Player: player, State: PlayerState13, Field76: 99, Field75: 91, Field21: math.Float32bits(0.25)}
		target.ObjClass, target.ObjSubClass, target.UpdateData = object.ClassPlayer, 0, unsafe.Pointer(update)
	default:
		t.Fatalf("unknown flame victim %q", victim)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), target.UpdateData, unsafe.Pointer(target.HealthData), unsafe.Pointer(flame)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("world FLAME pointer %p must be above 4 GiB", ptr)
			}
		}
	}
	return
}

func TestDefaultDamageWorldFlame4E0B30SignedTail(t *testing.T) {
	for _, victim := range []string{"monster", "NPC", "player"} {
		for _, owner := range []string{"nil", "self", "imaginary", "player", "NPC", "proxy-NPC"} {
			for _, raw := range []int32{-3, 0, 1, 9, 25} {
				for _, protection := range []float64{0, 0.5, 1} {
					t.Run(fmt.Sprintf("%s/%s/raw-%d/protection-%g", victim, owner, raw, protection), func(t *testing.T) {
						v, a, w := worldFlameFixture4E17B0(t, owner, victim)
						wantDamage := int32(math.RoundToEven(float64(float32((1 - float64(float32(protection))) * float64(raw)))))
						if wantDamage == 0 {
							wantDamage = 1
						}
						wantPosition := w.PrevPos
						if a == nil {
							wantPosition = types.Pointf{}
						}
						var events []string
						r := DefaultDamageWorldRuntime4E0B30{
							Frame:         func() uint32 { return 1400 },
							GameplayFlag1: func() bool { return true },
							IsEnemy:       func(*Object, *Object) bool { events = append(events, "enemy"); return false },
							FireProtection: func(got *Object) float64 {
								if got != v {
									t.Fatal("protection identity")
								}
								events = append(events, "fire")
								return protection
							},
							Audio: func(id int, got *Object) {
								if id != 104 || got != v {
									t.Fatal("fire voice")
								}
								events = append(events, "voice")
							},
							BuffOff: func(got *Object, enchant EnchantID) {
								if got != v || enchant != 0 || v.Pos132 != wantPosition {
									t.Fatal("visibility order")
								}
								events = append(events, "buff")
								v.Buffs &^= 1 << enchant
							},
							MonsterHasHitSound: func(*Object) bool { events = append(events, "lookup"); return false },
							DefaultDamageSound: func(got, attack *Object) {
								if got != v || attack != w || v.Obj130 != w || v.Field131 != 1 || v.Frame134 != 1400 {
									t.Fatal("sound attribution")
								}
								events = append(events, "sound")
							},
							PlayerSetState: func(got *Object, state PlayerState) bool {
								if victim != "player" || got != v || state != PlayerState30 {
									t.Fatal("hurt-state identity")
								}
								events = append(events, "hurt")
								return true
							},
							DamageClear: func(got *Object, damage int32) {
								if got != v || damage != wantDamage {
									t.Fatalf("HP amount=%d want=%d", damage, wantDamage)
								}
								events = append(events, "hp")
								v.HealthData.Cur = uint16(max(20-damage, 0))
							},
							CallDamage: func(*Object, *Object, *Object, int32, object.DamageType) bool {
								t.Fatal("world fire retaliated Shock")
								return false
							},
							ElectricProtection: func(*Object) float64 { t.Fatal("world fire used electric protection"); return 0 },
							Unsupported:        func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
						}
						if !DefaultDamageWorld4E0B30(v, a, w, raw, object.DamageFlame, r) {
							t.Fatal("FLAME returned zero")
						}
						var want []string
						if a != nil {
							want = append(want, "enemy")
						}
						want = append(want, "fire")
						if protection != 0 {
							want = append(want, "voice")
						}
						if a != nil {
							want = append(want, "buff")
						}
						if owner == "NPC" {
							want = append(want, "lookup")
						}
						want = append(want, "sound")
						if victim == "player" && wantDamage >= 20 {
							want = append(want, "hurt")
						}
						if owner == "NPC" || owner == "proxy-NPC" {
							want = append(want, "enemy")
						}
						want = append(want, "hp")
						if !slices.Equal(events, want) || v.HealthData.Cur != uint16(max(20-wantDamage, 0)) || v.Pos132 != wantPosition || !v.HasEnchant(22) || v.HasEnchant(0) != (a == nil) {
							t.Fatalf("events=%v want=%v HP=%d", events, want, v.HealthData.Cur)
						}
						if victim != "player" {
							ud := v.UpdateDataMonster()
							marker, markerType := uint32(1), uint32(w.TypeInd)
							if a == nil || a == w {
								marker, markerType = 2, 1
							}
							if ud.Field547 != marker || ud.Field546 != markerType || !ud.StatusFlags.Has(object.MonStatusOnFire|object.MonStatusInjured) || ud.Field1 != math.Float32bits(0.25) {
								t.Fatal("fire marker/status/carry")
							}
						} else if v.UpdateDataPlayer().Field21 != math.Float32bits(0.25) {
							t.Fatal("FLAME changed player carry")
						}
					})
				}
			}
		}
	}
}

func TestDefaultDamageWorldFlame4E0B30EarlyOrder(t *testing.T) {
	for _, gate := range []string{"invulnerable", "dead", "campaign-friendly", "no-update", "immune", "live-immune"} {
		t.Run(gate, func(t *testing.T) {
			v, a, w := worldFlameFixture4E17B0(t, "player", "monster")
			switch gate {
			case "invulnerable":
				v.Buffs |= 1 << 23
			case "dead":
				v.ObjFlags |= object.FlagDead
			case "no-update":
				v.ObjFlags |= object.FlagNoUpdate
			case "immune":
				v.ObjSubClass |= 0x400
				v.HealthData = nil
			}
			queries := 0
			r := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 1401 },
				GameplayFlag1: func() bool { return gate != "campaign-friendly" },
				IsEnemy: func(*Object, *Object) bool {
					queries++
					if gate == "live-immune" {
						v.ObjSubClass |= 0x400
						v.HealthData = nil
					}
					return false
				},
				FireProtection: func(*Object) float64 { t.Fatal("early protection"); return 0 },
				DamageClear:    func(*Object, int32) { t.Fatal("early HP") },
				Unsupported:    func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
			}
			if !DefaultDamageWorld4E0B30(v, a, w, 3, object.DamageFlame, r) {
				t.Fatal("early result")
			}
			wantQueries := 1
			if gate == "invulnerable" || gate == "dead" {
				wantQueries = 0
			}
			ud := v.UpdateDataMonster()
			wantMarker := uint32(0)
			if gate == "invulnerable" {
				wantMarker = 99
			}
			if queries != wantQueries || ud.Field547 != wantMarker || ud.Field546 != 91 || v.Frame134 != 77 || v.Pos132 != types.Ptf(31, 47) || ud.StatusFlags.Has(object.MonStatusOnFire) != (gate != "invulnerable") {
				t.Fatalf("queries=%d marker=%d status=%x", queries, ud.Field547, ud.StatusFlags)
			}
		})
	}
}

func TestDefaultDamageWorldFlame4E0B30Shield(t *testing.T) {
	v, a, w := worldFlameFixture4E17B0(t, "NPC", "player")
	v.Buffs |= 1 << 26
	var events []string
	r := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true }, IsEnemy: func(*Object, *Object) bool { return false },
		FireProtection:     func(*Object) float64 { events = append(events, "fire"); return .5 },
		BuffOff:            func(*Object, EnchantID) { events = append(events, "buff") },
		MonsterHasHitSound: func(*Object) bool { return false },
		ShieldReduce: func(got *Object, amount *int32, typ object.DamageType, attack *Object) {
			if got != v || attack != w || *amount != 4 || typ != object.DamageFlame {
				t.Fatal("Shield args/order")
			}
			events = append(events, "shield")
			*amount = 0
		},
		DamageClear: func(*Object, int32) { t.Fatal("depleted Shield reached HP") },
		Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(why) },
	}
	if DefaultDamageWorld4E0B30(v, a, w, 9, object.DamageFlame, r) || !slices.Equal(events, []string{"fire", "buff", "shield"}) || v.HealthData.Cur != 20 || v.Obj130 != w || v.Field131 != 1 {
		t.Fatalf("Shield events=%v", events)
	}
}
