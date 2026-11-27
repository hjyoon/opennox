package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestRawDefaultDamageItemEffectsGoHeap4E0B30(t *testing.T) {
	npcReflectLegacyServer4E17B0(t)

	preDamage := itemPreDamageFunctionsNative4E13B0()
	vampirism := &server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: preDamage.vampirism, Valf: 0.5}}
	sympathy := &server.ModifierEff{AttackPreDmg64: server.ModifierEffFnc{Fnc: preDamage.sympathy, Valf: 0.25}}
	weapon := &server.Object{InitData: unsafe.Pointer(&server.ModifierInitData{
		Modifiers: [4]*server.ModifierEff{vampirism, nil, nil, sympathy},
	})}

	armor := &server.ModifierEff{Defend76: server.ModifierEffFnc{
		Fnc: playerDamageLateDefendFunctionsNative4E1320().armorMultiplier, Valf: 0.25,
	}}
	item := &server.Object{
		ObjFlags: object.FlagEquipped,
		InitData: unsafe.Pointer(&server.ModifierInitData{Modifiers: [4]*server.ModifierEff{nil, nil, armor}}),
	}
	target := &server.Object{
		ObjClass:     object.ClassMonster,
		HealthData:   &server.HealthData{Cur: 20, Max: 20},
		InvFirstItem: item,
	}
	source := &server.Object{ObjClass: object.ClassMonster, HealthData: &server.HealthData{Cur: 19, Max: 20}}

	var damage int32
	call := rawDefaultDamageNativeService4E0B30{Arguments: []rawDefaultDamageNativeArgument4E0B30{
		{Kind: "object", Pointer: target.CObj()},
		{Kind: "object", Pointer: source.CObj()},
		{Kind: "object", Pointer: weapon.CObj()},
		{Kind: "damage", Damage: &damage},
		{Kind: "word", Word: uint32(object.DamageBite)},
	}}
	t.Run("pre-damage", func(t *testing.T) {
		call.Site = 0x4e0fef
		damage = 8
		result := rawDefaultDamageInvokeService4E0B30(call)
		// Vampirism reaches the 20 HP cap before Sympathy deducts 2 HP.
		if result.Word != 0 || damage != 8 || source.HealthData.Cur != 18 || target.HealthData.Cur != 20 {
			t.Fatalf("result=%d damage=%d source HP=%d target HP=%d", result.Word, damage, source.HealthData.Cur, target.HealthData.Cur)
		}
	})
	t.Run("equipped-defend", func(t *testing.T) {
		call.Site = 0x4e0f77
		damage = int32(math.Float32bits(8))
		result := rawDefaultDamageInvokeService4E0B30(call)
		if result.Word != 0 || uint32(damage) != math.Float32bits(2) {
			t.Fatalf("result=%d damage bits=%08x", result.Word, uint32(damage))
		}
	})
}
