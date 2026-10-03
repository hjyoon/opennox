package legacy

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Direct Go PlayerDamage with C-owned >4GiB records and the real global Quest
// flag. The controlled armor service calls real ArmorDamage/UnitDamageClear,
// then changes the flag; scale/definition values are fixture inputs. The NPC
// DefaultDamage runtime and both HP stores are production services. This is
// not an unmodified C-dispatcher or natural stock-map flag-transition test.
func TestPlayerDamageNPCMissileLiveNative4E17B0CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	if !srv.Objs.Init(16) {
		t.Fatal("native allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, typ := range []object.DamageType{object.DamageImpale, object.DamageImpact} {
		for _, entering := range []bool{false, true} {
			t.Run(fmt.Sprintf("type-%d/entering-%t", typ, entering), func(t *testing.T) {
				noxflags.UnsetGame(noxflags.GameModeQuest)
				if !entering {
					noxflags.SetGame(noxflags.GameModeQuest)
				}
				target, source := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
				missile, armor := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
				update, freeUpdate := alloc.New(server.MonsterUpdateData{})
				playerUpdate, freePlayerUpdate := alloc.New(server.PlayerUpdateData{})
				player, freePlayer := alloc.New(server.Player{})
				health, freeHealth := alloc.New(server.HealthData{})
				armorHealth, freeArmorHealth := alloc.New(server.HealthData{})
				armorUpdate, freeArmorUpdate := alloc.New(server.WeaponArmorUpdateData{})
				armorInit, freeArmorInit := alloc.New(server.ModifierInitData{})
				missileInit, freeMissileInit := alloc.New(server.ModifierInitData{})
				for _, free := range []func(){freeUpdate, freePlayerUpdate, freePlayer, freeHealth, freeArmorHealth, freeArmorUpdate, freeArmorInit, freeMissileInit} {
					t.Cleanup(free)
				}
				*update = server.MonsterUpdateData{Field518: math.Float32bits(0.25), Field1: math.Float32bits(0.25), Field547: 99, Field546: 77}
				*player = server.Player{PlayerInd: 7}
				*playerUpdate = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
				*health = server.HealthData{Cur: 60, Max: 60, Field2: 60}
				*armorHealth = server.HealthData{Cur: 25, Max: 25, Field2: 25}
				*armorUpdate = server.WeaponArmorUpdateData{Field0: math.Float32bits(0.25)}
				target.TypeInd, target.ObjClass, target.ObjSubClass, target.ObjFlags, target.Material = 71, object.ClassMonster, 0x11012, 0, 0x4000
				target.UpdateData, target.HealthData, target.InvFirstItem = unsafe.Pointer(update), health, armor
				source.TypeInd, source.ObjClass, source.ObjFlags = 72, object.ClassPlayer|object.ClassComplex|object.ClassLight, 0
				source.UpdateData, source.PrevPos = unsafe.Pointer(playerUpdate), types.Ptf(44, 9)
				missile.TypeInd, missile.ObjClass, missile.ObjSubClass, missile.ObjFlags = 66, object.ClassMissile|object.ClassWeapon|object.ClassComplex|object.ClassNotStackable, 0x10, 0
				missile.InitData, missile.ObjOwner, missile.PrevPos = unsafe.Pointer(missileInit), source, types.Ptf(20, 7)
				armor.ObjClass, armor.ObjFlags, armor.InvHolder = object.ClassArmor, object.FlagEquipped, target
				armor.UpdateData, armor.InitData, armor.HealthData = unsafe.Pointer(armorUpdate), unsafe.Pointer(armorInit), armorHealth
				// Callback identity is a C test probe; only the explicitly
				// supplied native service below is called, never the probe.
				armor.Damage = objectDamageNativeProbePtr()
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(missile), unsafe.Pointer(armor),
					unsafe.Pointer(update), unsafe.Pointer(playerUpdate), unsafe.Pointer(player), unsafe.Pointer(health), unsafe.Pointer(armorHealth),
					unsafe.Pointer(armorUpdate), unsafe.Pointer(armorInit), unsafe.Pointer(missileInit)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
						t.Fatalf("native pointer=%p, want >4 GiB", ptr)
					}
				}
				beforeSource, beforePlayerUpdate, beforePlayer, beforeMissile := *source, *playerUpdate, *player, *missile
				tail := defaultDamageWorldRuntime4E0B30(srv)
				var events []string
				r := server.PlayerDamageRuntime4E17B0{
					Frame: srv.Frame,
					QuestMode: func() bool {
						events = append(events, "quest-mode")
						return noxflags.HasGame(noxflags.GameModeQuest)
					},
					QuestDamageScale: func() float32 { events = append(events, "quest-scale"); return 0.5 },
					ItemArmorValue: func(v *server.Object) float32 {
						if v != armor {
							t.Fatal("armor definition identity")
						}
						return 0.25
					},
					CanDamageArmor: func(v *server.Object) bool { return v == armor && canEquipDamageNative4E16D0(v) },
					DamageArmor: func(v, a, w *server.Object, d int32, gotType object.DamageType) bool {
						if v != armor || a != source || w != missile || d != 2 || gotType != typ ||
							update.Field547 != 1 || update.Field546 != uint32(missile.TypeInd) || update.Field1 != math.Float32bits(0.25) {
							t.Fatal("C-owned wear identity/order")
						}
						events = append(events, "armor")
						result := server.ArmorDamage4E1500(v, a, w, d, gotType, func(item, _, _ *server.Object, amount int32, _ object.DamageType) bool {
							unitDamageClearCall4EE5E0(item, amount)
							return true
						})
						noxflags.UnsetGame(noxflags.GameModeQuest)
						if entering {
							noxflags.SetGame(noxflags.GameModeQuest)
						}
						return result
					},
					ReportArmorHealth: func(*server.Object, *server.Object, uint16, uint16) { t.Fatal("NPC reported player health") },
					DefaultDamage: func(v, a, w *server.Object, d int32, gotType object.DamageType) bool {
						want := int32(6)
						if entering {
							want = 3
						}
						if v != target || a != source || w != missile || d != want || gotType != typ {
							t.Fatalf("C-owned live Quest damage=%d, want %d", d, want)
						}
						events = append(events, "default")
						return server.DefaultDamageWorld4E0B30(v, a, w, d, gotType, tail)
					},
					Unsupported: func(reason string, _, _, _ *server.Object, _ int32, _ object.DamageType) {
						t.Fatalf("native missile: %s", reason)
					},
				}
				wantHP, wantEvents := uint16(54), []string{"armor", "quest-mode"}
				if entering {
					wantHP = 57
					wantEvents = append(wantEvents, "quest-scale")
				}
				wantEvents = append(wantEvents, "default")
				if h, result := server.PlayerDamageNative4E17B0(target, source, missile, 8, typ, r); !h || !result || health.Cur != wantHP ||
					armorHealth.Cur != 23 || armorUpdate.Field0 != math.Float32bits(0.25) || update.Field1 != math.Float32bits(0.25) ||
					update.Field547 != 1 || update.Field546 != uint32(missile.TypeInd) || !update.StatusFlags.Has(object.MonStatusInjured) ||
					target.Obj130 != missile || target.Field131 != uint32(typ) || target.Frame134 != srv.Frame() || target.Pos132 != source.PrevPos ||
					*source != beforeSource || *playerUpdate != beforePlayerUpdate || *player != beforePlayer || *missile != beforeMissile ||
					!slices.Equal(events, wantEvents) {
					t.Fatalf("native live Quest=%t/%t HP=%d armor=%d carry=%g events=%v, want %v", h, result, health.Cur, armorHealth.Cur, math.Float32frombits(update.Field1), events, wantEvents)
				}
				t.Logf("C-owned >4GiB NPC missile: target=%p source=%p missile=%p armor=%p type=%d entering=%t HP=60->%d armor=25->23", target, source, missile, armor, typ, entering, health.Cur)
			})
		}
	}
}
