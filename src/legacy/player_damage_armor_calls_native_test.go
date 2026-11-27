package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageArmorCallsGoOwned(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	noxflags.SetGame(noxflags.GameModeCoop)
	const armorType = 301
	definition := &server.Modifier{TypeInd: armorType, DamageCoeffOrArmor64: 0.25}
	oldDefinitions := srv.Modif.Dword_5d4594_251608
	srv.Modif.Dword_5d4594_251608 = definition
	t.Cleanup(func() { srv.Modif.Dword_5d4594_251608 = oldDefinitions })

	effect := &server.ModifierEff{Defend76: server.ModifierEffFnc{
		Fnc: playerDamageLateDefendFunctionsNative4E1320().armorMultiplier, Valf: 0.5,
	}}
	init := &server.ModifierInitData{}
	init.Modifiers[0] = effect
	owner := &server.Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10,
		UpdateData: unsafe.Pointer(&server.MonsterUpdateData{Field518: math.Float32bits(0.5)}),
	}
	item := &server.Object{
		ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, TypeInd: armorType,
		InvHolder: owner, InitData: unsafe.Pointer(init),
		UpdateData: unsafe.Pointer(&server.WeaponArmorUpdateData{}),
		HealthData: &server.HealthData{Cur: 10, Max: 10},
	}
	owner.InvFirstItem = item
	source := &server.Object{HealthData: &server.HealthData{Cur: 20}}
	effective := &server.Object{HealthData: &server.HealthData{Cur: 15}}

	t.Run("armor-value", func(t *testing.T) {
		if got := playerDamageRuntime4E17B0(srv).ItemArmorValue(item); got != 0.125 {
			t.Fatalf("armor value = %g, want 0.125", got)
		}
	})
	t.Run("armor-wear", func(t *testing.T) {
		calls := 0
		item.Damage = objectDamageNativeProbePtr()
		server.RegisterObjectDamageGo(fmt.Sprintf("GoOwnedArmorWear%d", objectDamageNativeTestSequence.Add(1)), item.Damage,
			func(target, attacker, weapon *server.Object, damage int32, typ object.DamageType) bool {
				calls++
				if target != item || attacker != source || weapon != effective || damage != 2 || typ != object.DamageImpale {
					t.Fatalf("armor damage arguments = %p/%p/%p/%d/%d", target, attacker, weapon, damage, typ)
				}
				target.HealthData.Cur -= uint16(damage)
				return true
			})
		Nox_xxx_playerDamageItems_4E2180(owner, source, effective, 8, object.DamageImpale)
		if calls != 1 || item.HealthData.Cur != 8 {
			t.Fatalf("armor wear called %d times with health %d", calls, item.HealthData.Cur)
		}
	})
}

func TestArmorDefendCachesAttributesBeforeLookup(t *testing.T) {
	fn := playerDamageLateDefendFunctionsNative4E1320().armorMultiplier
	old := &server.ModifierInitData{}
	replacement := &server.ModifierInitData{}
	replacement.Modifiers[0] = &server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fn, Valf: 0.25}}
	item := &server.Object{ObjClass: object.ClassArmor, TypeInd: 0xf301, InitData: unsafe.Pointer(old)}
	r := worldPlayerDefaults40A830
	r.ArmorDefinition = func(kind uint16) *server.Modifier {
		if kind != item.TypeInd {
			t.Fatalf("armor lookup type = %#x, want %#x", kind, item.TypeInd)
		}
		item.InitData = unsafe.Pointer(replacement)
		old.Modifiers[0] = &server.ModifierEff{Defend76: server.ModifierEffFnc{Fnc: fn, Valf: 0.5}}
		return &server.Modifier{DamageCoeffOrArmor64: 0.25}
	}
	defer pushWorldPlayerRuntime40A830(&r)()
	if got := itemDefendNative415C00(item); got != 0.125 {
		t.Fatalf("armor value with replaced attributes = %g, want 0.125", got)
	}
}
