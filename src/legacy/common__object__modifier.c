#include "memmap.h"
#include "GAME3_2.h"
#include "defs.h"
#include "client__gui__window.h"
#include "common__strman.h"

void nullsub_22() {}
void nullsub_36() {}
void nullsub_38() {}
void nullsub_39() {}
void nullsub_40() {}
void nullsub_41() {}
void nullsub_42() {}
void nullsub_43() {}
void nullsub_44() {}

//----- (00413480) --------------------------------------------------------
extern wchar2_t* nox_item_enchantment_tooltip_native_413480(uint8_t flag);
wchar2_t* sub_413480(char a1) {
	return nox_item_enchantment_tooltip_native_413480((uint8_t)a1);
}

// The original callbacks return the same address they receive. Keep that
// return in C so Go-originated scalar pointers can remain call-scoped.
void nox_modifierArmorMultiplierNative4E0370(void* effect, float* value);
void nox_modifierDurabilityMultiplierNative4E0380(void* effect, float* value);
void nox_modifierDamageMultiplierNative4E04C0(void* effect, float* damage);
void nox_modifierProjectileSpeedNative4E09B0(void* effect, nox_object_t* projectile);

float* sub_4E0370(void* effect, nox_object_t* item, uintptr_t context, nox_object_t* target, uintptr_t extra,
				  float* value) {
	nox_modifierArmorMultiplierNative4E0370(effect, value);
	return value;
}

float* sub_4E0380(void* effect, nox_object_t* item, uintptr_t context, nox_object_t* target, uintptr_t extra,
				  float* value) {
	nox_modifierDurabilityMultiplierNative4E0380(effect, value);
	return value;
}

float* nox_xxx_effectDamageMultiplier_4E04C0(
	void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, float* damage) {
	nox_modifierDamageMultiplierNative4E04C0(effect, damage);
	return damage;
}

nox_object_t* nox_xxx_effectProjectileSpeed_4E09B0(
	void* effect, nox_object_t* weapon, nox_object_t* owner, nox_object_t* target, nox_object_t* projectile) {
	nox_modifierProjectileSpeedNative4E09B0(effect, projectile);
	return projectile;
}
