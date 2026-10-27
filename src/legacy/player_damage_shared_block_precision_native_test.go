package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// The child isolates the real balance loader's data root. C dispatch,
// PlayerDamage, block wear, EquipDamage and registered ArmorDamage execute;
// unit/item HP and durability callbacks are not replaced.
func TestPlayerDamageSharedBlockNative4E17B0GearHP(t *testing.T) {
	if os.Getenv("NOX_SHARED_BLOCK_GEAR_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestPlayerDamageSharedBlockNative4E17B0GearHP$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), "NOX_SHARED_BLOCK_GEAR_CHILD=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual C shared block gear child:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	blob, err := os.ReadFile("testdata/shared_block_gear_chain_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		Rows      []struct {
			Input struct {
				Player       bool   `json:"player"`
				PlayerSource bool   `json:"player_source"`
				Defense      string `json:"defense"`
			} `json:"input"`
			Mode   string  `json:"mode"`
			Damage []int32 `json:"integer_gear_damage"`
			Carry  uint32  `json:"gear_carry_bits32"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Rows) != 16 {
		t.Fatal("incomplete original block/durability chain")
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, proc := range []string{"PlayerDamage", "ArmorDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeSharedBlock" + proc, OnDamage: &things.ProcFunc{Name: proc}}); err != nil {
			t.Fatal(err)
		}
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "gamedata.yml"), []byte("ItemDamageFromBlockPercentage: 0.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	datapath.SetData(root)
	if err := srv.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if srv.Balance.Float("ItemDamageFromBlockPercentage") != .3 {
		t.Fatal("real balance input unavailable")
	}
	srv.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int { return 1 }
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" || row.Input.Defense != "shield" {
			continue
		}
		t.Run(fmt.Sprintf("player-%t/source-player-%t", row.Input.Player, row.Input.PlayerSource), func(t *testing.T) {
			if len(row.Damage) != 1 || row.Damage[0] != 2 {
				t.Fatal("original chain did not forward two item damage")
			}
			makeUnit := func(player bool, index byte) *server.Object {
				hp := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 200, Field2: 200})
				u := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 1343, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, HealthData: hp, Material: 0x4000})
				if player {
					info := selfWeaponElectricNativeAlloc(t, server.Player{PlayerUnit: u, PlayerInd: index})
					ud := selfWeaponElectricNativeAlloc(t, server.PlayerUpdateData{Player: info, State: server.PlayerState16, Field21: math.Float32bits(.25)})
					info.ArmorEquip = 0x1000000
					u.TypeInd, u.ObjClass, u.ObjSubClass, u.UpdateData = 713, object.ClassPlayer, 0, unsafe.Pointer(ud)
				} else {
					ud := selfWeaponElectricNativeAlloc(t, server.MonsterUpdateData{ArmorEquipFlags: 0x1000000, Field1: math.Float32bits(.25)})
					ud.AIStack[0].Action = 21
					u.UpdateData = unsafe.Pointer(ud)
				}
				return u
			}
			target, source := makeUnit(row.Input.Player, 8), makeUnit(row.Input.PlayerSource, 9)
			target.TeamVal.ID, source.TeamVal.ID = 2, 1
			target.Damage = srv.Types.ByID("NativeSharedBlockPlayerDamage").Damage
			weapon := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 73, ObjClass: object.ClassMissile, ObjSubClass: 0x10, ObjOwner: source, PosVec: types.Ptf(19, 11), PrevPos: types.Ptf(17, 9)})
			carry := selfWeaponElectricNativeAlloc(t, server.WeaponArmorUpdateData{Field0: math.Float32bits(.3)})
			hp := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 100, Max: 100, Field2: 100})
			init := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{})
			item := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, InvHolder: target, HealthData: hp, UpdateData: unsafe.Pointer(carry), InitData: unsafe.Pointer(init), Damage: srv.Types.ByID("NativeSharedBlockArmorDamage").Damage})
			target.InvFirstItem = item
			if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(item.CObj()) <= math.MaxUint32 || uintptr(unsafe.Pointer(carry)) <= math.MaxUint32) {
				t.Fatal("expected native C-owned pointers above 4 GiB")
			}
			if Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), weapon.PrevPos)&1 == 0 {
				t.Fatal("native shield facing")
			}
			blockedResult := objectDamageDispatchCallNative(target, source, weapon, 9, object.DamageImpale)
			if blockedResult || target.HealthData.Cur != 200 || hp.Cur != uint16(100-row.Damage[0]) || carry.Field0 != row.Carry {
				t.Fatalf("real C shield: result=%t unit HP=%d item HP=%d carry=%08x want HP=%d carry=%08x", blockedResult, target.HealthData.Cur, hp.Cur, carry.Field0, 100-row.Damage[0], row.Carry)
			}
			t.Logf("registered C PlayerDamage -> EquipDamage -> ArmorDamage: unit HP200, item HP100->%d carry=%08x", hp.Cur, carry.Field0)
		})
	}
}
