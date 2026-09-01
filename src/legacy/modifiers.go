package legacy

/*
#include "defs.h"

void nullsub_22();
void nullsub_36();
void nullsub_38();
void nullsub_39();
void nullsub_40();
void nullsub_41();
void nullsub_42();
void nullsub_43();
void nullsub_44();

int nox_xxx_spellCastCleansingFlame_52D5C0(int a1, nox_object_t* a2p, nox_object_t* a3p, nox_object_t* a4p, void* a5p, int a6);

void sub_4DFE10(void* effect, nox_object_t* owner, const nox_object_t* item);
float* sub_4E0370(void* effect, nox_object_t* item, uintptr_t a3, nox_object_t* target, uintptr_t a5, float* value);
float* sub_4E0380(void* effect, nox_object_t* item, uintptr_t a3, nox_object_t* target, uintptr_t a5, float* value);
float* nox_xxx_effectDamageMultiplier_4E04C0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, float* damage);
void nox_xxx_attribContinualReplen_4E02C0(int a1, uint32_t* a2);
void nox_xxx_confuseEffect_4E0670(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_drainMEffect_4E0740(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, int32_t* damage);
void nox_xxx_sympathyEffect_4E08E0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, int32_t* damage);
nox_object_t* nox_xxx_effectProjectileSpeed_4E09B0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, nox_object_t* projectile);
void nox_xxx_buff_4DFD80(void* effect, nox_object_t* owner, const nox_object_t* item);
void nox_xxx_checkPoisonProtectEnch_4DFDE0(void* effect, nox_object_t* owner, const nox_object_t* item);
int nox_xxx_gripEffect_4E0480(int a1, int a2, int a3, int a4, int a5, int* a6);
void nox_xxx_effectRegeneration_4E01D0(int a1, int a2);
void nox_xxx_stunEffect_4E04D0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_fireEffect_4E0550(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_fireRingEffect_4E05B0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_blueFREffect_4E05F0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_recoilEffect_4E0640(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_lightngEffect_4E06F0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
void nox_xxx_vampirismEffect_4E07C0(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, int32_t* damage);
void nox_xxx_poisonEffect_4E0850(void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, void* context);
int nox_xxx_inversionEffect_4E03D0(int a1, int a2, int a3, int a4, int a5, int* a6);
void sub_4DFB50(void* effect, nox_object_t* owner, const nox_object_t* item);
void sub_4DFB80(void* effect, nox_object_t* owner, const nox_object_t* item);
void nox_xxx_effectSpeedEngage_4DFC30(void* effect, nox_object_t* owner, const nox_object_t* item);
void nox_xxx_effectSpeedDisengage_4DFCA0(void* effect, nox_object_t* owner, const nox_object_t* item);
void sub_4DFD10(void* effect, nox_object_t* owner, const nox_object_t* item);
void nox_xxx_modifFireProtection_4DFD40(void* effect, nox_object_t* owner, const nox_object_t* item);
void sub_4DFDB0(void* effect, nox_object_t* owner, const nox_object_t* item);
void sub_4E0140(void* effect, nox_object_t* owner, const nox_object_t* item);
void sub_4E0170(void* effect, nox_object_t* owner, const nox_object_t* item);
int nox_xxx_enchantItemTestInventory_4DFBB0(nox_object_t* owner, uint8_t flag);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

var (
	Sub_4A5E90_A                  func()
	Nox_xxx_fireRingEffect_4E05B0 func(a1 unsafe.Pointer, a2p, a3p, a4p *server.Object)
	Nox_xxx_blueFREffect_4E05F0   func(a1 unsafe.Pointer, a2p, a3p, a4p *server.Object)
)

// InversionEffectPointer4E03D0 exposes only the identity of the registered C
// effect. Native collision code compares it without narrowing or invoking an
// ABI32 modifier callback.
func InversionEffectPointer4E03D0() unsafe.Pointer {
	return C.nox_xxx_inversionEffect_4E03D0
}

// PoisonProtectEffectPointer4DFDE0 exposes only the identity compared by the
// restored poison-protection traversal. The ABI32 effect is never invoked.
func PoisonProtectEffectPointer4DFDE0() unsafe.Pointer {
	return C.nox_xxx_checkPoisonProtectEnch_4DFDE0
}

// Nox_xxx_itemCheckReadinessEffect_4E0960 restores the modifier lookup without
// sending the native item or ModifierEff pointers through the ABI32 C body.
// Readiness effects occupy the two enchantment modifier slots.
func Nox_xxx_itemCheckReadinessEffect_4E0960(item *server.Object) int32 {
	const readinessClasses = uint32(0x13001000)
	if item == nil || uint32(item.ObjClass)&readinessClasses == 0 || item.InitData == nil {
		return 0
	}
	for _, mod := range item.InitDataModifier().Modifiers[2:] {
		if mod != nil && mod.Attack40.Fnc == C.nullsub_22 {
			return mod.Attack40.Val
		}
	}
	return 0
}

func ReadinessEffectPointer4E0960() unsafe.Pointer {
	return C.nullsub_22
}

func modifierEngagePointerNative4DFBB0(flag byte) unsafe.Pointer {
	switch flag {
	case 8:
		return C.sub_4DFB50
	case 16:
		return C.nox_xxx_effectSpeedEngage_4DFC30
	case 1:
		return C.sub_4DFD10
	case 4:
		return C.nox_xxx_buff_4DFD80
	case 2:
		return C.nox_xxx_checkPoisonProtectEnch_4DFDE0
	case 32:
		return C.sub_4E0140
	default:
		return nil
	}
}

func modifierDisengagePointerNative4DFBB0(flag byte) unsafe.Pointer {
	switch flag {
	case 8:
		return C.sub_4DFB80
	case 16:
		return C.nox_xxx_effectSpeedDisengage_4DFCA0
	case 1:
		return C.nox_xxx_modifFireProtection_4DFD40
	case 4:
		return C.sub_4DFDB0
	case 2:
		return C.sub_4DFE10
	case 32:
		return C.sub_4E0170
	default:
		return nil
	}
}

func enchantItemTestInventoryNative4DFBB0(owner *server.Object, flag byte) bool {
	return C.nox_xxx_enchantItemTestInventory_4DFBB0(asObjectC(owner), C.uint8_t(flag)) != 0
}

//export nox_modifier_getColorRGB
func nox_modifier_getColorRGB(ptr unsafe.Pointer, index C.int) C.uint32_t {
	if ptr == nil || index < 0 || index >= 8 {
		return 0
	}
	cl := (*server.Modifier)(ptr).Colors12[int(index)]
	return C.uint32_t(uint32(cl.R) | uint32(cl.G)<<8 | uint32(cl.B)<<16)
}

//export nox_modifier_getEffectiveness
func nox_modifier_getEffectiveness(ptr unsafe.Pointer) C.int32_t {
	if ptr == nil {
		return 0
	}
	return C.int32_t((*server.Modifier)(ptr).Effectiveness36)
}

//export nox_modifier_getMaterial
func nox_modifier_getMaterial(ptr unsafe.Pointer) C.int32_t {
	if ptr == nil {
		return 0
	}
	return C.int32_t((*server.Modifier)(ptr).Material40)
}

//export nox_modifier_getPriEnchant
func nox_modifier_getPriEnchant(ptr unsafe.Pointer) C.int32_t {
	if ptr == nil {
		return 0
	}
	return C.int32_t((*server.Modifier)(ptr).PriEnchant44)
}

//export nox_modifier_getDurability
func nox_modifier_getDurability(ptr unsafe.Pointer) C.uint32_t {
	if ptr == nil {
		return 0
	}
	return C.uint32_t((*server.Modifier)(ptr).Durability52)
}

//export nox_modifier_getArmor
func nox_modifier_getArmor(ptr unsafe.Pointer) C.float {
	if ptr == nil {
		return 0
	}
	return C.float((*server.Modifier)(ptr).DamageCoeffOrArmor64)
}

//export nox_modifier_getRequiredStrength
func nox_modifier_getRequiredStrength(ptr unsafe.Pointer) C.uint16_t {
	if ptr == nil {
		return 0
	}
	return C.uint16_t((*server.Modifier)(ptr).ReqStrength60)
}

//export nox_modifier_getIndex
func nox_modifier_getIndex(ptr unsafe.Pointer) C.uint32_t {
	if ptr == nil {
		return 0
	}
	// Modifier and ModifierEff intentionally share their native name-pointer
	// and uint32 index prefix.
	return C.uint32_t((*server.Modifier)(ptr).TypeInd)
}

//export nox_modifier_effect_getDefendFunc
func nox_modifier_effect_getDefendFunc(ptr unsafe.Pointer) unsafe.Pointer {
	if ptr == nil {
		return nil
	}
	return (*server.ModifierEff)(ptr).Defend76.Fnc
}

//export nox_modifier_effect_getAttackFunc
func nox_modifier_effect_getAttackFunc(ptr unsafe.Pointer) unsafe.Pointer {
	if ptr == nil {
		return nil
	}
	return (*server.ModifierEff)(ptr).Attack40.Fnc
}

//export nox_modifier_effect_getPreHitFunc
func nox_modifier_effect_getPreHitFunc(ptr unsafe.Pointer) unsafe.Pointer {
	if ptr == nil {
		return nil
	}
	return (*server.ModifierEff)(ptr).AttackPreHit52.Fnc
}

//export nox_modifier_effect_getAttackInt
func nox_modifier_effect_getAttackInt(ptr unsafe.Pointer) C.int32_t {
	if ptr == nil {
		return 0
	}
	return C.int32_t((*server.ModifierEff)(ptr).Attack40.Val)
}

//export nox_modifier_effect_getAttackFloat
func nox_modifier_effect_getAttackFloat(ptr unsafe.Pointer) C.float {
	if ptr == nil {
		return 0
	}
	return C.float((*server.ModifierEff)(ptr).Attack40.Valf)
}

//export nox_modifier_effect_getPreHitInt
func nox_modifier_effect_getPreHitInt(ptr unsafe.Pointer) C.int32_t {
	if ptr == nil {
		return 0
	}
	return C.int32_t((*server.ModifierEff)(ptr).AttackPreHit52.Val)
}

//export nox_modifier_effect_getPreHitFloat
func nox_modifier_effect_getPreHitFloat(ptr unsafe.Pointer) C.float {
	if ptr == nil {
		return 0
	}
	return C.float((*server.ModifierEff)(ptr).AttackPreHit52.Valf)
}

//export nox_modifier_effect_getPreDamageInt
func nox_modifier_effect_getPreDamageInt(ptr unsafe.Pointer) C.int32_t {
	if ptr == nil {
		return 0
	}
	return C.int32_t((*server.ModifierEff)(ptr).AttackPreDmg64.Val)
}

//export nox_modifier_effect_getPreDamageFloat
func nox_modifier_effect_getPreDamageFloat(ptr unsafe.Pointer) C.float {
	if ptr == nil {
		return 0
	}
	return C.float((*server.ModifierEff)(ptr).AttackPreDmg64.Valf)
}

//export nox_modifier_effect_getDefendFloat
func nox_modifier_effect_getDefendFloat(ptr unsafe.Pointer) C.float {
	if ptr == nil {
		return 0
	}
	return C.float((*server.ModifierEff)(ptr).Defend76.Valf)
}

//export nox_modifier_effect_getEngageFloat
func nox_modifier_effect_getEngageFloat(ptr unsafe.Pointer) C.float {
	if ptr == nil {
		return 0
	}
	return C.float((*server.ModifierEff)(ptr).EngageFloat120)
}

//export nox_modifier_effect_getIdentDescription
func nox_modifier_effect_getIdentDescription(ptr unsafe.Pointer) *C.wchar2_t {
	if ptr == nil {
		return nil
	}
	return (*C.wchar2_t)(unsafe.Pointer((*server.ModifierEff)(ptr).IdentificationDescription16()))
}

//export nox_modifier_effect_getEngageFunc
func nox_modifier_effect_getEngageFunc(ptr unsafe.Pointer) unsafe.Pointer {
	if ptr == nil {
		return nil
	}
	return (*server.ModifierEff)(ptr).Engage112
}

//export nox_modifier_effect_getDisengageFunc
func nox_modifier_effect_getDisengageFunc(ptr unsafe.Pointer) unsafe.Pointer {
	if ptr == nil {
		return nil
	}
	return (*server.ModifierEff)(ptr).Disengage116
}

//export nox_modifier_getColorSlot
func nox_modifier_getColorSlot(ptr unsafe.Pointer, index C.int) C.int32_t {
	if ptr == nil || index < 0 || index >= 4 {
		return 0
	}
	return C.int32_t((*server.Modifier)(ptr).ColorIndexes()[int(index)])
}

//export nox_modifier_effect_getColorRGB
func nox_modifier_effect_getColorRGB(ptr unsafe.Pointer) C.uint32_t {
	if ptr == nil {
		return 0
	}
	cl := (*server.ModifierEff)(ptr).Color24
	return C.uint32_t(uint32(cl.R) | uint32(cl.G)<<8 | uint32(cl.B)<<16)
}

const modifierNativeSize = 88 + 6*(cgoABIPointerSize-4)

var _ = [1]struct{}{}[modifierNativeSize-unsafe.Sizeof(server.Modifier{})]

const modifierEffNativeSize = 144 + 16*(cgoABIPointerSize-4)

var _ = [1]struct{}{}[modifierEffNativeSize-unsafe.Sizeof(server.ModifierEff{})]

var (
	_ = nox_xxx_fireEffect_4E0550
	_ = nox_xxx_fireRingEffect_4E05B0
	_ = nox_xxx_blueFREffect_4E05F0
)

func init() {
	server.RegisterModifDamageEffect("DamageMultiplierEffect", C.nox_xxx_effectDamageMultiplier_4E04C0, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("StunEffect", C.nox_xxx_stunEffect_4E04D0, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("FireEffect", C.nox_xxx_fireEffect_4E0550, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("FireRingEffect", C.nox_xxx_fireRingEffect_4E05B0, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("BlueFireRingEffect", C.nox_xxx_blueFREffect_4E05F0, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("FrostEffect", C.nullsub_38, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("RecoilEffect", C.nox_xxx_recoilEffect_4E0640, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("ConfuseEffect", C.nox_xxx_confuseEffect_4E0670, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("LightningEffect", C.nox_xxx_lightngEffect_4E06F0, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("DrainManaEffect", C.nox_xxx_drainMEffect_4E0740, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("VampirismEffect", C.nox_xxx_vampirismEffect_4E07C0, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("PoisonEffect", C.nox_xxx_poisonEffect_4E0850, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("PanicEffect", C.nullsub_39, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("SympathyEffect", C.nox_xxx_sympathyEffect_4E08E0, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("ReadinessEffect", C.nullsub_22, server.ModEffectParseInt)
	server.RegisterModifDamageEffect("ProjectileSpeedEffect", C.nox_xxx_effectProjectileSpeed_4E09B0, server.ModEffectParseFloat)
	server.RegisterModifDamageEffect("ReplenishmentEffect", C.nullsub_36, server.ModEffectParseInt)

	server.RegisterModifDefendEffect("ArmorMultiplierEffect", C.sub_4E0370, server.ModEffectParseFloat)
	server.RegisterModifDefendEffect("DurabilityMultiplierEffect", C.sub_4E0380, server.ModEffectParseFloat)
	server.RegisterModifDefendEffect("ResilienceEffect", C.nullsub_40, server.ModEffectParseFloat)
	server.RegisterModifDefendEffect("InversionEffect", C.nox_xxx_inversionEffect_4E03D0, server.ModEffectParseInt)
	server.RegisterModifDefendEffect("GripEffect", C.nox_xxx_gripEffect_4E0480, server.ModEffectParseInt)
	server.RegisterModifDefendEffect("BreakingEffect", C.nullsub_41, server.ModEffectParseFloat)
	server.RegisterModifDefendEffect("PunctureProneEffect", C.nullsub_42, server.ModEffectParseFloat)

	server.RegisterModifUpdateEffect("RegenerationUpdate", C.nox_xxx_effectRegeneration_4E01D0, server.ModEffectParseInt)
	server.RegisterModifUpdateEffect("ParasiteUpdate", C.nullsub_43, server.ModEffectParseInt)
	server.RegisterModifUpdateEffect("AttractionUpdate", C.nullsub_44, server.ModEffectParseInt)
	server.RegisterModifUpdateEffect("ContinualReplenishmentUpdate", C.nox_xxx_attribContinualReplen_4E02C0, server.ModEffectParseInt)

	server.RegisterModifEngageEffect("BrillianceEngage", C.sub_4DFB50, server.ModEffectParseInt)
	server.RegisterModifEngageEffect("BrillianceDisengage", C.sub_4DFB80, server.ModEffectParseInt)
	server.RegisterModifEngageEffect("SpeedEngage", C.nox_xxx_effectSpeedEngage_4DFC30, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("SpeedDisengage", C.nox_xxx_effectSpeedDisengage_4DFCA0, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("FireProtectEngage", C.sub_4DFD10, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("FireProtectDisengage", C.nox_xxx_modifFireProtection_4DFD40, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("LightningProtectEngage", C.nox_xxx_buff_4DFD80, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("LightningProtectDisengage", C.sub_4DFDB0, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("PoisonProtectEngage", C.nox_xxx_checkPoisonProtectEnch_4DFDE0, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("PoisonProtectDisengage", C.sub_4DFE10, server.ModEffectParseFloat)
	server.RegisterModifEngageEffect("RegenerationEngage", C.sub_4E0140, nil)
	server.RegisterModifEngageEffect("RegenerationDisengage", C.sub_4E0170, nil)
}

//export nox_xxx_modifGetDescById_413330
func nox_xxx_modifGetDescById_413330(a1 int32) unsafe.Pointer {
	return GetServer().S().Modif.Nox_xxx_modifGetDescById413330(int(a1)).C()
}

//export nox_xxx_modifGetIdByName_413290
func nox_xxx_modifGetIdByName_413290(name *C.char) int32 {
	return int32(GetServer().S().Modif.Nox_xxx_modifGetIdByName413290(GoString(name)))
}

//export nox_xxx_getProjectileClassById_413250
func nox_xxx_getProjectileClassById_413250(a1 int32) unsafe.Pointer {
	return GetServer().S().Modif.Nox_xxx_getProjectileClassById413250(int(a1)).C()
}

//export nox_xxx_equipClothFindDefByTT_413270
func nox_xxx_equipClothFindDefByTT_413270(a1 int32) unsafe.Pointer {
	return GetServer().S().Modif.Nox_xxx_equipClothFindDefByTT413270(int(a1)).C()
}

//export sub_4A5E90_A
func sub_4A5E90_A() { Sub_4A5E90_A() }

//export nox_xxx_fireEffect_4E0550
func nox_xxx_fireEffect_4E0550(a1 unsafe.Pointer, a2p, a3p, a4p *nox_object_t, _ unsafe.Pointer) {
	GetServer().S().Nox_xxx_fireEffect_4E0550(a1, asObjectS(a2p), asObjectS(a3p), asObjectS(a4p))
}

//export nox_xxx_fireRingEffect_4E05B0
func nox_xxx_fireRingEffect_4E05B0(a1 unsafe.Pointer, a2p, a3p, a4p *nox_object_t, _ unsafe.Pointer) {
	Nox_xxx_fireRingEffect_4E05B0(a1, asObjectS(a2p), asObjectS(a3p), asObjectS(a4p))
}

//export nox_xxx_blueFREffect_4E05F0
func nox_xxx_blueFREffect_4E05F0(a1 unsafe.Pointer, a2p, a3p, a4p *nox_object_t, _ unsafe.Pointer) {
	Nox_xxx_blueFREffect_4E05F0(a1, asObjectS(a2p), asObjectS(a3p), asObjectS(a4p))
}
