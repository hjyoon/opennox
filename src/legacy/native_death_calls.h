#ifndef OPENNOX_NATIVE_DEATH_CALLS_H
#define OPENNOX_NATIVE_DEATH_CALLS_H

#include <stddef.h>
#include <stdio.h>
#include "GAME3_3.h"
#include "GAME4_2.h"
#include "map_push_units_52e040.h"
#include "random_reachable_point_4ed970.h"
#include "server__script__script.h"
#include "native_call_abi.h"

enum { NOX_DEATH_POSITION_OFFSET = offsetof(nox_object_t, x) };

// Carry the address as a scalar until the consuming native entry receives it.
// The point and script calls discard their pointer return registers.
NOX_NATIVE_CALL(void, nox_legacy_debris_point_at_address, sub_4ED970,
    (float radius, uintptr_t origin, float2* output),
    sub_4ED970(radius, (float2*)origin, output);)
NOX_NATIVE_CALL(void, nox_legacy_debris_force_at_address, nox_xxx_objectApplyForce_52DF80,
    (uintptr_t origin, nox_object_t* object, float force),
    nox_xxx_objectApplyForce_52DF80((float*)origin, object, force);)
NOX_NATIVE_CALL(int, nox_legacy_point_fx_at_address, nox_xxx_netSendPointFx_522FF0,
    (char id, uintptr_t origin),
    return nox_xxx_netSendPointFx_522FF0(id, (float2*)origin);)
NOX_NATIVE_CALL(void, nox_legacy_push_at_address, nox_xxx_mapPushUnitsAround_52E040,
    (uintptr_t origin, float outer, float inner, float force, nox_object_t* source, uintptr_t callback, uintptr_t arg),
    nox_xxx_mapPushUnitsAround_52E040((void*)origin, outer, inner, force, source, callback, arg);)
NOX_NATIVE_CALL(void, nox_legacy_damage_at_address, nox_xxx_mapDamageUnitsAround_4E25B0,
    (uintptr_t origin, float outer, float inner, int damage, int type, nox_object_t* source, nox_object_t* exclude),
    nox_xxx_mapDamageUnitsAround_4E25B0((float*)origin, outer, inner, damage, type, source, exclude);)
NOX_NATIVE_CALL(int, nox_legacy_spark_at_address, nox_xxx_netSparkExplosionFx_5231B0,
    (uintptr_t origin, char code),
    return nox_xxx_netSparkExplosionFx_5231B0((float*)origin, code);)
NOX_NATIVE_CALL(void, nox_legacy_death_script_at_address, nox_xxx_scriptCallByEventBlock_502490,
    (uintptr_t data, nox_object_t* caller, nox_object_t* trigger, int event),
    nox_xxx_scriptCallByEventBlock_502490((void*)data, caller, trigger, event);)
NOX_NATIVE_CALL(void, nox_legacy_generator_fx_at_address, nox_xxx_sendGeneratorBreakFX_523200,
    (uintptr_t origin, char intensity),
    nox_xxx_sendGeneratorBreakFX_523200((float*)origin, intensity);)

#endif
