package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// DefaultDamage's integer context is reinterpreted as binary32 by these
// original callbacks. Keep the real C dispatcher, item loop and HP setter.
func TestModifierDefendAdaptersNativeHP(t *testing.T) {
	blob, err := os.ReadFile("testdata/modifier_defend_multipliers_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA  string `json:"original_sha256"`
		Rows []struct {
			Function string `json:"function"`
			Name     string `json:"name"`
			Mode     string `json:"mode"`
			Inputs   struct {
				Value       uint32 `json:"value_bits"`
				Coefficient uint32 `json:"coefficient_bits"`
			} `json:"inputs"`
			Value uint32 `json:"value_bits"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 256 {
		t.Fatal("original multiplier fixture unavailable")
	}
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeModifierDefaultDamage", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	srv.NetSendPacketXxx = func(int, []byte, *server.Object, int, int) int { return 1 }
	oldSound, oldState := Nox_xxx_soundPlayerDamageSound_5328B0, Nox_xxx_playerSetState_4FA020
	var soundTarget, soundSource *server.Object
	Nox_xxx_soundPlayerDamageSound_5328B0 = func(target, source *server.Object) int {
		if target != soundTarget || source != soundSource {
			t.Fatal("native modifier sound attribution")
		}
		return 1
	}
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold modifier requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_soundPlayerDamageSound_5328B0, Nox_xxx_playerSetState_4FA020 = oldSound, oldState })
	fns := playerDamageLateDefendFunctionsNative4E1320()
	selected := 0
	for _, row := range fixture.Rows {
		wanted := row.Function == "ArmorMultiplierEffect" && row.Name == "subnormal-chop" ||
			row.Function == "DurabilityMultiplierEffect" && row.Name == "subnormal-tie"
		if row.Mode != "actual-startup-inherited" || !wanted {
			continue
		}
		selected++
		for _, slot := range []int{2, 3} {
			t.Run(fmt.Sprintf("%s/slot-%d", row.Function, slot), func(t *testing.T) {
				fn := fns.armorMultiplier
				if row.Function == "DurabilityMultiplierEffect" {
					fn = fns.durabilityMultiplier
				}
				hp := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 200, Field2: 200})
				target := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 713, ObjClass: object.ClassPlayer, HealthData: hp, Material: 0x4000})
				info := selfWeaponElectricNativeAlloc(t, server.Player{PlayerUnit: target, PlayerInd: 8})
				target.UpdateData = unsafe.Pointer(selfWeaponElectricNativeAlloc(t, server.PlayerUpdateData{Player: info, State: server.PlayerState13}))
				source := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 1343, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, Material: 0x4000, PrevPos: types.Ptf(12.5, -40.25)})
				source.UpdateData = unsafe.Pointer(selfWeaponElectricNativeAlloc(t, server.MonsterUpdateData{}))
				source.HealthData = selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 200})
				effect := selfWeaponElectricNativeAlloc(t, server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fn, Valf: math.Float32frombits(row.Inputs.Coefficient)}})
				init := selfWeaponElectricNativeAlloc(t, server.ModifierInitData{})
				init.Modifiers[slot] = effect
				itemHP := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 100, Max: 100})
				item := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init), InvHolder: target, HealthData: itemHP})
				target.InvFirstItem = item
				target.Damage = srv.Types.ByID("NativeModifierDefaultDamage").Damage
				target.TeamVal.ID, source.TeamVal.ID = 2, 1
				soundTarget, soundSource = target, source
				if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(source.CObj()) <= math.MaxUint32 || uintptr(item.CObj()) <= math.MaxUint32) {
					t.Fatal("expected native C-owned records above 4 GiB")
				}
				ok := objectDamageDispatchCallNative(target, source, source, int32(row.Inputs.Value), object.DamageBlade)
				wantHP := uint16(200 - row.Value)
				if !ok || hp.Cur != wantHP || itemHP.Cur != 100 || target.InvFirstItem != item || item.InvHolder != target || target.Obj130 != source || target.Field131 != uint32(object.DamageBlade) || target.Frame134 != 1400 {
					t.Fatalf("native modifier HP=%d want=%d result=%t item HP=%d", hp.Cur, wantHP, ok, itemHP.Cur)
				}
				t.Logf("registered DefaultDamage -> equipped multiplier -> UnitSetHP: raw=%d word=%d HP=200->%d slot=%d", row.Inputs.Value, row.Value, hp.Cur, slot)
			})
		}
	}
	if selected != 2 {
		t.Fatalf("native HP original contexts=%d", selected)
	}
}
