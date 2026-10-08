package legacy

import (
	"fmt"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackBowShoot2_539D80CachedAmmoLivePlayerAndMotion(t *testing.T) {
	for _, flags := range []uint32{4, 8, 12} {
		for _, trace := range []bool{false, true} {
			for _, allocate := range []bool{false, true} {
				for _, infinite := range []uint8{0, 1, 2} {
					t.Run(fmt.Sprintf("flags-%d/trace-%t/allocate-%t/infinite-%d", flags, trace, allocate, infinite), func(t *testing.T) {
						owner := &server.Object{ObjClass: object.ClassPlayer, PosVec: types.Ptf(100, 200), Direction1: server.Dir16(0xfffd)}
						owner.Shape.Circle.R = 6
						old := &server.PlayerUpdateData{Player: &server.Player{PlayerInd: 5}}
						live := &server.PlayerUpdateData{Player: &server.Player{PlayerInd: 23}}
						owner.UpdateData = unsafe.Pointer(old)
						ammo := &server.AmmoUseData{Charge0: 3, Charge1: 1, Field2: infinite}
						replacement := &server.AmmoUseData{Charge0: 9, Charge1: 77}
						attrs := &server.ModifierInitData{}
						quiver := &server.Object{UseData: server.UseDataPtr{Ptr: unsafe.Pointer(ammo)}, InitData: unsafe.Pointer(attrs)}
						weapon := &server.Object{InitData: unsafe.Pointer(&server.ModifierInitData{})}
						projectile := &server.Object{SpeedCur: 8, CollideData: unsafe.Pointer(&server.ArrowCollideData{}), InitData: unsafe.Pointer(&server.ModifierInitData{})}
						var events []string
						directionReads := 0
						h := playerAttackBowRuntime538960{playerAttackProjectileRuntime538960: playerAttackProjectileRuntime538960{
							Trace: func(from, to types.Pointf, f server.MapTraceFlags) bool {
								if from != types.Ptf(100, 200) || to != types.Ptf(107.5, 197.5) || f != 5 {
									t.Fatal("spawn/trace order")
								}
								quiver.UseData.Ptr, owner.UpdateData = unsafe.Pointer(replacement), unsafe.Pointer(live)
								events = append(events, "trace")
								return trace
							},
							NewObject: func(name string) *server.Object {
								want := "ArcherBolt"
								if flags == 4 {
									want = "ArcherArrow"
								}
								if name != want {
									t.Fatalf("projectile %q, want %q", name, want)
								}
								events = append(events, "allocate")
								if !allocate {
									return nil
								}
								return projectile
							},
							CreateAt: func(v, u *server.Object, pos types.Pointf) {
								if v != projectile || u != owner || pos != types.Ptf(107.5, 197.5) || (*server.ArrowCollideData)(v.CollideData).Owner != owner {
									t.Fatal("owner must be written before creation")
								}
								owner.Direction1 = 300
								events = append(events, "create")
							},
							ApplyModifierAttrs: func(v *server.Object, got *server.ModifierInitData) {
								if v != projectile || got != attrs {
									t.Fatal("quiver attributes")
								}
								events = append(events, "attributes")
							},
							DelayedDelete: func(v *server.Object) {
								if v != quiver {
									t.Fatal("deleted wrong quiver")
								}
								events = append(events, "delete")
							},
							AudioEvent: func(id sound.ID, v *server.Object) {
								want := sound.ID(886)
								if flags == 4 {
									want = 885
								}
								if id != want || v != owner {
									t.Fatal("shot sound")
								}
								events = append(events, "audio")
							},
						}, Direction: func(dir server.Dir16) types.Pointf {
							directionReads++
							switch directionReads {
							case 1:
								if int16(dir) != -3 {
									t.Fatal("direction was narrowed to byte")
								}
								return types.Ptf(0.75, -0.25)
							case 2:
								if dir != 300 {
									t.Fatal("X direction was cached before creation")
								}
								owner.Direction1 = 301
								return types.Ptf(2, 9)
							case 3:
								if dir != 301 {
									t.Fatal("Y direction was cached before X callback")
								}
								return types.Ptf(9, 3)
							default:
								t.Fatal("extra direction read")
								return types.Pointf{}
							}
						}, ReportCharges: func(index uint8, v *server.Object, charge, delay uint8) {
							if index != 23 || v != quiver || charge != 0 || delay != 3 || ammo.Charge1 != 0 {
								t.Fatal("cached ammo/live player report")
							}
							events = append(events, "report")
						}}
						playerAttackBowShoot2_539D80(owner, quiver, weapon, flags, h)
						want := []string{"trace"}
						if trace {
							want = append(want, "allocate")
							if allocate {
								want = append(want, "create", "attributes")
								if projectile.VelVec != types.Ptf(16, 24) || projectile.Direction1 != 301 || projectile.Direction2 != 301 || directionReads != 3 {
									t.Fatal("live projectile motion")
								}
							}
							if infinite == 0 {
								want = append(want, "report", "delete")
							}
							want = append(want, "audio")
						}
						charge := uint8(1)
						if trace && infinite == 0 {
							charge = 0
						}
						if !slices.Equal(events, want) || ammo.Charge1 != charge || replacement.Charge1 != 77 {
							t.Fatalf("events=%v want=%v charge=%d", events, want, ammo.Charge1)
						}
					})
				}
			}
		}
	}
}

func TestPlayerAttackBow538960NativeShots(t *testing.T) {
	for _, flag := range []object.WeaponClass{object.WeaponBow, object.WeaponCrossbow} {
		t.Run(flag.String(), func(t *testing.T) {
			srv := installPlayerAttackProjectileServer538960(t)
			player := &server.Player{WeaponEquip: uint32(flag)}
			player.Info().SetField2239(37)
			weapon := &server.Object{TypeInd: 0x3214, ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(flag), UseData: server.UseDataPtr{Ptr: unsafe.Pointer(&server.AmmoUseData{})}, InitData: unsafe.Pointer(&server.ModifierInitData{})}
			quiver := &server.Object{ObjClass: object.ClassWeapon, ObjSubClass: 2, ObjFlags: object.FlagEquipped, UseData: server.UseDataPtr{Ptr: unsafe.Pointer(&server.AmmoUseData{Charge1: 2})}, InitData: unsafe.Pointer(&server.ModifierInitData{})}
			update := &server.PlayerUpdateData{Player: player, EquippedWeapon: weapon, Field0: 100, Field59_1: 0x65, Field59_2: 0x9876}
			owner := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), InvFirstItem: quiver}
			projectile := &server.Object{SpeedCur: 7, CollideData: unsafe.Pointer(&server.ArrowCollideData{}), InitData: unsafe.Pointer(&server.ModifierInitData{})}
			modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd)}
			srv.Modif.Dword_5d4594_251600 = modifier
			var pin runtime.Pinner
			defer pin.Unpin()
			pinPlayerAttackProjectilePointers538960(t, &pin, unsafe.Pointer(owner), unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(weapon), weapon.UseData.Ptr, weapon.InitData, unsafe.Pointer(quiver), quiver.UseData.Ptr, quiver.InitData, unsafe.Pointer(projectile), projectile.CollideData, projectile.InitData, unsafe.Pointer(modifier))
			frame := uint32(3)
			if flag == object.WeaponCrossbow {
				frame = 1
			}
			created, reported := false, false
			h := playerAttackBowRuntime538960{playerAttackProjectileRuntime538960: playerAttackProjectileRuntime538960{
				Frame: func() uint32 { return frame }, Readiness: func(*server.Object) int32 { return 0 }, AnimFrames: func(int) (int, int) { return 4, 0 },
				WeaponInventoryFlags: func(v *server.Object) uint32 { return uint32(v.ObjSubClass) }, QuestMode: func() bool { return false },
				Trace: func(types.Pointf, types.Pointf, server.MapTraceFlags) bool { return true }, NewObject: func(string) *server.Object { return projectile },
				CreateAt: func(v, u *server.Object, _ types.Pointf) {
					created = v == projectile && u == owner && (*server.ArrowCollideData)(v.CollideData).Owner == owner
				}, ApplyModifierAttrs: func(*server.Object, *server.ModifierInitData) {}, AudioEvent: func(sound.ID, *server.Object) {},
			}, Direction: func(server.Dir16) types.Pointf { return types.Ptf(1, 0) }, ReportCharges: func(_ uint8, v *server.Object, charge, _ uint8) { reported = v == quiver && charge == 1 }}
			old := playerAttackBowRuntimeFactory538960
			playerAttackBowRuntimeFactory538960 = func() playerAttackBowRuntime538960 { return h }
			defer func() { playerAttackBowRuntimeFactory538960 = old }()
			wantResult, wantFrame := 0, uint8(3)
			if flag == object.WeaponCrossbow {
				wantResult, wantFrame = 1, 1
			}
			if result := playerAttackNativeEntry538960(owner); result != wantResult || update.Field59_0 != wantFrame || update.Field59_1 != 0x65 || update.Field59_2 != 0x9876 || !created || !reported || projectile.VelVec != types.Ptf(7, 0) {
				t.Fatalf("native shot=%d created=%t reported=%t frame=%d", result, created, reported, update.Field59_0)
			}
		})
	}
}

func TestPlayerAttackBowShoot1_539BD0ReloadAndQuest(t *testing.T) {
	for _, flag := range []uint32{4, 8} {
		for _, delay := range []uint8{0, 2} {
			for _, quest := range []bool{false, true} {
				for _, reload := range []int{0, 1} {
					t.Run(fmt.Sprintf("flags-%d/delay-%d/quest-%t/reload-%d", flag, delay, quest, reload), func(t *testing.T) {
						owner := &server.Object{ObjClass: object.ClassPlayer}
						ammo := &server.AmmoUseData{Charge0: delay}
						weapon := &server.Object{UseData: server.UseDataPtr{Ptr: unsafe.Pointer(ammo)}}
						var events []string
						h := playerAttackBowRuntime538960{playerAttackProjectileRuntime538960: playerAttackProjectileRuntime538960{
							WeaponInventoryFlags: func(*server.Object) uint32 { return flag }, QuestMode: func() bool { return quest },
							Trace: func(types.Pointf, types.Pointf, server.MapTraceFlags) bool {
								events = append(events, "trace")
								return true
							}, NewObject: func(name string) *server.Object {
								if name != "WeakArcherArrow" {
									t.Fatal(name)
								}
								events = append(events, "weak")
								return nil
							},
							AudioEvent: func(id sound.ID, _ *server.Object) { events = append(events, fmt.Sprint(id)) },
						}, Direction: func(server.Dir16) types.Pointf { return types.Ptf(1, 0) }, ReloadQuiver: func(*server.Object) int { events = append(events, "reload"); return reload }, Priority: func(_ *server.Object, id strman.ID, b byte) {
							if b != 0 {
								t.Fatal("message argument")
							}
							events = append(events, string(id))
						}, SetState: func(_ *server.Object, state server.PlayerState) bool {
							if state != server.PlayerState13 {
								t.Fatal(state)
							}
							events = append(events, "idle")
							return true
						}}
						if result := playerAttackBowShoot1_539BD0(owner, weapon, h); result != 0 {
							t.Fatal("reload returned successful shot")
						}
						var want []string
						if delay != 0 {
							want = append(want, "pattack.c:ReloadingQuiver")
						} else {
							if quest && flag == 4 {
								want = append(want, "trace", "weak", fmt.Sprint(sound.ID(885)))
							}
							want = append(want, "reload")
							if reload == 1 {
								want = append(want, "pattack.c:ReloadQuiver")
							} else if !quest || flag != 4 {
								want = append(want, "pattack.c:NoQuiver")
							}
						}
						if flag == 8 {
							want = append(want, fmt.Sprint(sound.ID(888)))
						} else if !quest {
							want = append(want, fmt.Sprint(sound.ID(887)))
						}
						if delay == 0 {
							want = append(want, "idle")
						}
						wantDelay := delay
						if delay != 0 {
							wantDelay--
						}
						if !slices.Equal(events, want) || ammo.Charge0 != wantDelay {
							t.Fatalf("events=%v want=%v delay=%d", events, want, ammo.Charge0)
						}
					})
				}
			}
		}
	}
}
