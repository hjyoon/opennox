package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type npcShieldAudioFile4E17B0 struct {
	server.File
	reader *bytes.Reader
}

func (f *npcShieldAudioFile4E17B0) ReadU8() uint8 {
	v, err := f.reader.ReadByte()
	if err != nil {
		panic(err)
	}
	return v
}

func (f *npcShieldAudioFile4E17B0) ReadI16() int16 {
	var data [2]byte
	if _, err := io.ReadFull(f.reader, data[:]); err != nil {
		panic(err)
	}
	return int16(binary.LittleEndian.Uint16(data[:]))
}

func (f *npcShieldAudioFile4E17B0) ReadString8() (string, error) {
	data := make([]byte, int(f.ReadU8()))
	_, err := io.ReadFull(f.reader, data)
	return string(data), err
}

func (f *npcShieldAudioFile4E17B0) Skip(size int) {
	if _, err := f.reader.Seek(int64(size), io.SeekCurrent); err != nil {
		panic(err)
	}
}

// Unlike a substituted durability/pop callback, this fixture uses the actual
// root DelayedDelete binding, typed C NPC dequip calls, inventory unlinking,
// owner lists, UnitSetHP, and AI pop. Every object has its allocator's server
// handle, and all transient records are C-owned native-width allocations.
func TestPlayerDamageNPCShieldNative4E17B0DurabilityAndBreak(t *testing.T) {
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
	if err := s.Balance.Read(); err != nil || s.Balance.Float("ItemDamageFromBlockPercentage") != 0.25 {
		t.Fatalf("synthetic balance load: %v", err)
	}
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	// Load one synthetic AVNT descriptor and enter the production audio loop;
	// without this, the ordinary sound service only queues next-frame events.
	name := sound.ID(878).String()
	audioData := append([]byte{byte(len(name))}, []byte(name)...)
	audioData = append(audioData, 7, 1, 'a', 0, 0)
	if !s.Audio.ReadAVNT(&npcShieldAudioFile4E17B0{reader: bytes.NewReader(audioData)}) || s.Audio.Field12(878) != 1 {
		t.Fatal("synthetic native block-sound descriptor load failed")
	}
	s.Audio.Reset()
	// Resolve the production callback identities through the ordinary type
	// parser; do not register a replacement HP/durability/damage callback.
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	for _, callback := range []string{"PlayerDamage", "ArmorDamage"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeShield" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.Init(160) {
		t.Fatal("native object allocator initialization failed")
	}
	t.Cleanup(s.Objs.FreeObjects)
	for _, armor := range []uint32{0x1000000, 0x2000000} {
		for _, flags := range []uint32{0, 2, 0x10, 0x40} {
			for _, typ := range []object.DamageType{object.DamageImpact, object.DamageImpale} {
				t.Run(fmt.Sprintf("armor-%x/subclass-%x/%s", armor, flags, typ), func(t *testing.T) {
					newObject := func(callback string, ind uint16) *server.Object {
						def := &server.ObjectType{}
						if callback != "" {
							def.Damage = s.Types.ByID("NativeShield" + callback).Damage
						}
						obj := s.Objs.NewObject(def)
						obj.TypeInd, obj.ObjFlags = ind, 0
						return obj
					}
					target, source, missile := newObject("PlayerDamage", 71), newObject("", 72), newObject("", 73)
					shield, decoy, later := newObject("ArmorDamage", 74), newObject("", 75), newObject("", 76)
					sibling, firstOwned := newObject("", 77), newObject("", 78)
					ud, freeUD := alloc.New(server.MonsterUpdateData{})
					playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					initData, freeInit := alloc.New(server.ModifierInitData{})
					itemUpdate, freeItemUpdate := alloc.New(server.WeaponArmorUpdateData{})
					npcHP, freeNPCHP := alloc.New(server.HealthData{})
					shieldHP, freeShieldHP := alloc.New(server.HealthData{})
					*npcHP, *shieldHP = server.HealthData{Cur: 60, Max: 60}, server.HealthData{Cur: 3, Max: 3}
					*playerUD, *player, *initData, *itemUpdate = server.PlayerUpdateData{}, server.Player{}, server.ModifierInitData{}, server.WeaponArmorUpdateData{}
					for _, free := range []func(){freeUD, freePlayerUD, freePlayer, freeInit, freeItemUpdate, freeNPCHP, freeShieldHP} {
						t.Cleanup(free)
					}
					target.ObjClass, target.ObjSubClass = object.ClassMonster, 0x11012
					target.UpdateData, target.HealthData = unsafe.Pointer(ud), npcHP
					source.ObjClass, source.UpdateData, playerUD.Player = object.ClassPlayer, unsafe.Pointer(playerUD), player
					*ud = server.MonsterUpdateData{ArmorEquipFlags: armor, AIStackInd: 1, Field547: 99, Field546: 77,
						Field1: math.Float32bits(0.25), Field2: 19, Field67: 20, Field74: 21, Field91: 22,
						Field120_1: 23, Field120_2: 24, Field120_3: 25, Field124: 26, Field137: 27}
					ud.AIStack[0].Action, ud.AIStack[1].Action = uint32(ai.ACTION_GUARD), uint32(ai.ACTION_BLOCK_ATTACK)
					shield.ObjClass, shield.ObjSubClass, shield.ObjFlags = object.ClassArmor, 2, object.FlagEquipped
					shield.InitData, shield.UpdateData, shield.HealthData = unsafe.Pointer(initData), unsafe.Pointer(itemUpdate), shieldHP
					decoy.ObjClass, decoy.ObjSubClass, later.ObjClass = object.ClassSimple, 2, object.ClassSimple
					target.InvFirstItem, decoy.InvNextItem, shield.InvNextItem = decoy, shield, later
					shield.Field125, later.Field125 = decoy, shield
					decoy.InvHolder, shield.InvHolder, later.InvHolder = target, target, target
					missile.ObjClass, missile.ObjSubClass = object.ClassMissile, object.SubClass(flags)
					missile.PrevPos, missile.PosVec = types.Pointf{X: 20}, types.Pointf{X: -20}
					missile.Direction1, missile.VelVec = 128, types.Pointf{X: -4}
					missile.ObjOwner, missile.Field128, source.Field129, target.Field129 = source, sibling, missile, firstOwned
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(source), unsafe.Pointer(missile), unsafe.Pointer(shield), unsafe.Pointer(decoy), unsafe.Pointer(later), unsafe.Pointer(sibling), unsafe.Pointer(firstOwned), target.UpdateData, source.UpdateData, shield.InitData, shield.UpdateData, unsafe.Pointer(npcHP), unsafe.Pointer(shieldHP), unsafe.Pointer(player)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("native shield pointer=%p, want address above 4 GiB", pointer)
						}
					}
					if !s.IsEnemyTo(target, source) || legacy.Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), missile.PrevPos)&1 == 0 {
						t.Fatal("fixture must be hostile and facing the projectile's previous position")
					}
					blockAudio := 0
					s.Audio.OnSound(func(id sound.ID, _ int, owner *server.Object, _ types.Pointf) {
						if id != sound.ID(878) || owner != target {
							return
						}
						blockAudio++
						if ud.Field547 != 1 || ud.Field546 != uint32(missile.TypeInd) || target.HealthData.Cur != 60 || ud.AIStackInd != 1 {
							t.Fatal("production block audio occurred before attribution or after wear/pop")
						}
					})
					before := *ud
					before.Field547, before.Field546 = 1, uint32(missile.TypeInd)
					// This is the same registered PlayerDamage call used by the
					// game, including real C HP/dequip boundaries. First wear is
					// nonlethal; the second hit destroys and unlinks the shield.
					for hit := 1; hit <= 2; hit++ {
						s.AI.StackChanged = false
						// Reuse the native projectile for a second incoming flight,
						// retaining its real owner-list state from the first block.
						missile.Direction1, missile.VelVec, missile.NewPos = 128, types.Pointf{X: -4}, types.Pointf{}
						if target.CallDamage(source, missile, 8, typ) || npcHP.Cur != 60 || blockAudio != hit || itemUpdate.Field0 != 0 || shield.Obj130 != missile || shield.Field131 != uint32(typ) || shield.Frame134 != 1400 {
							t.Fatalf("registered NPC block hit=%d HP=%d item=%d audio=%d attribution=%p", hit, npcHP.Cur, shieldHP.Cur, blockAudio, shield.Obj130)
						}
						if flags&0x70 == 0 {
							if missile.Direction1 != 0 || missile.NewPos != missile.PrevPos || missile.VelVec.X <= 0 || missile.VelVec.Y != 0 {
								t.Fatal("production projectile reflection lost direction/position/velocity")
							}
							if flags&2 == 0 {
								if missile.ObjOwner != target || target.Field129 != missile || missile.Field128 != firstOwned || source.Field129 != sibling {
									t.Fatal("production reflection lost native owner-list links")
								}
							} else if missile.ObjOwner != source || source.Field129 != missile || missile.Field128 != sibling || target.Field129 != firstOwned {
								t.Fatal("subclass bit 2 must reflect without transferring owner")
							}
						} else if missile.Direction1 != 128 || missile.VelVec.X != -4 || missile.NewPos != (types.Pointf{}) || missile.ObjOwner != source || source.Field129 != missile {
							t.Fatal("unreflectable missile changed velocity/direction/owner")
						}
						if hit == 1 {
							if shieldHP.Cur != 1 || shield.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || *ud != before || s.AI.StackChanged || shield.InvHolder != target || decoy.InvNextItem != shield || later.Field125 != shield || s.Objs.DeletedList != nil {
								t.Fatal("nonlethal production shield wear changed NPC action/inventory")
							}
							continue
						}
						if shieldHP.Cur != 0 || !shield.Flags().Has(object.FlagDead|object.FlagDestroyed) || shield.Flags().Has(object.FlagEquipped) || shield.InvHolder != nil || target.InvFirstItem != decoy || decoy.InvNextItem != later || later.Field125 != decoy || s.Objs.DeletedList != shield || shield.DeletedAt != 1400 {
							t.Fatal("lethal production shield wear did not dequip/unlink/queue deletion")
						}
						if ud.AIStackInd != 0 || ud.AIStack[0].Type() != ai.ACTION_GUARD || !s.AI.StackChanged || ud.Field1 != math.Float32bits(0.25) || ud.Field2 != 0 || ud.Field67 != 0 || ud.Field74 != 0 || ud.Field91 != 0 || ud.Field120_1 != 0 || ud.Field120_2 != 0 || ud.Field120_3 != 0 || ud.Field124 != 1400 || ud.Field137 != 1400 {
							t.Fatal("actual destroyed-shield AI pop did not reset the native stack")
						}
					}
					// Preserve the original stale-equipment-mask case after the
					// shield has been unlinked: the real EquipDamage nil-item guard
					// must not damage HP or pop the newly active blocking action.
					ud.ArmorEquipFlags, ud.AIStackInd = armor, 1
					ud.AIStack[1].Action = uint32(ai.ACTION_BLOCK_ATTACK)
					s.AI.StackChanged = false
					missile.Direction1, missile.VelVec, missile.NewPos = 128, types.Pointf{X: -4}, types.Pointf{}
					if target.CallDamage(source, missile, 8, typ) || npcHP.Cur != 60 || blockAudio != 3 ||
						ud.AIStackInd != 1 || s.AI.StackChanged || s.Objs.DeletedList != shield ||
						target.InvFirstItem != decoy || decoy.InvNextItem != later || later.Field125 != decoy {
						t.Fatal("production nil-item shield wear changed HP/action/inventory/deletion")
					}
					t.Logf("registered NPC shield: target=%p missile=%p shield=%p type=%d HP=60->60 durability=3->1->0 action=BLOCK_ATTACK->GUARD; post-unlink nil-item wear preserves HP/action", target, missile, shield, typ)
					s.Objs.DeletedList = nil
				})
			}
		}
	}
}
