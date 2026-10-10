package opennox

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The one substituted boundary is ObserveClear, which introduces MISSILE
// after armed-EXPLOSION admission. This is a controlled callback fixture, not
// a claim that an ordinary stock Sword naturally changes class during combat.
// The registered damage path, enemy/facing/reflection/owner/audio/RNG/state,
// WeaponDamage/DefaultDamage and C UnitSetHP services are not substituted.
func TestPlayerDamageArmedGreatSwordNative4E17B0RegisteredDurabilityAndHP(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	oldServer, oldGetServer := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGetServer })
	oldGame, oldEngine, oldGameplay := noxflags.GetGame(), noxflags.GetEngine(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldGame)
	noxflags.UnsetEngine(oldEngine)
	noxflags.UnsetGamePlay(oldGameplay)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGame(oldGame)
		noxflags.SetEngine(oldEngine)
		noxflags.SetGamePlay(oldGameplay)
	})
	s.SetFrame(1400)
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "gamedata.yml"), []byte("ItemDamageFromBlockPercentage: 0.25\n"), 0600); err != nil {
		t.Fatal(err)
	}
	oldData := datapath.Data()
	datapath.SetData(dataDir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := s.Balance.Read(); err != nil || s.Balance.Float("ItemDamageFromBlockPercentage") != .25 {
		t.Fatalf("synthetic block balance: %v", err)
	}
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.ID(890).String()
	audioData := append([]byte{byte(len(name))}, []byte(name)...)
	audioData = append(audioData, 7, 1, 'a', 0, 0)
	if !s.Audio.ReadAVNT(&npcShieldAudioFile4E17B0{reader: bytes.NewReader(audioData)}) || s.Audio.Field12(890) != 1 {
		t.Fatal("synthetic GreatSword sound descriptor")
	}
	s.Audio.Reset()
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "GameBall"}); err != nil {
		t.Fatal(err)
	}
	for _, callback := range []string{"PlayerDamage", "WeaponDamage"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeArmedGreatSword" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(80) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, playerSource := range []bool{false, true} {
		for _, subclass := range []uint32{0, 2} {
			for _, block := range []bool{false, true} {
				t.Run(fmt.Sprintf("player-source-%t/subclass-%x/block-%t", playerSource, subclass, block), func(t *testing.T) {
					newObject := func(callback string, ind uint16) *server.Object {
						def := &server.ObjectType{}
						if callback != "" {
							def.Damage = s.Types.ByID("NativeArmedGreatSword" + callback).Damage
						}
						obj := s.Objs.NewObject(def)
						obj.TypeInd, obj.ObjFlags = ind, 0
						return obj
					}
					v, a, w, item := newObject("PlayerDamage", 713), newObject("", 1343), newObject("", 1174), newObject("WeaponDamage", 777)
					sibling, firstOwned := newObject("", 778), newObject("", 779)
					pu, freePU := alloc.New(server.PlayerUpdateData{})
					apu, freeAPU := alloc.New(server.PlayerUpdateData{})
					amu, freeAMU := alloc.New(server.MonsterUpdateData{})
					info, freeInfo := alloc.New(server.Player{})
					ainfo, freeAInfo := alloc.New(server.Player{})
					init, freeInit := alloc.New(server.ModifierInitData{})
					winit, freeWInit := alloc.New(server.ModifierInitData{})
					iu, freeIU := alloc.New(server.WeaponArmorUpdateData{})
					hp, freeHP := alloc.New(server.HealthData{})
					ihp, freeIHP := alloc.New(server.HealthData{})
					for _, free := range []func(){freePU, freeAPU, freeAMU, freeInfo, freeAInfo, freeInit, freeWInit, freeIU, freeHP, freeIHP} {
						t.Cleanup(free)
					}
					*info, *ainfo = server.Player{PlayerInd: 9, WeaponEquip: 0x400}, server.Player{PlayerInd: 10}
					*pu = server.PlayerUpdateData{Player: info, State: server.PlayerState13,
						Field57: math.Float32bits(.5), Field21: math.Float32bits(.25), Field76: 99, Field75: 77}
					*apu = server.PlayerUpdateData{Player: ainfo, State: server.PlayerState13}
					*hp, *ihp = server.HealthData{Cur: 2000, Max: 2000, Field2: 2000}, server.HealthData{Cur: 20, Max: 20, Field2: 20}
					v.ObjClass, v.ObjFlags, v.UpdateData, v.HealthData = object.Class(2621444), object.Flags(16777732), unsafe.Pointer(pu), hp
					a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = object.Class(2097154), 69778, object.Flags(16777732), unsafe.Pointer(amu)
					if playerSource {
						a.TypeInd, a.ObjClass, a.ObjSubClass, a.UpdateData = 713, object.Class(2621444), 0, unsafe.Pointer(apu)
					}
					a.PrevPos = types.Ptf(32.5, -18.25)
					w.InitData, w.ObjFlags = unsafe.Pointer(winit), object.Flags(17301776)
					w.PrevPos, w.PosVec = types.Ptf(20, 0), types.Ptf(-20, 0)
					w.ObjOwner, w.Field128, a.Field129, v.Field129 = a, sibling, w, firstOwned
					item.ObjClass, item.ObjSubClass, item.ObjFlags = object.ClassWeapon, 0x400, object.FlagEquipped
					item.InitData, item.UpdateData, item.HealthData = unsafe.Pointer(init), unsafe.Pointer(iu), ihp
					v.InvFirstItem, item.InvHolder = item, v
					v.TeamVal.ID, a.TeamVal.ID = 2, 1
					if !s.IsEnemyTo(v, a) || legacy.Nox_server_testTwoPointsAndDirection_4E6E50(v.PosVec, int16(v.Direction1), w.PrevPos)&1 == 0 {
						t.Fatal("production hostility/facing")
					}
					for _, ptr := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(item), unsafe.Pointer(sibling), unsafe.Pointer(firstOwned), v.UpdateData, a.UpdateData, item.InitData, item.UpdateData, w.InitData, unsafe.Pointer(info), unsafe.Pointer(ainfo), unsafe.Pointer(hp), unsafe.Pointer(ihp)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
							t.Fatalf("C-owned record=%p, want >4 GiB", ptr)
						}
					}
					oldObserve := legacy.Nox_xxx_playerObserveClear_4DDEF0
					defer func() { legacy.Nox_xxx_playerObserveClear_4DDEF0 = oldObserve }()
					observeCalls, audioCalls, activeHit := 0, 0, 0
					legacy.Nox_xxx_playerObserveClear_4DDEF0 = func(got *server.Object) {
						if got != v || pu.Field76 != 0 || w.Class().Has(object.ClassMissile) || w.SubClass() != 256 {
							t.Fatal("observer must execute after marker clear and ordinary Sword admission")
						}
						observeCalls++
						w.ObjClass |= object.ClassMissile
						w.ObjSubClass = object.SubClass(subclass)
					}
					if block {
						info.Field3680, info.CameraFollowObj = 2, a
					}
					s.Audio.OnSound(func(id sound.ID, _ int, owner *server.Object, _ types.Pointf) {
						if id != sound.ID(890) || owner != v {
							return
						}
						audioCalls++
						if pu.Field76 != 1 || pu.Field75 != 1174 || hp.Cur != 2000 || ihp.Cur != uint16(20-2*(activeHit-1)) {
							t.Fatal("production audio must follow attribution and precede actual wear/HP")
						}
					})
					var hpSteps, wearSteps [3]uint16
					for hit := 1; hit <= 3; hit++ {
						activeHit = hit
						w.ObjClass, w.ObjSubClass = object.Class(85458944), 256
						w.Direction1, w.VelVec, w.NewPos = 128, types.Ptf(-4, 0), types.Pointf{}
						result := v.CallDamage(a, w, 8, object.DamageExplosion)
						if block {
							if result || hp.Cur != 2000 || ihp.Cur != uint16(20-2*hit) || observeCalls != hit || audioCalls != hit ||
								pu.State < server.PlayerState18 || pu.State > server.PlayerState20 || iu.Field0 != 0 ||
								item.Obj130 != w || item.Field131 != 7 || item.Frame134 != 1400 || v.Obj130 != nil {
								t.Fatalf("actual block hit=%d result=%t HP=%d durability=%d observer=%d audio=%d state=%d", hit, result, hp.Cur, ihp.Cur, observeCalls, audioCalls, pu.State)
							}
							if w.Direction1 != 0 || w.NewPos != w.PrevPos || w.VelVec.X <= 0 || w.VelVec.Y != 0 {
								t.Fatal("actual projectile reflection")
							}
							if subclass == 0 {
								if w.ObjOwner != v || v.Field129 != w || w.Field128 != firstOwned || a.Field129 != sibling {
									t.Fatal("actual reflected owner-list transfer")
								}
							} else if w.ObjOwner != a || a.Field129 != w || w.Field128 != sibling || v.Field129 != firstOwned {
								t.Fatal("bit 2 must retain actual owner lists")
							}
						} else if !result || hp.Cur != uint16(2000-4*hit) || ihp.Cur != 20 || observeCalls != 0 || audioCalls != 0 ||
							pu.State != server.PlayerState13 || v.Obj130 != w || v.Field131 != 7 || v.Frame134 != 1400 || item.Obj130 != nil ||
							w.ObjOwner != a || a.Field129 != w || v.Field129 != firstOwned || w.Direction1 != 128 || w.NewPos != (types.Pointf{}) {
							t.Fatalf("ordinary type-7 control hit=%d result=%t HP=%d durability=%d state=%d", hit, result, hp.Cur, ihp.Cur, pu.State)
						}
						if pu.Field21 != math.Float32bits(.25) || pu.Field76 != 1 || pu.Field75 != 1174 || item.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
							item.InvHolder != v || v.InvFirstItem != item || s.Objs.DeletedList != nil {
							t.Fatal("nonlethal registered damage changed carry/inventory/deletion")
						}
						hpSteps[hit-1], wearSteps[hit-1] = hp.Cur, ihp.Cur
					}
					// CallDamage is the native Go registry, not a C-entry probe.
					// Item HP and unblocked unit HP reach the real C HP boundary.
					t.Logf("registered production >4 GiB player-source=%t owner-bit=%d observer-missile=%t damage=8/type=7 target=%p item=%p HP=2000->%v durability=20->%v", playerSource, subclass, block, v, item, hpSteps, wearSteps)
				})
			}
		}
	}
}
